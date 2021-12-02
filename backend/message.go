package backend

import (
	"github.com/hacash/core/interfaces"
	"github.com/hacash/node/handler"
	"time"
)

// OnConnected
func (hn *Backend) OnConnected(msghandler interfaces.P2PMsgCommunicator, peer interfaces.P2PMsgPeer) {

	//fmt.Println("-8-8-8-8-8-8-8-8-88-************ (hn *Backend) OnConnected: ", peer.Describe())

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
func (hn *Backend) OnDisconnected(peer interfaces.P2PMsgPeer) {

}

// OnConnected
func (hn *Backend) OnMsgData(cmtr interfaces.P2PMsgCommunicator, peer interfaces.P2PMsgPeer, msgty uint16, msgbody []byte) {

	//fmt.Println("OnMsgData", peer.Describe(), msgty, msgbody)

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
		handler.GetBlocksData(hn.p2p, cmtr, hn.blockchain, peer, msgbody)
		return
	}

	if msgty == handler.MsgTypeSubmitTransaction {
		handler.GetTransactionSubmit(hn.p2p, hn.txpool, peer, msgbody)
		return
	}

	if msgty == handler.MsgTypeDiscoverNewBlock {
		//fmt.Println("msgty == handler.MsgTypeDiscoverNewBlock:", peer.Describe())
		handler.GetBlockDiscover(hn.p2p, hn.msghandler, hn.blockchain, peer, msgbody)
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
