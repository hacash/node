package handler

import (
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
	"github.com/hacash/node/p2p"
	"time"
)

const (
	time_format_layout = "01/02 15:04:05"
)

func GetBlockDiscover(p2p *p2p.P2PManager, blockchain interfaces.BlockChain, peer interfaces.MsgPeer, msgbody []byte) {
	//fmt.Println("GetBlockDiscover", 1)
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
	fmt.Printf("discover new block height: %d, txs: %d, hash: %s, time: %s, try to inserting ... ",
		block.GetHeight(), block.GetTransactionCount()-1, block.Hash().ToHex(), time.Unix(int64(block.GetTimestamp()), 0).Format(time_format_layout))
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
		fmt.Printf("need height %d but got %d, ignore.\n", mylastblockheight+1, block.GetHeight())
		return // error block height
	}
	// do insert
	err := blockchain.InsertBlock(block, "discover")
	if err == nil {
		fmt.Printf("ok.\n")
	} else {
		fmt.Printf("\n\n---------------\n-- %s\n---------------\n\n", err.Error())
	}
}

func GetTransactionSubmit(p2p *p2p.P2PManager, pool interfaces.TxPool, peer interfaces.MsgPeer, msgbody []byte) {
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
