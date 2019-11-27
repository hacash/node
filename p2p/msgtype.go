package p2p

// version
const P2PMustVersion = uint16(1)

// msg type
const (

	// tcp
	MsgTypePing = uint16(1)
	MsgTypePong = uint16(2)

	MsgTypeHandShake = uint16(3)

	MsgTypeOtherNodeWantToConnect = uint16(4)

	MsgTypeNotifyIsPublicNode = uint16(5)

	MsgTypeDiscoverNewNodeJoin = uint16(6)

	// udp dial
	MsgTypeReportTCPListenPort = uint16(20001)
	MsgTypeWantToConnectNode   = uint16(20002)

	// other data
	MsgTypeData = uint16(65535)
)
