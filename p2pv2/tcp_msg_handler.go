package p2pv2

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"strings"
	"time"
)

func (p *P2P) handleConnMsg(connid uint64, conn net.Conn, peer *Peer, msg []byte) {

	// fmt.Println("- - - - p2p.handleConnMsg", conn.RemoteAddr().String(), ":", msg)

	ct := time.Now()
	p.PeerChangeMux.Lock()
	peer.activeTime = &ct // Active time
	p.PeerChangeMux.Unlock()

	// Processing messages
	msgty := msg[0]
	msgbody := msg[1:]

	switch msgty {

	case P2PMsgTypeRemindMeIsPublicPeer:
		// I was informed that I was a public network node
		p.MyselfIsPublicPeer = true
		break

	case P2PMsgTypeCustomer:
		// Client message arrival
		if len(msgbody) < 2 {
			break
		}
		if p.msgHandler != nil {
			mty := binary.BigEndian.Uint16(msgbody[0:2])
			mbody := msgbody[2:]
			//fmt.Println("P2PMsgTypeCustomer", mty, mbody)
			p.msgHandler.OnMsgData(p, peer, mty, mbody)
		}
		break

	case P2PMsgTypePing:
		// ping pong
		go sendTcpMsg(conn, P2PMsgTypePong, nil)
		break

	case P2PMsgTypePong:
		// Receiving a Pong message, do nothing
		break

	case P2PMsgTypeRequestNearestPublicNodes:
		// Send all connected public network nodes
		nodes := make([]*fdNodes, 0)
		for _, pid := range p.BackboneNodeTable {
			if peer := p.getPeerByID(pid); peer != nil && peer.PublicIpPort != nil {
				nodes = append(nodes, &fdNodes{
					TcpAddr: peer.PublicIpPort,
					ID:      peer.ID,
				})
				if len(nodes) >= 100 {
					break
				}
			}
		}
		buf := bytes.NewBuffer([]byte{uint8(len(nodes))}) // 第一位为数量 <= 200
		buf.Write(serializeFindNodesToBytes(nodes))
		conn.Write(buf.Bytes()) // Send all public network nodes
		conn.Close()            // Close connection now
		break

	case P2PMsgTypeRequestIDForPublicNodeCheck:
		conn.Write(p.peerSelf.ID) // Send my ID
		conn.Close()              // Close connection now
		break

	case P2PMsgTypeAnswerIdKeepConnectAsPeer:
		// The other party is also willing to connect permanently
		if len(msgbody) != PeerIDSize+PeerNameSize {
			break
		}
		if peer.ID != nil {
			break
		}
		// The other party I actively connect to must be a public node, so there is no need to detect and judge
		peerId := msgbody[0:PeerIDSize]
		peerName := string(msgbody[PeerIDSize:])
		peer.ID = peerId
		peer.Name = strings.TrimRight(peerName, " ")
		// Add public network node table
		p.PeerChangeMux.Lock()
		p.addPeerAllNodesUnsafe(peer)
		p.addPeerIntoTargetTableUnsafe(&p.BackboneNodeTable, p.Config.BackboneNodeTableSizeMax, peer)
		p.PeerChangeMux.Unlock()
		// Notification connection succeeded
		peer.notifyConnect()
		// Notify the other party to be a public network node
		go sendTcpMsg(conn, P2PMsgTypeRemindMeIsPublicPeer, nil)
		// Determine whether to perform the first node lookup
		p.PeerChangeMux.RLock()
		var backbonetablelen = len(p.BackboneNodeTable)
		p.PeerChangeMux.RUnlock()

		if backbonetablelen == 1 {
			p.findNodes()
		}
		break

	case P2PMsgTypeReportIdKeepConnectAsPeer:
		// The other party requests a persistent connection
		//fmt.Println("P2PMsgTypeReportIdKeepConnectAsPeer")
		if len(msgbody) != 4+PeerIDSize+PeerNameSize {
			break
		}
		if peer.ID != nil {
			break
		}
		port := binary.BigEndian.Uint32(msgbody[0:4])
		peerId := msgbody[4 : 4+PeerIDSize]
		peerName := string(msgbody[4+PeerIDSize:])
		peer.ID = peerId
		peer.Name = strings.TrimRight(peerName, " ")
		// Connect join node
		oldPeerIsBackboneNode := false // 旧节点是否为骨干节点
		if oldp, hasp := p.AllNodes.Load(string(peerId)); hasp {
			oldpeer := oldp.(*Peer)
			// What should I do if it already exists?
			oldPeerIsBackboneNode = oldpeer.PublicIpPort != nil
			peer.ReplacingCopyInfo(oldpeer)
			oldpeer.Disconnect()        // Disconnect old connections directly
			time.Sleep(time.Second * 1) // Sleep for 1 second
		}
		// Reply to the message that I am willing to connect
		rplidbuf := bytes.NewBuffer(p.peerSelf.ID)
		rplidbuf.Write(p.peerSelf.NameBytes())
		e3 := sendTcpMsg(conn, P2PMsgTypeAnswerIdKeepConnectAsPeer, rplidbuf.Bytes())
		if e3 != nil {
			// Error sending message
			conn.Close()
			break
		}
		// Add as new node
		p.PeerChangeMux.Lock()
		p.addPeerAllNodesUnsafe(peer)
		p.addPeerIntoUnfamiliarTableUnsafe(peer)
		p.PeerChangeMux.Unlock()

		newPeerIsBackboneNode := false // 新节点是否为骨干节点
		go func() {
			// Notify the official connection
			defer func() {
				// Backbone node mentions node table upgrade
				p.PeerChangeMux.Lock()
				bkbndsLess := len(p.BackboneNodeTable) < p.Config.BackboneNodeTableSizeMax
				if newPeerIsBackboneNode && (oldPeerIsBackboneNode || bkbndsLess) {
					if rmok, newtb := removePeerIDFromTable(p.UnfamiliarNodesTable, peer.ID); rmok {
						p.UnfamiliarNodesTable = newtb // Removed successfully
						// 放入公网节点表 // fmt.Println("放入公网节点表")
						p.addPeerIntoTargetTableUnsafe(&p.BackboneNodeTable, p.Config.BackboneNodeTableSizeMax, peer)
					}
				}
				p.PeerChangeMux.Unlock()
				peer.notifyConnect()
			}()
			// Send a message to determine whether it is a public network node
			tcp, e0 := net.ResolveTCPAddr("tcp", conn.RemoteAddr().String())
			if e0 != nil {
				return
			}
			tcp.Port = int(port) // Public network listening port
			// Try to connect
			isclosed := false
			ckpubconn, e1 := dialTimeoutWithHandshakeSignal("tcp", tcp.String(), time.Second*5)
			if e1 != nil {
				return
			}
			go func() {
				// Close after 10 seconds
				time.Sleep(time.Second * 10)
				if !isclosed {
					ckpubconn.Close()
				}
			}()
			e3 := sendTcpMsg(ckpubconn, P2PMsgTypeRequestIDForPublicNodeCheck, nil)
			if e3 != nil {
				isclosed = true
				ckpubconn.Close()
				return
			}
			checkpid := make([]byte, PeerIDSize)
			rdn, e2 := io.ReadFull(ckpubconn, checkpid)
			if e2 != nil {
				isclosed = true
				ckpubconn.Close()
				return
			}
			if rdn == PeerIDSize {
				// Judge whether the node is a public network node
				if bytes.Compare(checkpid, peerId) == 0 {
					//fmt.Println("OK PublicIpPort:", hex.EncodeToString(peerId), tcp.String())
					peer.PublicIpPort = tcp // Write public network node
					newPeerIsBackboneNode = true
					// Notify the other party to be a public network node
					go sendTcpMsg(conn, P2PMsgTypeRemindMeIsPublicPeer, nil)
				}
			}
			isclosed = true
			ckpubconn.Close()
		}()

		break

	}

}
