package node

import (
	"github.com/hacash/node/handler"
	"github.com/hacash/node/p2p"
)

// OnConnected
func (hn *HacashNode) OnConnected(msghandler p2p.MsgCommunicator, peer p2p.MsgPeer) {
	// req status and hand shake
	peer.SendDataMsg(handler.MsgTypeRequestStatus, nil)

}

// OnDisconnected
func (hn *HacashNode) OnDisconnected(peer p2p.MsgPeer) {

}

// OnConnected
func (hn *HacashNode) OnMsgData(msghandler p2p.MsgCommunicator, peer p2p.MsgPeer, msgty uint16, msgbody []byte) {

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

}
