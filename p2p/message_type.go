package p2p

// version
const P2PMustVersion = uint16(1)

// TCP
const (
	TCPMsgTypePing uint16 = 1
	TCPMsgTypePong uint16 = 2

	TCPMsgTypeHandshake        uint16 = 3
	TCPMsgTypeHandshakeSuccess uint16 = 4

	TCPMsgTypeConnectRefuse uint16 = 5

	TCPMsgTypeGetPublicConnectedPeerAddrs uint16 = 6 // 请求公网节点数据
	TCPMsgTypePublicConnectedPeerAddrs    uint16 = 7 // 收到公网节点数据

	TCPMsgTypeReplyPublic  uint16 = 8 // yes im public ip
	TCPMsgTypeTellPublicIP uint16 = 9 // tell me is public

	TCPMsgTypeDiscoverPublicPeerJoin uint16 = 10

	// customer data
	TCPMsgTypeData uint16 = 65535
)

// UDP
const (
	UDPMsgTypeEnquirePublic uint16 = 20001 // is public ip node ?

)
