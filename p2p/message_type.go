package p2p

// version
const P2PMustVersion = uint16(1)

// TCP
const (
	TCPMsgTypePing uint16 = 1
	TCPMsgTypePong uint16 = 2

	TCPMsgTypeHandShake     uint16 = 3
	TCPMsgTypeConnectRefuse uint16 = 4

	TCPMsgTypeGetPublicConnectedPeerAddrs uint16 = 5
	TCPMsgTypePublicConnectedPeerAddrs    uint16 = 6

	TCPMsgTypeReplyPublic  uint16 = 7 // yes im public ip
	TCPMsgTypeTellPublicIP uint16 = 8 // tell me is public

	TCPMsgTypeDiscoverPublicPeerJoin uint16 = 9

	// customer data
	TCPMsgTypeData uint16 = 65535
)

// UDP
const (
	UDPMsgTypeEnquirePublic uint16 = 20001 // is public ip node ?

)
