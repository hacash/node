package p2p

import "time"

func (p2p *P2PManager) loop() {

	dropUnHandShakeTiker := time.NewTicker(time.Second * 13)
	dropNotReplyPublicTiker := time.NewTicker(time.Second * 5)
	reconnectSomePublicPeerTiker := time.NewTicker(time.Second * 21)
	getAddrsFromPublicPeersTiker := time.NewTicker(time.Minute * 23)

	for {

		select {

		case <-getAddrsFromPublicPeersTiker.C:
			go func() {
				pubpeers := p2p.peerManager.publicPeerGroup.peers.ToSlice()
				for _, p := range pubpeers {
					peer := p.(*Peer)
					peer.SendMsg(TCPMsgTypeGetPublicConnectedPeerAddrs, nil)
					<-time.Tick(time.Second * 7)
				}
			}()

		case <-reconnectSomePublicPeerTiker.C:

			// -------- TEST START --------
			//peers := p2p.peerManager.publicPeerGroup.peers.ToSlice()
			//peers = append(peers, p2p.peerManager.interiorPeerGroup.peers.ToSlice()...)
			//allpnames := ""
			//for _, p := range peers {
			//	allpnames += p.(*Peer).Name + "  "
			//}
			//fmt.Println("current peers num ", len(peers), "(  "+allpnames+")")
			// -------- TEST END   --------

			go func() {
				publicconnCount := p2p.peerManager.publicPeerGroup.peers.Cardinality()
				if publicconnCount < 3 {
					ipport := p2p.recordOldPublicPeerTCPAddrs.Pop()
					if ipport == nil && publicconnCount == 0 {
						ipport = p2p.recordStaticPublicPeerTCPAddrs.Pop()
						if ipport != nil {
							p2p.recordStaticPublicPeerTCPAddrs.Add(ipport) // reput in
						}
					}
					if ipport != nil {
						ipports := []byte(ipport.(string))
						addr := ParseIPPortToTCPAddrByByte(ipports)
						if addr != nil {
							go func() {
								connerr := p2p.TryConnectToPeer(nil, addr)
								if connerr == nil { // reput in
									// do not add // p2p.AddOldPublicPeerAddrByBytes(ipports)
								}
							}()
						}
					} else {
						//fmt.Println(fmt.Errorf("Cannot get any peer addr to connect."))
					}
				}
			}()

		case <-dropNotReplyPublicTiker.C:
			go func() {
				p2p.changeStatusLock.Lock()
				peers := make([]*Peer, 0, len(p2p.waitToReplyIsPublicPeer))
				ress := make([]*struct {
					curt time.Time
					code uint32
				}, 0, len(p2p.waitToReplyIsPublicPeer))
				for peer, res := range p2p.waitToReplyIsPublicPeer {
					peers = append(peers, peer)
					ress = append(ress, &res)
				}
				tnow := time.Now()
				for i := 0; i < len(peers); i++ {
					peer, res := peers[i], ress[i]
					if res.curt.Add(time.Second * 5).Before(tnow) {
						delete(p2p.waitToReplyIsPublicPeer, peer)
						if peer.publicIPv4 == nil {
							p2p.AddPeerToTargetGroup(p2p.peerManager.interiorPeerGroup, peer)
						}
					}
				}
				p2p.changeStatusLock.Unlock()
			}()

		case <-dropUnHandShakeTiker.C:
			tnow := time.Now()
			peers := p2p.lookupPeers.ToSlice()
			go func() {
				for _, p := range peers {
					peer := p.(*Peer)
					if len(peer.ID) == 0 {
						if peer.connTime.Add(time.Second * 9).Before(tnow) {
							peer.TcpConn.Close() // drop it over time not do hand shake
						}
					}
				}
			}()

		}

	}

}
