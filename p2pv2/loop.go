package p2pv2

import (
	"fmt"
	"time"
)

func (p *P2P) loop() {

	pingBackboneNodeTiker := time.NewTicker(time.Minute * 3)
	checkBackboneNodeActiveTiker := time.NewTicker(time.Minute * 5)
	findNodesTiker := time.NewTicker(time.Minute * 17)
	reconnectBootNodesTiker := time.NewTicker(time.Hour * 6)  // 6小时boot重连一次
	upgradeNodeLevelTiker := time.NewTicker(time.Second * 70) // 提升节点等级 70s

	for {

		select {

		case <-upgradeNodeLevelTiker.C:
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

		case <-reconnectBootNodesTiker.C:
			// 尝试重连静态节点
			p.tryConnectToStaticBootNodes()
			break

		case <-pingBackboneNodeTiker.C:
			// 给所有节点发送ping消息
			ct := time.Now()
			for _, peer := range p.AllNodes {
				if peer != nil {
					if peer.activeTime.Add(time.Minute * 5).Before(ct) {
						sendTcpMsg(peer.conn, P2PMsgTypePing, nil) // send ping
					}
				}
			}

		case <-checkBackboneNodeActiveTiker.C:
			// 检查所有节点的活跃度
			ct := time.Now()
			for _, peer := range p.AllNodes {
				if peer != nil {
					if peer.activeTime.Add(time.Minute * 10).Before(ct) {
						// 10分钟无消息，为失去活跃的节点
						peer.ActiveOvertime = true
						peer.Disconnect()
					}
				}
			}

		case <-findNodesTiker.C:
			if len(p.BackboneNodeTable) == 0 {
				// 没有公网节点连接时，尝试连接 Static Boot Nodes
				p.tryConnectToStaticBootNodes()
			} else {
				// 寻找最近的节点
				p.findNodes()
			}
			// 打印最新的连接情况
			fmt.Printf("[P2P] connected peers: %d public, %d private, total: %d nodes, %d conns.\n",
				len(p.BackboneNodeTable), len(p.OrdinaryNodeTable), len(p.AllNodes), len(p.TemporaryConns))

		}

	}

}
