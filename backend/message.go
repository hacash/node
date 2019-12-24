package backend

import (
	"github.com/hacash/core/interfaces"
	"github.com/hacash/node/handler"
	"time"
)

// OnConnected
func (hn *Backend) OnConnected(msghandler interfaces.MsgCommunicator, peer interfaces.MsgPeer) {
	if hn.msghandler == nil {
		hn.msghandler = msghandler
		// download txs from pool
		go func() {
			time.Sleep(time.Second)
			peer.SendDataMsg(handler.MsgTypeRequestTxDatas, nil)
		}()
	}
	// req status and hand shake
	peer.SendDataMsg(handler.MsgTypeRequestStatus, nil)
}

// OnDisconnected
func (hn *Backend) OnDisconnected(peer interfaces.MsgPeer) {

}

// OnConnected
func (hn *Backend) OnMsgData(msghandler interfaces.MsgCommunicator, peer interfaces.MsgPeer, msgty uint16, msgbody []byte) {

	// fmt.Println("OnMsgData", peer.Describe(), msgty, msgbody)

	if msgty == handler.MsgTypeRequestStatus {
		handler.SendStatusToPeer(hn.blockchain, peer)
		return
	}

	if msgty == handler.MsgTypeStatus {
		handler.GetStatus(hn.blockchain, peer, msgbody)
		return
	}

	if msgty == handler.MsgTypeRequestBlockHashList {
		handler.SendBlockHashList(hn.blockchain, peer, msgbody)
		return
	}

	if msgty == handler.MsgTypeBlockHashList {
		handler.GetBlockHashList(hn.blockchain, peer, msgbody)
		return
	}

	if msgty == handler.MsgTypeRequestBlocks {
		handler.SendBlocksData(hn.blockchain, peer, msgbody)
		return
	}

	if msgty == handler.MsgTypeBlocks {
		handler.GetBlocksData(hn.blockchain, peer, msgbody)
		return
	}

	if msgty == handler.MsgTypeSubmitTransaction {
		handler.GetTransactionSubmit(hn.p2p, hn.txpool, peer, msgbody)
		return
	}

	if msgty == handler.MsgTypeDiscoverNewBlock {
		//fmt.Println("msgty == handler.MsgTypeDiscoverNewBlock:", peer.Describe())
		handler.GetBlockDiscover(hn.p2p, hn.blockchain, peer, msgbody)
		return
	}

	if msgty == handler.MsgTypeRequestTxDatas {
		handler.GetRequestTxDatas(hn.txpool, peer)
		return
	}

	if msgty == handler.MsgTypeTxDatas {
		handler.GetTxDatas(hn.txpool, msgbody)
		return
	}

}
