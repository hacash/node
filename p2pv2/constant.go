package p2pv2

const (
	P2PHandshakeSignal uint32 = 3418609257               // Fixed signal value
	P2PMsgDataMaxSize  uint32 = uint32(10) * 1024 * 1024 // Maximum length of a single message 10 MB

	PeerNameSize int = 16
	PeerIDSize   int = 16
)

const (
	// 请求/应答消息
	P2PMsgTypeReportIdKeepConnectAsPeer uint8 = 1 // Report my port + peerid + peername, and request that you want to connect as a persistent node
	P2PMsgTypeAnswerIdKeepConnectAsPeer uint8 = 2 // Reply to my peerid + peername and agree to use it as a persistent connection
	P2PMsgTypePing                      uint8 = 3 // ping
	P2PMsgTypePong                      uint8 = 4 // pong

	// Message without reply
	P2PMsgTypeRemindMeIsPublicPeer uint8 = 151 // The other party reminds me that I am a public network node

	// Message to disconnect immediately after reply
	P2PMsgTypeRequestIDForPublicNodeCheck uint8 = 201 // My peerid is used to judge whether it is a public network. You can disconnect immediately after you reply
	P2PMsgTypeRequestNearestPublicNodes   uint8 = 202 // The other party requests the public network node list (within 200), and can immediately disconnect after replying

	// Customer upper level message
	P2PMsgTypeCustomer uint8 = 255
)
