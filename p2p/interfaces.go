package p2p

type MsgPeer interface {
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
	FindRandomOnePeerBetterBePublic() MsgPeer
	BroadcastMessageToUnawarePeers(ty uint16, msgbody []byte, KnowledgeKey string, KnowledgeValue string)
}
