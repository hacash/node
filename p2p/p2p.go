package p2p

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"sync"
)

type P2PManagerConfig struct {
	TcpListenPort int
	UdpListenPort int
}

func NewP2PManagerConfig() *P2PManagerConfig {
	cnf := &P2PManagerConfig{
		TcpListenPort: 3337,
		UdpListenPort: 3336,
	}
	return cnf
}

type P2PManager struct {
	config *P2PManagerConfig

	peerManager *PeerManager

	selfPeerName            string
	selfPeerId              []byte
	selfPublicTCPListenAddr net.Addr

	//selfRemoteAddr net.Addr

	changeStatusLock sync.Mutex
}

func NewP2PManager(cnf *P2PManagerConfig, pmcnf *PeerManagerConfig) (*P2PManager, error) {

	p2p := &P2PManager{
		config:                  cnf,
		selfPublicTCPListenAddr: nil,
	}

	// test
	p2p.selfPeerId, _ = hex.DecodeString("12a1633cafcc01ebfb6d78e39f687a1f0995c62fc95f51ead10a02ee0be551b5")
	rand.Read(p2p.selfPeerId) // test
	p2p.selfPeerName = "hcx_test_node"

	// pmcnf := &PeerManagerConfig{}
	p2p.peerManager = NewPeerManager(pmcnf, p2p)

	return p2p, nil
}

func (p2p *P2PManager) Start() {

	go p2p.startListenTCP()

	go p2p.startListenUDP()

	p2p.peerManager.Start()

}
