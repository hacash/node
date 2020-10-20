package p2pv2

const (
	P2PHandshakeSignal uint32 = 3418609527 // 固定不变的信号值

	PeerNameSize int = 16
	PeerIDSize   int = 16
)

const (
	// 请求/应答消息
	P2PMsgTypeReportIdKeepConnectAsPeer uint8 = 1 // 报告我的 Port + PeerID + PeerName ，请求希望作为持久节点来连接
	P2PMsgTypeAnswerIdKeepConnectAsPeer uint8 = 2 // 回复我的 PeerID + PeerName 同意作为持久连接
	P2PMsgTypePing                      uint8 = 3 // ping
	P2PMsgTypePong                      uint8 = 4 // pong

	// 无需回复的消息
	P2PMsgTypeRemindMeIsPublicPeer uint8 = 151 // 对方提示我自己是公网节点

	// 答复后立即断开连接的消息
	P2PMsgTypeRequestIDForPublicNodeCheck uint8 = 201 // 对方询问我的 PeerID 用于判断是否为公网，答复后即可立即断开连接
	P2PMsgTypeRequestNearestPublicNodes   uint8 = 202 // 对方请求公网节点列表(200以内)， 答复后可立即断开连接

	// 客户上层消息
	P2PMsgTypeCustomer uint8 = 255
)
