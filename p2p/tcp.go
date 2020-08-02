package p2p

import (
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
		if p2p.lookupPeers.Cardinality() >= p2p.config.lookupConnectMaxLen {
			peer.SendMsg(TCPMsgTypeConnectRefuse, nil)
			peer.Close()
			return
		}

	}

	// handshake
	p2p.sendHandShakeMessageToConn(conn)

	// handle msg
	go func() {
		for {
			msgdataitem := <-peer.tcpMsgDataCh
			if msgdataitem.ty == 0 {
				break // end
			}
			p2p.handleTCPMessage(peer, msgdataitem.ty, msgdataitem.v)
		}
	}()

	// read msg
	segdata := make([]byte, 4069)
	fullmsgdata := make([]byte, 0, 6)
	for {
	STARTDATACHECK:

		if len(fullmsgdata) >= 4 {

			//fmt.Println("fullmsgdata", fullmsgdata)

			msgty := binary.BigEndian.Uint16(fullmsgdata[0:2])
			length := 0
			value := make([]byte, 0)
			if msgty == TCPMsgTypeData {
				if len(fullmsgdata) >= 6 {
					length = int(binary.BigEndian.Uint32(fullmsgdata[2:6]))
					value = fullmsgdata[6:]
				} else {
					goto NEXTREAD
				}
			} else {
				length = int(binary.BigEndian.Uint16(fullmsgdata[2:4]))
				value = fullmsgdata[4:]
			}

			//fmt.Println("msgty", msgty, "length", length, "value", len(value), value)

			var segvalue []byte = nil
			if len(value) < length {
				goto NEXTREAD
			} else if len(value) > length {
				//fmt.Println("len(value) > length")
				segvalue = value[0:length]
				fullmsgdata = value[length:]
			} else {
				segvalue = value
				fullmsgdata = make([]byte, 0, 6)
			}

			//fmt.Println("msgty", msgty, "length", length, "value", len(segvalue), segvalue)
			//fmt.Println("fullmsgdata", fullmsgdata)

			peer.tcpMsgDataCh <- struct {
				ty uint16
				v  []byte
			}{ty: msgty, v: segvalue}

			if len(fullmsgdata) >= 4 {
				goto STARTDATACHECK
			}

		}
	NEXTREAD:
		// read
		rn, e1 := conn.Read(segdata)
		if e1 != nil {
			//fmt.Println(e1)
			break
		}
		fullmsgdata = append(fullmsgdata, segdata[0:rn]...)
	}

	// msg end or error and drop peer
	if peer.ID != nil {
		if peer.publicIPv4 != nil {
			if peer.isPermitCompleteNode {
				p2p.AddOldPublicPeerAddr(peer.publicIPv4, peer.tcpListenPort) // add addr back
			}
			addr := ""
			if peer.TcpConn != nil {
				addr = peer.TcpConn.RemoteAddr().String()
			}
			fmt.Println("[Peer] Disconnected @public peer name:", peer.Name, "addr:", addr)
		} else {
			fmt.Println("[Peer] Disconnected peer:", peer.Name)
		}
		if p2p.customerDataHandler != nil {
			p2p.customerDataHandler.OnDisconnected(peer) // disconnect event call
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
	copy(data[6:22], p2p.myselfpeer.ID)
	copy(data[22:54], p2p.myselfpeer.Name)

	tmp_p.SendMsg(TCPMsgTypeHandshake, data)
}
