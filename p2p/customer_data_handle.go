package p2p

type MsgDataHandler interface {
	OnConnected(cator MsgCommunicator, p *Peer)

	OnMsgData(cator MsgCommunicator, p *Peer, msgty uint16, msgbody []byte)

	OnDisconnected(*Peer)
}

type MsgCommunicator interface {
	PeerLen() int

	FindRandomOnePeerBetterBePublic() *Peer

	BroadcastMessageToUnawarePeers(ty uint16, msgbody []byte, KnowledgeKey string, KnowledgeValue string)
}
