package handle

const (
	MsgTypeRequestStatus uint16 = 1
	MsgTypeStatus        uint16 = 2

	MsgTypeRequestBlocks uint16 = 3
	MsgTypeBlocks        uint16 = 4

	MsgTypeRequestBlockHashList uint16 = 5
	MsgTypeBlockHashList        uint16 = 6

	MsgTypeSubmitTransaction uint16 = 7
	MsgTypeDiscoverNewBlock  uint16 = 8
)
