package handler

import (
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
	"sync/atomic"
	"time"
)

const (
	time_format_layout = "01/02 15:04:05"
)

// 当前正在执行插入的区块高度
var currentInsertingBlockHeight uint64 = 0

func GetBlockDiscover(p2p interfaces.P2PManager, blockchain interfaces.BlockChain, peer interfaces.P2PMsgPeer, msgbody []byte) {
	//fmt.Println("GetBlockDiscover", 1)
	// 先解析区块头，以提升并发收到区块时的性能
	blockhead, _, e0 := blocks.ParseBlockHead(msgbody, 0)
	if e0 != nil {
		return // error end
	}
	curblkhei := blockhead.GetHeight()
	// 判断正在插入的区块高度
	if atomic.CompareAndSwapUint64(&currentInsertingBlockHeight, 0, curblkhei) {
		// 未插入区块，继续插入
	} else {
		// 判断插入的区块是否和当前的相同
		if atomic.CompareAndSwapUint64(&currentInsertingBlockHeight, curblkhei, curblkhei) {
			return // 正在插入相同的区块，忽略消息
		}
		// 标记正在插入区块的高度
		atomic.StoreUint64(&currentInsertingBlockHeight, curblkhei)
	}
	// 状态重置
	go func() {
		time.Sleep(time.Second * 5)
		atomic.StoreUint64(&currentInsertingBlockHeight, 0)
	}()
	// 解析完整区块
	block, _, e1 := blocks.ParseBlock(msgbody, 0)
	if e1 != nil {
		return // error end
	}
	//fmt.Println("GetBlockDiscover", 2)
	blockhxstr := string(block.Hash())
	if p2p.CheckKnowledge("block", blockhxstr) {
		return //
	}
	//fmt.Println("GetBlockDiscover", 3)
	p2p.AddKnowledge("block", blockhxstr)
	peer.AddKnowledge("block", blockhxstr)
	// insert block
	//fmt.Println("GetBlockDiscover", 4)
	//fmt.Println("get: MrklRoot", block.GetMrklRoot().ToHex(), hex.EncodeToString(msgbody), msgbody)
	// check lastest block
	lastest, e4 := blockchain.State().ReadLastestBlockHeadAndMeta()
	if e4 != nil {
		return // errro end
	}
	mylastblockheight := lastest.GetHeight()
	if block.GetHeight() > mylastblockheight+1 {
		fmt.Printf("need height %d but got %d, sync the new blocks ...\n", mylastblockheight+1, block.GetHeight())
		// check hash fork
		msgParseSendRequestBlockHashList(peer, 8, mylastblockheight)
		return
	} else if block.GetHeight() < mylastblockheight+1 {
		peer.Disconnect()
		//fmt.Printf("need height %d but got %d, ignore.\n", mylastblockheight+1, block.GetHeight())
		return // error block height
	}
	// note
	fmt.Printf("discover new block height: %d, txs: %d, hash: %s, time: %s, try to inserting ... ",
		block.GetHeight(), block.GetCustomerTransactionCount(), block.Hash().ToHex(),
		time.Unix(int64(block.GetTimestamp()), 0).Format(time_format_layout))
	// do insert
	err := blockchain.InsertBlock(block, "discover")
	if err == nil {
		fmt.Printf("ok.\n")
	} else {
		//fmt.Printf("\n\n---------------\n-- %s\n---------------\n\n", err.Error())
		fmt.Println("\n" + err.Error() + "\n")
	}
}

func GetTransactionSubmit(p2p interfaces.P2PManager, pool interfaces.TxPool, peer interfaces.P2PMsgPeer, msgbody []byte) {
	if pool == nil {
		return // no pool
	}
	tx, _, e1 := transactions.ParseTransaction(msgbody, 0)
	if e1 != nil {
		return // error end
	}
	txhxstr := string(tx.HashWithFee())
	if p2p.CheckKnowledge("tx", txhxstr) {
		return //
	}
	peer.AddKnowledge("tx", txhxstr)

	//fmt.Println("GetTransactionSubmit to Add Txpool:", tx.Hash().ToHex())
	pool.AddTx(tx)
}
