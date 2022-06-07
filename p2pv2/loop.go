package p2pv2

import (
	"time"
)

func (p *P2P) loop() {

	pingAllNodesTiker := time.NewTicker(time.Minute * 3)
	checkNodesActiveTiker := time.NewTicker(time.Minute * 5)
	findNodesTiker := time.NewTicker(time.Minute * 77) // 77分钟 findnodes 一次
	// findNodesTiker:= time.NewTicker(time.Second * 15)            // 测试
	upgradeNodeLevelTiker := time.NewTicker(time.Second * 70) // 提升节点等级 70s
	// forceReconnectBootNodesTiker:= time.NewTicker(time.Hour * 8) // 8小时强制boot重连一次

	// task
	for {

		select {

		case <-upgradeNodeLevelTiker.C:
			if len(p.BackboneNodeTable) == 0 {
				// Try to connect to static boot nodes when no public nodes are connected
				p.tryConnectToStaticBootNodes()
			} else {
				ct := time.Now()
				// Upgrade node level
				// Promote the nodes in the temporary area that have been connected for more than half an hour to the public or private network table
				p.PeerChangeMux.Lock()
				if len(p.UnfamiliarNodesTable) > 0 {
					curpid := p.UnfamiliarNodesTable[0]
					peer := p.getPeerByID(curpid)
					if peer == nil {
						// Delete invalid
						p.UnfamiliarNodesTable = p.UnfamiliarNodesTable[1:]
					} else {
						// Check time, connected for 30 minutes
						if peer.createTime.Add(time.Minute * 30).Before(ct) {
							p.upgradeOneUnfamiliarNodeLevelUnsafe() // Increase the level of the last node
						}
					}
				}
				p.PeerChangeMux.Unlock()
				break
			}

		//case <-forceReconnectBootNodesTiker.C:
		//	// 尝试重连静态节点
		//	p.tryConnectToStaticBootNodes()
		//	break

		case <-pingAllNodesTiker.C:
			// Send Ping message to all nodes
			ct := time.Now()
			p.AllNodes.Range(func(key, value interface{}) bool {
				peer := value.(*Peer)
				if peer != nil {
					// No active nodes for more than 5 minutes
					p.PeerChangeMux.RLock()
					if peer.activeTime.Add(time.Minute * 5).Before(ct) {
						go sendTcpMsg(peer.conn, P2PMsgTypePing, nil) // send ping
					}
					p.PeerChangeMux.RUnlock()
				}
				return true
			})

		case <-checkNodesActiveTiker.C:
			// Check the activity of all nodes
			ct := time.Now()
			p.AllNodes.Range(func(key, value interface{}) bool {
				peer := value.(*Peer)
				if peer != nil {
					if peer.activeTime.Add(time.Minute * 10).Before(ct) {
						// If there is no message for 10 minutes, the node is inactive
						peer.RemoveActiveOvertime = true
						peer.Disconnect()
					}
				}
				return true
			})

		case <-findNodesTiker.C:
			if len(p.BackboneNodeTable) == 0 {
				// Try to connect to static boot nodes when no public nodes are connected
				p.tryConnectToStaticBootNodes()
			} else {
				// Find nearest node
				p.findNodes()
			}
		}

	}

}
