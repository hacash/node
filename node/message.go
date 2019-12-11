package node

import (
	"github.com/hacash/node/handle"
	"github.com/hacash/node/p2p"
)

// OnConnected
func (hn *HacashNode) OnConnected(msghandler p2p.MsgCommunicator, peer *p2p.Peer) {

}

// OnDisconnected
func (hn *HacashNode) OnDisconnected(peer *p2p.Peer) {

}

// OnConnected
func (hn *HacashNode) OnMsgData(msghandler p2p.MsgCommunicator, peer *p2p.Peer, msgty uint16, msgbody []byte) {

	if msgty == handle.MsgTypeRequestStatus {

		return
	}

	if msgty == handle.MsgTypeStatus {

		return
	}

}
