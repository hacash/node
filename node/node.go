package node

import (
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint/blockchain"
	"github.com/hacash/node/p2p"
	"math/rand"
	"time"
)

type HacashNode struct {
	config *HacashNodeConfig

	p2p *p2p.P2PManager

	blockchain interfaces.BlockChain
}

func NewHacashNode(config *HacashNodeConfig) (*HacashNode, error) {

	hacashnode := &HacashNode{
		config: config,
	}

	// p2p
	p2pcnf := p2p.NewP2PManagerConfig(config.cnffile)
	peercnf := p2p.NewPeerManagerConfig(config.cnffile)
	p2pmng, err := p2p.NewP2PManager(p2pcnf, peercnf)
	if err != nil {
		return nil, err
	}
	hacashnode.p2p = p2pmng
	p2pmng.SetMsgHandler(hacashnode) // handle msg

	// blockchain
	bccnf := blockchain.NewBlockChainConfig(config.cnffile)
	bc, err2 := blockchain.NewBlockChain(bccnf)
	if err2 != nil {
		return nil, err2
	}
	hacashnode.blockchain = bc

	// return
	return hacashnode, nil
}

// Start
func (hn *HacashNode) Start() {

	rand.Seed(time.Now().Unix())

	hn.p2p.Start()

	go hn.loop()

}
