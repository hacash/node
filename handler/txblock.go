package handler

import (
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
	"github.com/hacash/node/p2p"
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
	fmt.Printf("discover new block height: %d hash: %s try to inserting... ", block.GetHeight(), block.Hash().ToHex())
	block.SetOriginMark("discover")
	err := blockchain.InsertBlock(block)
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
	txhxstr := string(tx.Hash())
	if p2p.CheckKnowledge("tx", txhxstr) {
		return //
	}
	peer.AddKnowledge("tx", txhxstr)
	pool.AddTx(tx)
}
