package p2p

type MsgPeer interface {
	AddKnowledge(KnowledgeKey string, KnowledgeValue string) bool
	SendDataMsg(msgty uint16, msgbody []byte)
	Describe() string
	Disconnect()
}

type MsgDataHandler interface {
	OnConnected(cator MsgCommunicator, p MsgPeer)
	OnMsgData(cator MsgCommunicator, p MsgPeer, msgty uint16, msgbody []byte)
	OnDisconnected(MsgPeer)
}

type MsgCommunicator interface {
	PeerLen() int
	FindAnyOnePeerBetterBePublic() MsgPeer
	BroadcastMessageToUnawarePeers(ty uint16, msgbody []byte, KnowledgeKey string, KnowledgeValue string)
}
