package p2p

import (
	"encoding/binary"
	"math/rand"
	"net"
	"strings"
	"time"
)

func (p2p *P2PManager) readyHandshakeAndSuccessfulToRequiredata(peer *Peer) {
	if peer.isPermitCompleteNode == false || peer.ID == nil {
		return
	}
	// ready to do
	if peer.publicIPv4 != nil {
		// is public and add peer
		p2p.AddPeerToTargetGroup(p2p.peerManager.publicPeerGroup, peer)
		// get addrs
		peer.SendMsg(TCPMsgTypeGetPublicConnectedPeerAddrs, nil)
	} else {
		// ask is public
		p2p.sendMsgToEnquirePublic(peer)
	}
}

func (p2p *P2PManager) handleTCPMessage(peer *Peer, msgty uint16, msgbody []byte) {

	peer.activeTime = time.Now() // do live mark

	if TCPMsgTypePublicConnectedPeerAddrs == msgty {
		// got addrs
		if len(msgbody)%6 != 0 {
			return // data error
		}
		for i := 0; i < len(msgbody)/6; i++ {
			k := i * 6
			one := msgbody[k : k+6]
			//fmt.Println("AddOldPublicPeerAddr", one)
			p2p.AddOldPublicPeerAddrByBytes(one)
		}
		return
	}

	if TCPMsgTypeGetPublicConnectedPeerAddrs == msgty {
		addrsbytes := p2p.peerManager.GetAllCurrentConnectedPublicPeerAddressBytes()
		// send addrs
		peer.SendMsg(TCPMsgTypePublicConnectedPeerAddrs, addrsbytes)
		return
	}

	if TCPMsgTypeConnectRefuse == msgty {
		peer.Close()
		return
	}

	if TCPMsgTypeHandshakeSuccess == msgty {
		peer.isPermitCompleteNode = true
		// connect is finish successfully
		p2p.readyHandshakeAndSuccessfulToRequiredata(peer)
		return
	}

	// hand shake
	if TCPMsgTypeHandshake == msgty {
		//fmt.Println("MsgTypeHandShake", msgbody)
		msglen := 2 + 2 + 2 + 16 + 32 // 54
		if len(msgbody) != msglen {
			peer.Close()
			return
		}
		peerVersion := binary.BigEndian.Uint16(msgbody[0:2])
		if peerVersion != P2PMustVersion {
			peer.Close()
			return
		}
		peer.tcpListenPort = int(binary.BigEndian.Uint16(msgbody[2:4]))
		peer.udpListenPort = int(binary.BigEndian.Uint16(msgbody[4:6]))
		peerID := msgbody[6:22]
		// check have
		havp := p2p.GetPeerByID(peerID)
		if havp != nil {
			// already have id // 断开以前的连接，使用最新的连接
			p2p.peerManager.DropPeer(peer)
			havp.Close()
			/*peer.SendMsg(TCPMsgTypeConnectRefuse, nil)
			peer.Close()
			return*/
		}
		peer.ID = peerID
		peer.AddKnownPeerId(peer.ID)
		peer.Name = strings.Trim(string(msgbody[22:54]), string([]byte{0}))
		// add to lookup
		p2p.lookupPeers.Add(peer)
		// send handshake success msg
		peer.SendMsg(TCPMsgTypeHandshakeSuccess, nil)
		// connect is finish successfully
		p2p.readyHandshakeAndSuccessfulToRequiredata(peer)
		return
	}

	////////////////////////////////////////////////////////

	// check is hand shake
	if peer.ID == nil {
		return
	}

	if TCPMsgTypeDiscoverPublicPeerJoin == msgty {
		if len(msgbody) != 16+6 {
			return
		}
		// broadcast again
		peerID := msgbody[0:16]
		ipportbts := msgbody[16:22]
		//fmt.Println("TCPMsgTypeDiscoverPublicPeerJoin", peerID, ipportbts)
		peer.AddKnownPeerId(peerID)
		p2p.peerManager.BroadcastFindNewNodeMsgToUnawarePublicPeersByBytes(peerID, ipportbts)
		//fmt.Println("TCPMsgTypeDiscoverPublicPeerJoin", len(msgbody), msgbody)
		if p2p.peerManager.publicPeerGroup.IsCanAddToRelationshipPeerTable(peerID) {
			tcpaddr := ParseIPPortToTCPAddrByByte(ipportbts)
			//fmt.Println("ParseIPPortToTCPAddrByByte", ipportbts, tcpaddr.String())
			go func() {
				connerr := p2p.TryConnectToPeer(nil, tcpaddr)
				if connerr == nil { // reput in
					p2p.AddOldPublicPeerAddrByBytes(ipportbts)
				}
			}()
		}
		return
	}

	if TCPMsgTypeTellPublicIP == msgty {
		if len(msgbody) != 4 {
			return
		}
		if p2p.selfRemotePublicIP == nil {
			p2p.selfRemotePublicIP = msgbody[0:4]
		}
		return
	}

	if TCPMsgTypeReplyPublic == msgty {
		if len(msgbody) != 4 {
			return
		}

		p2p.changeStatusLock.Lock()
		defer p2p.changeStatusLock.Unlock()

		checkcode := binary.BigEndian.Uint32(msgbody[0:4])

		//fmt.Println("checkcode", checkcode)

		if res, ldok := p2p.waitToReplyIsPublicPeer[peer]; ldok {
			delete(p2p.waitToReplyIsPublicPeer, peer)
			tcp, e := net.ResolveTCPAddr("tcp", peer.TcpConn.RemoteAddr().String())
			if e != nil {
				return
			}
			if res.code == checkcode {
				peer.publicIPv4 = tcp.IP.To4() // ok is public
				//fmt.Println("res checkcode",  res.code, len(tcp.IP), []byte(tcp.IP), peer.publicIPv4, tcp.IP.String())
				// send public ip
				peer.SendMsg(TCPMsgTypeTellPublicIP, peer.publicIPv4)
				// add peer
				p2p.AddPeerToTargetGroup(p2p.peerManager.publicPeerGroup, peer)
				// broadcast msg find new public peer
				p2p.peerManager.BroadcastFindNewNodeMsgToUnawarePublicPeers(peer)
			}
		} else {
			//fmt.Println("not find in waitToReplyIsPublicPeer")
		}
		return
	}

	////////////////////////////////////////////////////////

	// ping pong
	if TCPMsgTypePing == msgty {
		peer.SendMsg(TCPMsgTypePong, nil)
		return
	}

	// msg handle
	if TCPMsgTypeData == msgty {
		if p2p.customerDataHandler == nil {
			return
		}
		if len(msgbody) < 2 {
			return
		}
		// call handle func
		go p2p.customerDataHandler.OnMsgData(p2p.peerManager, peer, binary.BigEndian.Uint16(msgbody[0:2]), msgbody[2:])
		return
	}

}

func (p2p *P2PManager) sendMsgToEnquirePublic(peer *Peer) {

	p2p.changeStatusLock.Lock()
	defer p2p.changeStatusLock.Unlock()

	if peer.TcpConn == nil {
		return
	}

	checkcode := rand.Uint32()
	p2p.waitToReplyIsPublicPeer[peer] = struct {
		curt time.Time
		code uint32
	}{curt: time.Now(), code: checkcode}
	// send udp msg
	udplistenaddr, _ := net.ResolveUDPAddr("udp", peer.TcpConn.RemoteAddr().String())
	udplistenaddr.Port = peer.udpListenPort
	udpconn, e := net.DialUDP("udp", nil, udplistenaddr)
	if e != nil {
		return
	}
	data := make([]byte, 2+16+4)
	binary.BigEndian.PutUint16(data[0:2], UDPMsgTypeEnquirePublic)
	copy(data[2:18], p2p.myselfpeer.ID)
	binary.BigEndian.PutUint32(data[18:22], checkcode)
	udpconn.Write(data)
	udpconn.Close()
}
