package handler

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
)

func GetBlocksData(blockchain interfaces.BlockChain, peer interfaces.MsgPeer, msgbody []byte) {
	if len(msgbody) < 3*8 {
		return
	}
	lastestHeight := binary.BigEndian.Uint64(msgbody[0:8])
	startHeight := binary.BigEndian.Uint64(msgbody[8:16])
	endHeight := binary.BigEndian.Uint64(msgbody[16:24])
	realEndHeight := uint64(0)
	allBlockDatas := msgbody[24:]
	alldtslen := len(allBlockDatas)
	// print
	fmt.Print("got blocks: ", startHeight, " ~ ", endHeight, ", inserting... ")
	// parse block
	seek := uint32(0)
	for {
		if seek >= uint32(alldtslen) {
			break // end
		}
		oneblock, sk, err := blocks.ParseBlock(allBlockDatas, seek)
		if err != nil || oneblock == nil {
			return // block data error
		}
		if seek == 0 && startHeight != oneblock.GetHeight() {
			return // block error
		}
		seek = sk
		// append
		insert_error := blockchain.InsertBlock(oneblock)
		if insert_error != nil {
			fmt.Println("[Error] GetBlocksData to InsertBlock:", insert_error)
			return
		} else {
			//fmt.Println("++++ InsertBlock:", oneblock.GetHeight())
		}
		realEndHeight = oneblock.GetHeight()
	}
	fmt.Println("OK")
	if realEndHeight == lastestHeight {
		fmt.Println("all block sync successfully.")
	}
	// req next datas
	if endHeight != realEndHeight || realEndHeight >= lastestHeight {
		return
	}
	// req next data
	msgParseSendRequestBlocks(peer, endHeight+1)
}

func SendBlocksData(blockchain interfaces.BlockChain, peer interfaces.MsgPeer, msgbody []byte) {

	//fmt.Println("SendBlocksData", msgbody)

	if len(msgbody) != 8 {
		return
	}
	mylastblock, err := blockchain.State().ReadLastestBlockHeadAndMeta()
	if err != nil {
		return
	}
	blockstore := blockchain.State().BlockStore()
	lastestHeight := mylastblock.GetHeight()
	startHeight := binary.BigEndian.Uint64(msgbody)
	endHeight := uint64(0)
	// read block data
	readatas := bytes.NewBuffer(bytes.Repeat([]byte{0}, 8*3))
	maxsendsize := int(1024 * 512)
	totalsize := 0
	for curhei := startHeight; curhei <= lastestHeight; curhei++ {
		_, oneblkbts, err := blockstore.ReadBlockBytesByHeight(curhei, 0)
		if err != nil {
			return
		}
		totalsize += len(oneblkbts)
		readatas.Write(oneblkbts)
		if totalsize >= maxsendsize || curhei == lastestHeight {
			endHeight = curhei
			break //
		}
	}
	// send msg data
	sendmsgdatas := readatas.Bytes()
	binary.BigEndian.PutUint64(sendmsgdatas[0:8], lastestHeight)
	binary.BigEndian.PutUint64(sendmsgdatas[8:16], startHeight)
	binary.BigEndian.PutUint64(sendmsgdatas[16:24], endHeight)
	//
	fmt.Println("send to ", peer.Describe(), " blocks data height: ", startHeight, " ~ ", endHeight)
	// send
	peer.SendDataMsg(MsgTypeBlocks, sendmsgdatas)
}
