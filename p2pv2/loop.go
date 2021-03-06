package p2pv2

import (
	"time"
)

func (p *P2P) loop() {

	pingAllNodesTiker := time.NewTicker(time.Minute * 3)
	checkNodesActiveTiker := time.NewTicker(time.Minute * 5)
	findNodesTiker := time.NewTicker(time.Minute * 77) // 77分钟 findnodes 一次
	//findNodesTiker := time.NewTicker(time.Second * 15)            // 测试
	upgradeNodeLevelTiker := time.NewTicker(time.Second * 70) // 提升节点等级 70s
	//forceReconnectBootNodesTiker := time.NewTicker(time.Hour * 8) // 8小时强制boot重连一次

	// 任务
	for {

		select {

		case <-upgradeNodeLevelTiker.C:
			if len(p.BackboneNodeTable) == 0 {
				// 没有公网节点连接时，尝试连接 Static Boot Nodes
				p.tryConnectToStaticBootNodes()
			} else {
				ct := time.Now()
				// 提升节点等级
				// 将临时区域内连接超过半小时的节点提升到公网或私网表
				p.PeerChangeMux.Lock()
				if len(p.UnfamiliarNodesTable) > 0 {
					curpid := p.UnfamiliarNodesTable[0]
					peer := p.getPeerByID(curpid)
					if peer == nil {
						// 删除无效的
						p.UnfamiliarNodesTable = p.UnfamiliarNodesTable[1:]
					} else {
						// 检查时间, 已经连接30分钟
						if peer.createTime.Add(time.Minute * 30).Before(ct) {
							p.upgradeOneUnfamiliarNodeLevelUnsafe() // 提升最后节点的等级
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
			// 给所有节点发送ping消息
			ct := time.Now()
			p.AllNodes.Range(func(key, value interface{}) bool {
				peer := value.(*Peer)
				if peer != nil {
					// 超过5分钟没有活跃的节点
					if peer.activeTime.Add(time.Minute * 5).Before(ct) {
						sendTcpMsg(peer.conn, P2PMsgTypePing, nil) // send ping
					}
				}
				return true
			})

		case <-checkNodesActiveTiker.C:
			// 检查所有节点的活跃度
			ct := time.Now()
			p.AllNodes.Range(func(key, value interface{}) bool {
				peer := value.(*Peer)
				if peer != nil {
					if peer.activeTime.Add(time.Minute * 10).Before(ct) {
						// 10分钟无消息，为失去活跃的节点
						peer.RemoveActiveOvertime = true
						peer.Disconnect()
					}
				}
				return true
			})

		case <-findNodesTiker.C:
			if len(p.BackboneNodeTable) == 0 {
				// 没有公网节点连接时，尝试连接 Static Boot Nodes
				p.tryConnectToStaticBootNodes()
			} else {
				// 寻找最近的节点
				p.findNodes()
			}
		}

	}

}
