package p2pv2

/**
 * 广播消息
 */
func (p *P2P) broadcastMessageToUnawarePeers(ty uint16, msgbody []byte, KnowledgeKind string, KnowledgeKey string) error {

	p.AllNodes.Range(func(key, value interface{}) bool {
		peer := value.(*Peer)
		peer.SendUnawareMsg(ty, msgbody, KnowledgeKind, KnowledgeKey)
		return true
	})
	return nil
}
