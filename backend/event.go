package backend

import (
	"github.com/hacash/core/interfaces"
	"github.com/hacash/node/handler"
)

func (hn *Backend) broadcastNewBlockDiscover(block interfaces.Block) {
	hn.msgFlowLock.Lock()
	defer hn.msgFlowLock.Unlock()

	//fmt.Println("broadcastNewBlockDiscover:", 1)
	if hn.msghandler == nil {
		return
	}
	//fmt.Println("broadcastNewBlockDiscover:", 2)
	blkhxstr := string(block.Hash())
	//fmt.Println("broadcastNewBlockDiscover:", 3)
	hn.p2p.AddKnowledge("block", blkhxstr)
	// send
	blockdata, e1 := block.Serialize()
	if e1 != nil {
		return
	}
	//fmt.Println("broadcastNewBlockDiscover:", 4)
	// send to all
	//fmt.Println("send: MrklRoot", block.GetMrklRoot().ToHex() , hex.EncodeToString(blockdata), blockdata)
	hn.msghandler.BroadcastDataMessageToUnawarePeers(handler.MsgTypeDiscoverNewBlock, blockdata, "block", blkhxstr)
}

func (hn *Backend) broadcastNewTxSubmit(tx interfaces.Transaction) {
	hn.msgFlowLock.Lock()
	defer hn.msgFlowLock.Unlock()

	if hn.msghandler == nil {
		return
	}
	txhxstr := string(tx.Hash())
	hn.p2p.AddKnowledge("tx", txhxstr)
	// send
	txdata, e1 := tx.Serialize()
	if e1 != nil {
		return
	}
	// send to all
	hn.msghandler.BroadcastDataMessageToUnawarePeers(handler.MsgTypeSubmitTransaction, txdata, "tx", txhxstr)
}
