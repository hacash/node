package handler

import (
	"encoding/binary"
	"fmt"
	"github.com/hacash/core/interfaces"
)

const (
	MsgTypeRequestStatus uint16 = 1
	MsgTypeStatus        uint16 = 2

	MsgTypeRequestBlockHashList uint16 = 3
	MsgTypeBlockHashList        uint16 = 4

	MsgTypeRequestBlocks uint16 = 5
	MsgTypeBlocks        uint16 = 6

	MsgTypeSubmitTransaction uint16 = 7
	MsgTypeDiscoverNewBlock  uint16 = 8

	MsgTypeRequestTxDatas uint16 = 9
	MsgTypeTxDatas        uint16 = 10
)

func msgParseSendRequestBlocks(peer interfaces.MsgPeer, startheigit uint64) {
	fmt.Print("sync blocks: ", startheigit, "... ")
	startheight := make([]byte, 8)
	binary.BigEndian.PutUint64(startheight, startheigit)
	peer.SendDataMsg(MsgTypeRequestBlocks, startheight)
}

func msgParseSendRequestBlockHashList(peer interfaces.MsgPeer, reqnum uint8, startheigit uint64) {
	buf := make([]byte, 1+8)
	buf[0] = reqnum
	binary.BigEndian.PutUint64(buf[1:], startheigit)
	peer.SendDataMsg(MsgTypeRequestBlockHashList, buf)
}

////////////////////////////////////////////////

func printWarning(content string) {
	upgrade_tip :=
		"\n\n/*-*-*-*-*-*-*-*-*-*-*- warning start -*-*-*-*-*-*-*-*-*-*-*/\n\n" + content +
			"\n\n/*-*-*-*-*-*-*-*-*-*-*- warning end -*-*-*-*-*-*-*-*-*-*-*-*/\n\n"
	fmt.Println(upgrade_tip)
}
