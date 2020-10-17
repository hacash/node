package p2pv2

import (
	"encoding/hex"
	"github.com/hacash/core/interfaces"
	"net"
	"sync"
)

type P2P struct {
	Config *P2PConfig

	msgHandler interfaces.P2PMsgDataHandler

	BackboneNodeTable    []PeerID
	OrdinaryNodeTable    []PeerID
	UnfamiliarNodesTable []PeerID

	AllNodes map[string]*Peer // map[PeerID] // 全部节点池

	TemporaryConns map[uint64]net.Conn // 临时连接池

	PeerChangeMux sync.Mutex

	// my peer
	peerSelf           *Peer
	MyselfIsPublicPeer bool // 我自己是不是公网节点
}

func NewP2P(cnf *P2PConfig) *P2P {

	p2pobj := &P2P{
		Config:               cnf,
		BackboneNodeTable:    []PeerID{},
		OrdinaryNodeTable:    []PeerID{},
		UnfamiliarNodesTable: []PeerID{},
		AllNodes:             make(map[string]*Peer),
		TemporaryConns:       make(map[uint64]net.Conn),
		PeerChangeMux:        sync.Mutex{},
		msgHandler:           nil,
		MyselfIsPublicPeer:   false,
	}

	var peerSelf = NewEmptyPeer(p2pobj, p2pobj.msgHandler)
	peerSelf.ID = cnf.ID
	if len(cnf.Name) == 0 {
		cnf.Name = "hn_" + string([]byte(hex.EncodeToString(cnf.ID))[0:10])
	}
	peerSelf.Name = cnf.Name

	// set myself peer
	p2pobj.peerSelf = peerSelf

	return p2pobj
}

func (p *P2P) doStart() {

	go p.loop()
	go p.listen(p.Config.TCPListenPort)

	go p.tryConnectToStaticBootNodes()

	go p.continuePingToKeepTcpLinkAlive()

}
