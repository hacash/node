package p2p

import (
	"encoding/binary"
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
		if err != nil {
			continue
		}
		if rn > 1024 || rn < 2 {
			continue
		}
		msgtype := binary.BigEndian.Uint16(data[:2])
		msgbody := data[2:]

		go p2p.handleUDPMessage(rmtaddr, msgtype, msgbody)

	}
}

func (p2p *P2PManager) handleUDPMessage(addr *net.UDPAddr, msgty uint16, msgbody []byte) {

	if msgty == MsgTypeReportTCPListenPort {
		msglen := 32 + 2
		if len(msgbody) != msglen {
			return
		}
		tarpeerid := msgbody[0:32]
		tarlocaltcpport := int(binary.BigEndian.Uint16(msgbody[32:34]))

		tarpeer, _ := p2p.peerManager.GetPeerByID(tarpeerid)
		if tarpeer == nil {
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

	if msgty == MsgTypeWantToConnectNode {

		msglen := 32
		if len(msgbody) != msglen {
			return
		}
		tarpeerid := msgbody[0:32]
		tarpeer, _ := p2p.peerManager.GetPeerByID(tarpeerid)
		if tarpeer == nil {
			return
		}
		tarmsgbody := make([]byte, 21) // 255.255.255.255:65535
		copy(tarmsgbody, addr.String())
		// send msg
		tarpeer.SendMsg(MsgTypeOtherNodeWantToConnect, tarmsgbody)
		return
	}

}
