package p2p

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
)

func (p2p *P2PManager) allowConnectNodeListenTCP(localaddr *net.UDPAddr, allowaddr *net.UDPAddr) {

	var localtcpaddr *net.TCPAddr

	// UDP call to out of NAT
	socket, err := net.DialUDP("udp4", localaddr, allowaddr)
	if err != nil {
		fmt.Println("allowConnectNodeListenTCP error", err)
		//os.Exit(1)
		return
	}
	localtcpaddr, err = net.ResolveTCPAddr("tcp", socket.LocalAddr().String())
	if err != nil {
		return
	}
	socket.Write([]byte("nat_pass"))
	socket.Close()

	// start listen
	listener, err := net.ListenTCP("tcp", localtcpaddr)
	if err != nil {
		fmt.Println("allowConnectNodeListenTCP error:", err)
		return
	}

	var gotconn net.Conn = nil

	go func() {
		<-time.Tick(time.Second * 99)
		if gotconn == nil {
			listener.Close()
			// time out to close listener
		}
	}()

	//fmt.Println("allowConnectNodeListenTCP  ok!!!")

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}
	gotconn = conn
	p2p.handleNewConn(conn)
	// close listener
	listener.Close()
	return
}

func (p2p *P2PManager) startListenTCP() {

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{net.IPv4zero, p2p.config.TcpListenPort, ""})
	//laddr := net.TCPAddr{net.IPv4zero, p2p.config.TcpListenPort, ""}
	//listener, err := reuseport.Listen("tcp", laddr.String())
	if err != nil {
		fmt.Println("startListenTCP error:", err)
		os.Exit(1)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			break
		}
		go p2p.handleNewConn(conn)
	}

}

func (p2p *P2PManager) handleNewConn(conn net.Conn) {

	peer := NewPeer(nil, "")

	fmt.Println("handleNewConn LocalAddr", conn.LocalAddr().String(), "RemoteAddr", conn.RemoteAddr().String())

	peer.TcpConn = conn
	// time out for sign up
	go func() {
		<-time.Tick(time.Second * 35)
		if len(peer.Id) == 0 {
			fmt.Println("peer.Close() <- time.Tick(time.Second * 35)")
			peer.Close() // drop peer and close connect at time out
		}
	}()

	//RemoteAddr := conn.RemoteAddr()
	//fmt.Println("Connect Remote Addr", RemoteAddr)

	// handshake
	go p2p.sendHandShakeMessageToConn(conn)

	// read msg
	segdata := make([]byte, 4069)
	holdbuf := bytes.NewBuffer([]byte{})
	notreadwait := false
	for {
		var readnum int = 0
		if !notreadwait {
			rn, e1 := conn.Read(segdata)
			if e1 != nil {
				fmt.Println(e1)
				break
			}
			if rn <= 0 {
				continue
			}
			readnum = rn
		}
		notreadwait = false

		//fmt.Println("conn.Read(segdata)", segdata[:readnum])
		holdbuf.Write(segdata[:readnum])
		holdsize := holdbuf.Len()
		if holdsize < 4 {
			continue
		}
		data := holdbuf.Bytes()
		msgType := binary.BigEndian.Uint16(data[:2])
		msgLen := int(0)
		msgLenRealSegSize := int(0)
		msgBody := []byte{}
		if msgType == MsgTypeData {
			if uint32(holdsize) < 6 {
				continue
			}
			msgLen = int(binary.BigEndian.Uint32(data[2:6]))
			msgLenRealSegSize = holdsize - 6
			msgBody = data[6:]
		} else {
			msgLen = int(binary.BigEndian.Uint16(data[2:4]))
			msgLenRealSegSize = holdsize - 4
			msgBody = data[4:]
		}
		if msgLenRealSegSize < msgLen {
			continue // next segdata
		} else if msgLenRealSegSize > msgLen {
			holdbuf = bytes.NewBuffer(msgBody[msgLen:]) // cache
			msgBody = msgBody[:msgLen]
			notreadwait = true
		}
		//fmt.Println("p2p.headleMessage", msgType, msgBody)
		// deal real msg
		go p2p.headleMessage(peer, msgType, msgBody)
		// reset
		if !notreadwait {
			holdbuf.Truncate(0) // clear
		}

	}

	// msg error and drop peer
	if peer.Id != nil {
		fmt.Println("handleNewConn DropPeer", peer.Name)
		p2p.peerManager.DropPeer(peer)
	}

}

func (p2p *P2PManager) sendHandShakeMessageToConn(conn net.Conn) {
	tmp_p := NewPeer(nil, "")
	tmp_p.TcpConn = conn

	hsml := 2 + 2 + 2 + 32 + 32 // 70
	data := make([]byte, hsml)

	binary.BigEndian.PutUint16(data[0:2], P2PMustVersion)
	binary.BigEndian.PutUint16(data[2:4], uint16(p2p.config.TcpListenPort))
	binary.BigEndian.PutUint16(data[4:6], uint16(p2p.config.UdpListenPort))
	copy(data[6:38], p2p.selfPeerId)
	copy(data[38:70], p2p.selfPeerName)

	tmp_p.SendMsg(MsgTypeHandShake, data)
}
