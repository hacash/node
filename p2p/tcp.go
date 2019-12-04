package p2p

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
)

func (p2p *P2PManager) startListenTCP() {

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{net.IPv4zero, p2p.config.TCPListenPort, ""})
	//laddr := net.TCPAddr{net.IPv4zero, p2p_other.config.TCPListenPort, ""}
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
		go p2p.handleNewConn(conn, false)
	}

}

func (p2p *P2PManager) handleNewConn(conn net.Conn, isConnectToPublic bool) {

	peer := NewPeer(nil, "")

	//fmt.Println("handleNewConn LocalAddr", conn.LocalAddr().String(), "RemoteAddr", conn.RemoteAddr().String())

	peer.TcpConn = conn
	peer.connTime = time.Now()

	// time out for sign up
	//go func() {
	//	<-time.Tick(time.Second * 35)
	//	if len(peer.Id) == 0 {
	//		fmt.Println("peer.Close() <- time.Tick(time.Second * 35)")
	//		peer.Close() // drop peer and close connect at time out
	//	}
	//}()

	//RemoteAddr := conn.RemoteAddr()
	//fmt.Println("Connect Remote Addr", RemoteAddr)

	if isConnectToPublic {
		// record public ip
		tcp, e := net.ResolveTCPAddr("tcp", peer.TcpConn.RemoteAddr().String())
		if e == nil {
			peer.publicIPv4 = tcp.IP.To4() // ok is public
		}
		//
	} else {
		if p2p.lookupPeers.Cardinality() >= p2p.config.LookupConnectMaxLen {
			peer.SendMsg(TCPMsgTypeConnectRefuse, nil)
			peer.Close()
		}

	}

	// handshake and get addrs
	go func() {
		p2p.sendHandShakeMessageToConn(conn)
		<-time.Tick(time.Second * 1)
		if isConnectToPublic {
			// get addr
			peer.SendMsg(TCPMsgTypeGetPublicConnectedPeerAddrs, nil)
		}
	}()

	// read msg
	segdata := make([]byte, 4069)
	holdbuf := bytes.NewBuffer([]byte{})
	notreadwait := false
	for {
		var readnum int = 0
		if !notreadwait {
			rn, e1 := conn.Read(segdata)
			if e1 != nil {
				//fmt.Println(e1)
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
		if msgType == TCPMsgTypeData {
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
		//fmt.Println("p2p_other.headleMessage", msgType, msgBody)
		// deal real msg
		go p2p.headleMessage(peer, msgType, msgBody)
		// reset
		if !notreadwait {
			holdbuf.Truncate(0) // clear
		}

	}

	// msg error and drop peer
	if len(peer.Id) > 0 {
		if peer.publicIPv4 != nil && peer.IsPermitCompleteNode {
			p2p.AddOldPublicPeerAddr(peer.publicIPv4, peer.tcpListenPort)
			fmt.Println("[Peer] disconnected @public peer name:", peer.Name, "addr:", peer.TcpConn.RemoteAddr().String())
		} else {
			fmt.Println("[Peer] disconnected peer:", peer.Name)
		}
		if p2p.customerDataHandler != nil && peer.IsPermitCompleteNode {
			go p2p.customerDataHandler.OnDisconnected(peer) // disconnect event call
		}
		//fmt.Println("handleNewConn DropPeer", peer.Name)
		p2p.peerManager.DropPeer(peer)
		p2p.lookupPeers.Remove(peer)
		peer.Close()
	}

}

func (p2p *P2PManager) sendHandShakeMessageToConn(conn net.Conn) {
	tmp_p := NewPeer(nil, "")
	tmp_p.TcpConn = conn

	hsml := 2 + 2 + 2 + 16 + 32 // len(70) = version + tcpport + udpport + id + name
	data := make([]byte, hsml)

	binary.BigEndian.PutUint16(data[0:2], P2PMustVersion)
	binary.BigEndian.PutUint16(data[2:4], uint16(p2p.config.TCPListenPort))
	binary.BigEndian.PutUint16(data[4:6], uint16(p2p.config.UDPListenPort))
	copy(data[6:22], p2p.selfPeerId)
	copy(data[22:54], p2p.selfPeerName)

	tmp_p.SendMsg(TCPMsgTypeHandShake, data)
}
