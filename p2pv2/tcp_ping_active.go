package p2pv2

import "time"

/**
 * 持续的 ping 保持 tcp 连接的保活
 */
func (p *P2P) continuePingToKeepTcpLinkAlive() {

	// 如果我不是公网节点，则没隔 127 秒向全部节点发送一次 ping 消息保持活跃在线
	// 直到有其他节点告诉我我是公网节点为止
	for {
		time.Sleep(time.Second * 127)
		if p.MyselfIsPublicPeer {
			break // 我是公网节点，停止ping保活
		}
		// 给所有节点发送ping消息
		for _, peer := range p.AllNodes {
			if peer != nil {
				sendTcpMsg(peer.conn, P2PMsgTypePing, nil) // send ping
			}
		}
	}

}
