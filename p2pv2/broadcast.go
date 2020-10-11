package p2pv2

/**
 * 广播消息
 */
func (p *P2P) broadcastMessageToUnawarePeers(ty uint16, msgbody []byte, KnowledgeKind string, KnowledgeKey string) error {
	go func() {
		p.PeerChangeMux.Lock()
		defer p.PeerChangeMux.Unlock()
		for _, peer := range p.AllNodes {
			peer.SendUnawareMsg(ty, msgbody, KnowledgeKind, KnowledgeKey)
		}
	}()
	return nil
}
