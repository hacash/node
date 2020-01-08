package handler

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/hacash/core/interfaces"
	"sync"
)


var sendBlockHashListMutex sync.Mutex


func SendBlockHashList(blockchain interfaces.BlockChain, peer interfaces.MsgPeer, msgbody []byte) {
	if len(msgbody) != 1+8 {
		return // error len
	}
	reqnum := uint64(uint8(msgbody[0]))
	if reqnum > 80 {
		reqnum = 80 // max len 80
	}

	sendBlockHashListMutex.Lock()
	defer sendBlockHashListMutex.Unlock()

	reqlastheight := binary.BigEndian.Uint64(msgbody[1:])
	mylastblk, err := blockchain.State().ReadLastestBlockHeadAndMeta()
	if err != nil {
		return
	}
	if mylastblk.GetHeight() < reqlastheight {
		return // not find this block
	}
	// read from store
	blockstore := blockchain.State().BlockStore()
	resdatas := bytes.NewBuffer(msgbody[1:])
	for curhei := reqlastheight; curhei > 0 && curhei >= reqlastheight-reqnum; curhei-- {
		// read blk hash
		oneblkhx, err := blockstore.ReadBlockHashByHeight(curhei)
		if err != nil {
			return // error
		}
		resdatas.Write(oneblkhx)
	}
	// send msg block hash list
	peer.SendDataMsg(MsgTypeBlockHashList, resdatas.Bytes())
}

func GetBlockHashList(blockchain interfaces.BlockChain, peer interfaces.MsgPeer, msgbody []byte) {
	if len(msgbody) < 8 {
		return // error len
	}
	lastestHeight := binary.BigEndian.Uint64(msgbody[0:8])
	mylastblk, err := blockchain.State().ReadLastestBlockHeadAndMeta()
	if err != nil {
		return
	}
	if lastestHeight > mylastblk.GetHeight() {
		return // not find target height block
	}
	allhashes := msgbody[8:]
	if len(allhashes)%32 != 0 {
		return // error len
	}
	bigHei := lastestHeight
	hashnum := uint64((len(allhashes) / 32))
	if hashnum > lastestHeight {
		return
	}
	if hashnum == 0 {
		return
	}
	smallHei := lastestHeight - hashnum
	rollbackToHeight := uint64(0)
	// read my block hash
	blockstore := blockchain.State().BlockStore()
	i := 0
	for curhei := bigHei; curhei > smallHei; curhei-- {
		myheihash, e := blockstore.ReadBlockHashByHeight(curhei)
		if e != nil {
			return
		}
		tarhash := allhashes[i*32 : i*32+32]
		equalForNow := myheihash.Equal(tarhash)
		if curhei == bigHei && equalForNow {
			// get blocks data
			msgParseSendRequestBlocks(peer, curhei+1)
			return
		}
		if equalForNow {
			rollbackToHeight = curhei
		}
		i++
	}

	if rollbackToHeight > 0 {
		curhei, err := blockchain.RollbackToBlockHeight(rollbackToHeight)
		if err != nil {
			fmt.Println("RollbackToBlockHeight error", err)
			return
		}
		// get block data
		msgParseSendRequestBlocks(peer, curhei+1)
		return
	}

	// fork too long
	if hashnum == 80 {
		printWarning(
			"[Warning] Block data fork is serious and cannot be synchronized. Please delete all data and restart.\n" +
				"【警告】你的区块数据分叉严重，已超过80个区块与全网不匹配，若想恢复请删除全部数据，并从头开始同步区块。")
		return
	}

	// not find any equal hash
	// request more hash list  len = 80
	if hashnum == 4 && lastestHeight > 80 {
		msgParseSendRequestBlockHashList(peer, 80, mylastblk.GetHeight())
	}
}
