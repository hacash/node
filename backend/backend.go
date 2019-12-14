package backend

import (
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint/blockchain"
	"github.com/hacash/node/p2p"
)

type Backend struct {
	config *BackendConfig

	p2p *p2p.P2PManager

	blockchain interfaces.BlockChain
}

func NewBackend(config *BackendConfig) (*Backend, error) {

	backend := &Backend{
		config: config,
	}

	// p2p
	p2pcnf := p2p.NewP2PManagerConfig(config.cnffile)
	peercnf := p2p.NewPeerManagerConfig(config.cnffile)
	p2pmng, err := p2p.NewP2PManager(p2pcnf, peercnf)
	if err != nil {
		return nil, err
	}
	backend.p2p = p2pmng
	p2pmng.SetMsgHandler(backend) // handle msg

	// blockchain
	bccnf := blockchain.NewBlockChainConfig(config.cnffile)
	bc, err2 := blockchain.NewBlockChain(bccnf)
	if err2 != nil {
		return nil, err2
	}
	backend.blockchain = bc

	// return
	return backend, nil
}

// Start
func (hn *Backend) Start() {

	hn.p2p.Start()

	go hn.loop()

}
