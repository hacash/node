package handler

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"sync"
)

var sendBlockDataMutex sync.Mutex

func GetBlocksData(p2p interfaces.P2PManager, cmtr interfaces.P2PMsgCommunicator, blockchain interfaces.BlockChain, peer interfaces.P2PMsgPeer, msgbody []byte) {
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
			fmt.Println("blocks.ParseBlock Error: ", err)
			return // block data error
		}
		if seek == 0 && startHeight != oneblock.GetHeight() {
			fmt.Println("seek == 0 && startHeight != oneblock.GetHeight()")
			return // block error
		}
		seek = sk
		// append
		insert_error := blockchain.GetChainEngineKernel().InsertBlock(oneblock.(interfaces.Block), "sync")
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
	// change peer
	if p2p.GetConfigOfBootNodeFastSync() == false {
		acpubpeer := cmtr.FindAnyOnePeerBetterBePublic() // 请求一个新节点
		if acpubpeer != nil {
			peer = acpubpeer // ac != nil
		}
	}
	msgParseSendRequestBlocks(peer, endHeight+1)
}

func SendBlocksData(blockchain interfaces.BlockChain, peer interfaces.P2PMsgPeer, msgbody []byte) {
	sendBlockDataMutex.Lock()
	defer sendBlockDataMutex.Unlock()

	//fmt.Println("SendBlocksData", msgbody)
	if len(msgbody) != 8 {
		return
	}
	mylastblock, _, err := blockchain.GetChainEngineKernel().LatestBlock()
	if err != nil {
		return
	}
	blockstore := blockchain.GetChainEngineKernel().StateRead().BlockStoreRead()
	lastestHeight := mylastblock.GetHeight()
	startHeight := binary.BigEndian.Uint64(msgbody)
	endHeight := uint64(0)
	// read block data
	readatas := bytes.NewBuffer(bytes.Repeat([]byte{0}, 8*3))
	maxsendblknum := int(1000)
	maxsendsize := int(1024 * 1024 * 8) // 最大发送不超过 8 MB
	totalsize := 0
	totalblknum := 0
	for curhei := startHeight; curhei <= lastestHeight; curhei++ {
		//fmt.Println("curhei", curhei)
		_, oneblkbts, err := blockstore.ReadBlockBytesByHeight(curhei)
		//fmt.Println("curhei", curhei, "ReadBlockBytesLengthByHeight")
		if err != nil {
			fmt.Println("P2P Message SendBlocksData ReadBlockBytesLengthByHeight Error:", err.Error())
			return
		}
		if oneblkbts == nil {
			endHeight = curhei - 1
			break // not find
		}
		totalblknum += 1
		totalsize += len(oneblkbts)
		readatas.Write(oneblkbts)
		if curhei == lastestHeight || totalblknum >= maxsendblknum || totalsize >= maxsendsize {
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
