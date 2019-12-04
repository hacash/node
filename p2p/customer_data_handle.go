package p2p

type P2PMsgDataHandler interface {
	OnConnected(*Peer)

	OnMsgData(p *Peer, msgty uint16, msgbody []byte)

	OnDisconnected(*Peer)
}
