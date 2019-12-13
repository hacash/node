package p2p

import (
	"net"
)

func (p2p *P2PManager) handleUDPMessage(addr *net.UDPAddr, msgty uint16, msgbody []byte) {

	//fmt.Println("handleUDPMessage", msgty, len(msgbody), msgbody)

	if UDPMsgTypeEnquirePublic == msgty {
		//fmt.Println("UDPMsgTypeEnquirePublic == msgty")
		if len(msgbody) != 16+4 {
			return
		}
		if addr.IP.IsLoopback() {
			return // local ip
		}
		callpeer := p2p.GetPeerByID(msgbody[0:16])
		if callpeer == nil {
			return
		}
		// send msg
		//fmt.Println("UDPMsgTypeEnquirePublic", msgty, len(msgbody), msgbody)
		callpeer.SendMsg(TCPMsgTypeReplyPublic, msgbody[16:20])
		return
	}

}
