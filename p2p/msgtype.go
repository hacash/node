package p2p

// version
const P2PMustVersion = uint16(1)

// msg type
const (

	// tcp
	MsgTypePing = uint16(1)
	MsgTypePong = uint16(2)

	MsgTypeHandShake = uint16(3)

	MsgTypeOtherNodeWantToConnect  = uint16(4)
	MsgTypeAllowOtherNodeToConnect = uint16(5)

	MsgTypeNotifyIsPublicNode = uint16(6)

	MsgTypeDiscoverNewNodeJoin = uint16(7)

	MsgTypeConnectRefuse = uint16(8)

	// udp dial
	MsgTypeUDPWantToConnectNode  = uint16(20001)
	MsgTypeUDPAllowToConnectNode = uint16(20002)

	MsgTypeReportTCPListenPort = uint16(20003)

	// other data
	MsgTypeData = uint16(65535)
)
