package p2p

import (
	"encoding/binary"
	"net"
	"strings"
	"time"
)

func (p2p *P2PManager) headleMessage(peer *Peer, msgty uint16, msgbody []byte) {

	peer.activeTime = time.Now() // do live mark

	// ping pong
	if msgty == MsgTypePing {
		peer.SendMsg(MsgTypePong, nil)
		return
	}

	// discovery new node
	if msgty == MsgTypeDiscoverNewNodeJoin {
		msglen := 32 + 1 + 21 // id + ispublic + ip:port
		if len(msgbody) != msglen {
			return
		}
		newpeerid := msgbody[0:32]
		newpeerispublic := msgbody[32] == 1
		newpeeraddrstr := strings.Trim(string(msgbody[33:54]), string([]byte{0}))
		if !p2p.peerManager.IsCanAddToRelationshipPeerTable(newpeerid) {
			return
		}
		// connect to node
		newpeerTcpAddr, e := net.ResolveTCPAddr("tcp", newpeeraddrstr)
		if e != nil {
			return
		}
		if newpeerispublic {
			go p2p.TryConnectToNode(nil, newpeerTcpAddr)
			return
		}
		if peer.udpListenPort == 0 {
			return
		}
		newpeerUdpAddr, e := net.ResolveUDPAddr("udp4", newpeeraddrstr)
		if e != nil {
			return
		}
		socket, err := net.DialUDP("udp4", nil, newpeerUdpAddr)
		if err != nil || socket == nil {
			return
		}
		udplocaladdr := socket.LocalAddr()
		myConnTcpAddr, e := net.ResolveTCPAddr("tcp", udplocaladdr.String())
		if e != nil {
			return
		}
		// send data
		data := make([]byte, 2+32)
		binary.BigEndian.PutUint16(data[0:2], MsgTypeWantToConnectNode)
		copy(data[2:34], newpeerid)
		socket.Write(data) /// Send MsgTypeWantToConnectNode
		socket.Close()
		//try connect to new node
		go func() {
			<-time.Tick(time.Second * 10)
			p2p.TryConnectToNode(&net.TCPAddr{net.IPv4zero, myConnTcpAddr.Port, ""}, newpeerTcpAddr)
		}()
		return
	}

	// im public node
	if msgty == MsgTypeNotifyIsPublicNode {
		msglen := 21 // ip:port
		if len(msgbody) != msglen {
			return
		}
		newpeerAddrStr := strings.Trim(string(msgbody), string([]byte{0}))
		newpeerAddr, e := net.ResolveTCPAddr("tcp", newpeerAddrStr)
		if e != nil {
			return
		}
		if newpeerAddr.Port != p2p.config.TcpListenPort {
			return
		}
		p2p.selfPublicTCPListenAddr = newpeerAddr // set public ip and port
		return
	}

	// other want connect
	if msgty == MsgTypeOtherNodeWantToConnect {
		msglen := 21 // ip:port
		if len(msgbody) != msglen {
			return
		}
		newpeerAddrStr := strings.Trim(string(msgbody), string([]byte{0}))
		newpeerAddr, e := net.ResolveUDPAddr("udp", newpeerAddrStr)
		if e != nil {
			return
		}
		// UDP pass through out of NAT
		localaddr := &net.UDPAddr{net.IPv4zero, p2p.config.TcpListenPort, ""}
		socket, err := net.DialUDP("udp4", localaddr, newpeerAddr)
		if err != nil || socket == nil {
			return
		}
		// send data
		socket.Write([]byte("hello!"))
		socket.Close()
		return
	}

	// hand shake
	if msgty == MsgTypeHandShake {
		msglen := 2 + 2 + 2 + 32 + 32 // 70
		if len(msgbody) != msglen {
			return
		}
		peerVersion := binary.BigEndian.Uint16(msgbody[0:2])
		if peerVersion != P2PMustVersion {
			peer.Close()
			return
		}
		peer.tcpListenPort = int(binary.BigEndian.Uint16(msgbody[2:4]))
		peer.udpListenPort = int(binary.BigEndian.Uint16(msgbody[4:6]))
		peer.Id = msgbody[6:38]
		peer.Name = strings.Trim(string(msgbody[38:70]), string([]byte{0}))

		// add to manager
		p2p.peerManager.AddPeer(peer)

		return
	}

}
