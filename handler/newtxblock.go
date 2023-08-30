package handler

import (
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
	"github.com/hacash/mint/blockchainv3"
	"sync"
	"time"
)

const (
	time_format_layout = "01/02 15:04:05"
)

// Block height currently executing insert
var blockDiscoverMutex sync.Mutex
var transactionSubmitMutex sync.Mutex

func GetBlockDiscover(p2p interfaces.P2PManager, msgcator interfaces.P2PMsgCommunicator, blockchain interfaces.BlockChain, peer interfaces.P2PMsgPeer, msgbody []byte) {
	blockDiscoverMutex.Lock()
	defer blockDiscoverMutex.Unlock()

	//fmt.Println("GetBlockDiscover", 1)
	// Parse the block header first to improve the performance when blocks are received concurrently
	blockhead, _, e0 := blocks.ParseExcludeTransactions(msgbody, 0)
	if e0 != nil {
		return // error end
	}
	//fmt.Println("GetBlockDiscover", 2)
	blockhxstr := string(blockhead.Hash())
	if p2p.CheckKnowledge("block", blockhxstr) {
		return // Block received
	}
	//fmt.Println("GetBlockDiscover", 3)
	// Add a known block, and the broadcast will not be repeated at the event of "successful block insertion"
	p2p.AddKnowledge("block", blockhxstr)
	peer.AddKnowledge("block", blockhxstr)
	// Parse complete block
	block, _, e1 := blocks.ParseBlock(msgbody, 0)
	if e1 != nil {
		return // error end
	}
	// insert block
	//fmt.Println("GetBlockDiscover", 4)
	//fmt.Println("get: MrklRoot", block.GetMrklRoot().ToHex(), hex.EncodeToString(msgbody), msgbody)
	// check lastest block
	lastest, immblk, e4 := blockchain.GetChainEngineKernel().LatestBlock()
	if e4 != nil {
		return // errro end
	}
	mylastblockheight := lastest.GetHeight()
	if block.GetHeight() > mylastblockheight+1 {
		fmt.Printf("need height %d but got %d, sync the new blocks ...\n", mylastblockheight+1, block.GetHeight())
		// check hash fork
		// Sync from my mature block height
		msgParseSendRequestBlocks(peer, immblk.GetHeight()+1)
		//msgParseSendRequestBlockHashList(peer, 8, immblk.GetHeight())
		return
	} else if block.GetHeight() <= mylastblockheight-blockchainv3.ImmatureBlockMaxLength {
		//peer.Disconnect()
		// Block height data lower than or equal to mature confirmation is not accepted
		fmt.Printf("need height %d but got %d, ignore.\n", mylastblockheight+1, block.GetHeight())
		return // error block height
	}
	// note
	fmt.Printf("[%s] discover new block height: %d, txs: %d, hash: %s, time: %s, try to inserting ... ",
		time.Now().Format(time_format_layout),
		block.GetHeight(), block.GetCustomerTransactionCount(), block.Hash().ToHex(),
		time.Unix(int64(block.GetTimestamp()), 0).Format("15:04:05"))
	// do insert
	//testInsertTimeStart := time.Now()
	err := blockchain.GetChainEngineKernel().InsertBlock(block, "discover")
	//testInsertTimeEnd := time.Now()
	//fmt.Println("insert time millisecond: ", testInsertTimeEnd.Sub(testInsertTimeStart ).Milliseconds())
	if err == nil {
		fmt.Printf("ok.\n")
		// After successful insertion, the block will be broadcasted out at the fastest speed regardless of whether it is a forked block or not
		msgcator.BroadcastDataMessageToUnawarePeers(MsgTypeDiscoverNewBlock, msgbody, "block", blockhxstr)
	} else {
		//fmt.Printf("\n\n---------------\n-- %s\n---------------\n\n", err.Error())
		fmt.Println("\n" + err.Error() + "\n")
	}
}

func GetTransactionSubmit(p2p interfaces.P2PManager, pool interfaces.TxPool, peer interfaces.P2PMsgPeer, msgbody []byte) {
	transactionSubmitMutex.Lock()
	defer transactionSubmitMutex.Unlock()

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
	pool.AddTx(tx.(interfaces.Transaction))
}
