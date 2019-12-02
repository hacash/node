package p2p

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"os"
)

func (p2p *P2PManager) startListenUDP() {

	udpconn, err := net.ListenUDP("udp", &net.UDPAddr{net.IPv4zero, p2p.config.UdpListenPort, ""})
	if err != nil {
		fmt.Println("startListenUDP error:", err)
		os.Exit(1)
	}

	for {
		data := make([]byte, 1025)
		rn, rmtaddr, err := udpconn.ReadFromUDP(data)
		//fmt.Println("udpconn.ReadFromUDP(data)", rmtaddr.String(), data[0:rn])
		if err != nil {
			continue
		}
		if rn > 1024 || rn < 2 {
			continue
		}
		msgtype := binary.BigEndian.Uint16(data[0:2])
		msgbody := data[2:rn]

		go p2p.handleUDPMessage(rmtaddr, msgtype, msgbody)

	}
}

func (p2p *P2PManager) handleUDPMessage(addr *net.UDPAddr, msgty uint16, msgbody []byte) {

	//fmt.Println("handleUDPMessage", msgty, len(msgbody), msgbody)

	if msgty == MsgTypeUDPAllowToConnectNode {
		msglen := 32 + 32
		if len(msgbody) != msglen {
			return
		}
		listenpeerid := msgbody[0:32]
		tarpeerid := msgbody[32:64]
		if bytes.Compare(p2p.selfPeerId, tarpeerid) == 0 {
			return // is my self
		}
		tarpeer, _ := p2p.peerManager.GetPeerByID(tarpeerid)
		if tarpeer == nil {
			return // not find
		}
		// send msg
		msgbody := make([]byte, 32+21)
		copy(msgbody[0:32], listenpeerid)
		copy(msgbody[32:53], addr.String())

		fmt.Println("MsgTypeUDPAllowToConnectNode", addr.String())

		tarpeer.SendMsg(MsgTypeAllowOtherNodeToConnect, msgbody)
		return
	}

	if msgty == MsgTypeUDPWantToConnectNode {
		msglen := 32 + 32
		if len(msgbody) != msglen {
			return
		}
		callpeerid := msgbody[0:32]
		tarpeerid := msgbody[32:64]
		tarpeer, _ := p2p.peerManager.GetPeerByID(tarpeerid)
		if tarpeer == nil {
			//fmt.Println("MsgTypeUDPWantToConnectNode   GetPeerByID(tarpeerid)  not find", hex.EncodeToString(tarpeerid))
			return
		}
		tarmsgbody := make([]byte, 32+21) // 255.255.255.255:65535
		copy(tarmsgbody[0:32], callpeerid)
		copy(tarmsgbody[32:53], addr.String())

		fmt.Println("MsgTypeUDPWantToConnectNode", addr.String())

		// send msg
		//fmt.Println("tarpeer.SendMsg(MsgTypeOtherNodeWantToConnect, tarmsgbody)")
		tarpeer.SendMsg(MsgTypeOtherNodeWantToConnect, tarmsgbody)
		return
	}

	if msgty == MsgTypeReportTCPListenPort {
		msglen := 32 + 2
		if len(msgbody) != msglen {
			return
		}
		tarpeerid := msgbody[0:32]
		tarlocaltcpport := int(binary.BigEndian.Uint16(msgbody[32:34]))

		tarpeer, _ := p2p.peerManager.GetPeerByID(tarpeerid)
		if tarpeer == nil {
			fmt.Println("not find peer id", hex.EncodeToString(tarpeerid))
			return
		}
		// check node ip is public ?
		if tarpeer.tcpListenPort == tarlocaltcpport {
			tarpeer.PublicListenTcpAddr = &net.TCPAddr{addr.IP, addr.Port, ""}
			msgbody := make([]byte, 21)
			copy(msgbody, tarpeer.PublicListenTcpAddr.String())
			tarpeer.SendMsg(MsgTypeNotifyIsPublicNode, msgbody) // Notify node
		}
		//

		return
	}

}

func (p2p *P2PManager) reportTCPListen(target_udp_addr *net.UDPAddr) {

	// UDP call to out of NAT
	socket, err := net.DialUDP("udp4",
		&net.UDPAddr{net.IPv4zero, p2p.config.TcpListenPort, ""},
		target_udp_addr,
	)
	//laddr := net.UDPAddr{net.IPv4zero, p2p.config.TcpListenPort, ""}
	//socket, err := reuseport.Dial("udp", laddr.String(), target_udp_addr.String())
	if err != nil {
		fmt.Println("reportTCPListen DialUDP error", err)
		//os.Exit(1)
		return
	}
	// send data
	data := make([]byte, 2+32+2)
	binary.BigEndian.PutUint16(data[0:2], MsgTypeReportTCPListenPort)
	copy(data[2:34], p2p.selfPeerId)
	binary.BigEndian.PutUint16(data[34:36], uint16(p2p.config.TcpListenPort))

	socket.Write(data) /// Send MsgTypeReportTCPListenPort
	//n, err := socket.Write([]byte("hello!"))
	//fmt.Println(n, err)
	//socket.Close()

	fmt.Println("reportTCPListen", target_udp_addr.String())
}

func (p2p *P2PManager) natPassOutTcpAddr(localaddr *net.TCPAddr, allowaddr *net.TCPAddr) error {

	localudpaddr, err := net.ResolveUDPAddr("udp", localaddr.String())
	if err != nil {
		return err
	}
	allowudpaddr, err := net.ResolveUDPAddr("udp", allowaddr.String())
	if err != nil {
		return err
	}
	// UDP call to out of NAT
	return p2p.natPassOut(localudpaddr, allowudpaddr)
}

func (p2p *P2PManager) natPassOut(localaddr *net.UDPAddr, allowaddr *net.UDPAddr) error {

	localaddr.IP = net.IPv4zero
	fmt.Println("natPassOut", localaddr.String(), "=>", allowaddr.String())

	// UDP call to out of NAT
	socket, err := net.DialUDP("udp4", localaddr, allowaddr)
	if err != nil {
		fmt.Println("natPassOut error", err)
		//os.Exit(1)
		return err
	}
	_, err = socket.Write([]byte("nat_pass"))
	err = socket.Close()
	if err != nil {
		return err
	}
	return nil
}
