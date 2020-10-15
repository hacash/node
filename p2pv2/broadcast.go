package p2pv2

/**
 * 广播消息
 */
func (p *P2P) broadcastMessageToUnawarePeers(ty uint16, msgbody []byte, KnowledgeKind string, KnowledgeKey string) error {
	var peers = []*Peer{}
	for _, peer := range p.AllNodes {
		peers = append(peers, peer)
	}
	go func() {
		for _, peer := range peers {
			peer.SendUnawareMsg(ty, msgbody, KnowledgeKind, KnowledgeKey)
		}
	}()
	return nil
}
