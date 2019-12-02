package p2p

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
		msglen := 32 // id
		if len(msgbody) != msglen {
			return
		}
		newpeerid := msgbody[0:32]
		if bytes.Compare(newpeerid, p2p.selfPeerId) == 0 {
			return // is my self
		}
		p2p.peerManager.AddKnownPeerId(newpeerid)
		if _, ldok := p2p.peerManager.waitToConnectNode.Load(string(newpeerid)); ldok {
			return // has wait to connect
		}
		if !p2p.peerManager.IsCanAddToRelationshipPeerTable(newpeerid) {
			return // not relation ship
		}
		servernodeUdpAddr, e := net.ResolveUDPAddr("udp", peer.TcpConn.RemoteAddr().String())
		if e != nil {
			return
		}
		servernodeUdpAddr.Port = peer.udpListenPort
		//fmt.Println("servernodeUdpAddr.Port = peer.udpListenPort", servernodeUdpAddr.String())
		socket, err := net.DialUDP("udp4", nil, servernodeUdpAddr)
		if err != nil || socket == nil {
			return
		}
		udplocaladdr := socket.LocalAddr()
		// send data
		data := make([]byte, 2+32+32)
		binary.BigEndian.PutUint16(data[0:2], MsgTypeUDPWantToConnectNode)
		copy(data[2:34], p2p.selfPeerId)
		copy(data[34:66], newpeerid)
		socket.Write(data) /// Send MsgTypeUDPWantToConnectNode
		socket.Close()
		// save addr
		p2p.peerManager.waitToConnectNode.Store(string(newpeerid), udplocaladdr)
		go func() {
			<-time.Tick(time.Second * 77)
			// check connect to node to delete
			p2p.peerManager.waitToConnectNode.Delete(string(newpeerid))
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

	if msgty == MsgTypeAllowOtherNodeToConnect {

		fmt.Println("MsgTypeAllowOtherNodeToConnect", len(msgbody), msgbody)

		msglen := 32 + 21 // public ip:port
		if len(msgbody) != msglen {
			return
		}
		newpeerId := msgbody[0:32]
		waitnode, ldok := p2p.peerManager.waitToConnectNode.Load(string(newpeerId))
		if !ldok {
			return // not find or time out
		}
		localaddr := waitnode.(net.Addr)
		localtcpaddr, _ := net.ResolveTCPAddr("tcp", localaddr.String())
		newpeerAddrStr := strings.Trim(string(msgbody[32:53]), string([]byte{0}))
		newpeerAddr, e := net.ResolveTCPAddr("tcp", newpeerAddrStr)
		if e != nil {
			return
		}
		p2p.peerManager.AddKnownPeerId(newpeerId)
		// start tcp connect
		go func() {
			// UDP call to out of NAT
			err := p2p.natPassOutTcpAddr(localtcpaddr, newpeerAddr)
			if err != nil {
				return
			}
			<-time.Tick(time.Second * 3)
			go p2p.TryConnectToNode(localtcpaddr, newpeerAddr)
			// clear data
			p2p.peerManager.waitToConnectNode.Delete(string(newpeerId))
		}()
	}

	// other want connect
	if msgty == MsgTypeOtherNodeWantToConnect {

		fmt.Println("MsgTypeOtherNodeWantToConnect", len(msgbody), string(msgbody))

		msglen := 32 + 21 // public ip:port
		if len(msgbody) != msglen {
			return
		}
		newpeerid := msgbody[0:32]
		newpeerAddrStr := strings.Trim(string(msgbody[32:53]), string([]byte{0}))
		newpeerAddr, e := net.ResolveUDPAddr("udp", newpeerAddrStr)
		if e != nil {
			return
		}
		p2p.peerManager.AddKnownPeerId(newpeerid)
		// notify server node
		server_udp_addr, _ := net.ResolveUDPAddr("udp", peer.TcpConn.RemoteAddr().String())
		server_udp_addr.Port = peer.udpListenPort
		socket, err := net.DialUDP("udp4", nil, server_udp_addr)
		if err != nil || socket == nil {
			return
		}
		local_addr := socket.LocalAddr()
		// send msg
		data := make([]byte, 2+32+32)
		binary.BigEndian.PutUint16(data[0:2], MsgTypeUDPAllowToConnectNode)
		copy(data[2:34], p2p.selfPeerId)
		copy(data[34:66], newpeerid)
		socket.Write(data)
		socket.Close()
		// start tcp listen
		fmt.Println("allowConnectNodeListenTCP ", local_addr.(*net.UDPAddr).String(), "NAT pass", newpeerAddr.String())
		go p2p.allowConnectNodeListenTCP(local_addr.(*net.UDPAddr), newpeerAddr)
		return
	}

	// hand shake
	if msgty == MsgTypeHandShake {
		//fmt.Println("MsgTypeHandShake", msgbody)
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
		addok, _ := p2p.peerManager.AddPeer(peer)
		if addok {
			// find new node msg
			peer.AddKnownPeerId(peer.Id)
			p2p.peerManager.SendFindNewNodeMsgToUnawarePeers(peer)

			//addr, e := net.ResolveUDPAddr("udp", peer.TcpConn.RemoteAddr().String())
			//if e != nil {
			//	return
			//}
			////fmt.Println(peer.udpListenPort, msgbody)
			//addr.Port = peer.udpListenPort
			//go func() {
			//	<- time.Tick(time.Second * 3)
			//	p2p.reportTCPListen(addr)
			//}()
		}

		return
	}

}
