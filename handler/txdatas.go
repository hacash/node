package handler

import (
	"bytes"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
	"time"
)

func GetRequestTxDatas(txpool interfaces.TxPool, peer interfaces.MsgPeer) {

	txs := txpool.CopyTxsOrderByFeePurity(5, 0, 1024*1024*50)

	tsdatas := bytes.NewBuffer([]byte{})

	for i := 0; i < len(txs); i++ {
		txdts, e := txs[i].Serialize()
		if e != nil {
			continue
		}
		// fmt.Println( "SendTxDatas to", peer.Describe(), ":", txs[i].Hash().ToHex() )
		tsdatas.Write(txdts)
		// send data segment
		if i%15 == 0 {
			peer.SendDataMsg(MsgTypeTxDatas, tsdatas.Bytes())
			tsdatas.Truncate(0) // clean
			time.Sleep(time.Millisecond * 250)
		}
	}

}

func GetTxDatas(txpool interfaces.TxPool, msgbody []byte) {

	txpool.PauseEventSubscribe()
	defer txpool.RenewalEventSubscribe()

	seek := uint32(0)

	for {
		if seek >= uint32(len(msgbody)) {
			break
		}
		tx, sk, e := transactions.ParseTransaction(msgbody, seek)
		if e != nil {
			break
		}
		//fmt.Println( "GetTxDatas:", tx.Hash().ToHex() )
		// add to pool
		txpool.AddTx(tx)
		seek = sk
	}

}
