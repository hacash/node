package p2p

func (p2p *P2PManager) AddKnowledge(KnowledgeKey string, KnowledgeValue string) bool {
	return p2p.myselfpeer.AddKnowledge(KnowledgeKey, KnowledgeValue)
}

func (p2p *P2PManager) CheckKnowledge(KnowledgeKey string, KnowledgeValue string) bool {
	return p2p.myselfpeer.CheckKnowledge(KnowledgeKey, KnowledgeValue)
}
