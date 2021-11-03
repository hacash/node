package backend

import (
	"fmt"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint/blockchain"
	"github.com/hacash/node/p2pv2"
	"strings"
	"sync"
)

type Backend struct {
	config *BackendConfig

	p2p        interfaces.P2PManager
	msghandler interfaces.P2PMsgCommunicator

	txpool interfaces.TxPool

	addTxToPoolSuccessCh      chan interfaces.Transaction
	discoverNewBlockSuccessCh chan interfaces.Block

	blockchain interfaces.BlockChain

	msgFlowLock sync.Mutex
}

func NewBackend(config *BackendConfig) (*Backend, error) {

	backend := &Backend{
		config:                    config,
		msghandler:                nil,
		addTxToPoolSuccessCh:      make(chan interfaces.Transaction, 5),
		discoverNewBlockSuccessCh: make(chan interfaces.Block, 5),
	}

	// p2p
	p2pcnf := p2pv2.NewP2PConfig(config.cnffile)
	p2pmng := p2pv2.NewP2P(p2pcnf)
	backend.p2p = p2pmng
	p2pmng.SetMsgHandler(backend) // handle msg

	// blockchain
	bccnf := blockchain.NewBlockChainConfig(config.cnffile)
	bc, err2 := blockchain.NewBlockChain(bccnf)
	if err2 != nil {
		fmt.Println("blockchain.NewBlockChain Error", err2)
		return nil, err2
	}
	backend.blockchain = bc

	// insert block success
	bc.SubscribeValidatedBlockOnInsert(backend.discoverNewBlockSuccessCh)

	// return
	return backend, nil
}

// Start
func (hn *Backend) Start() error {

	if hn.blockchain != nil {
		err := hn.blockchain.Start()
		if err != nil {
			return err
		}
	} else {
		err := fmt.Errorf("[Backend] blockchain is nil.")
		return err
	}

	go hn.loop()

	return hn.p2p.Start()

}

//
func (hn *Backend) BlockChain() interfaces.BlockChain {
	return hn.blockchain
}

// set
func (hn *Backend) SetTxPool(pool interfaces.TxPool) {
	hn.txpool = pool
	// add tx feed
	pool.SubscribeOnAddTxSuccess(hn.addTxToPoolSuccessCh)

}

func (hn *Backend) AllPeersDescribe() string {
	if hn.msghandler == nil {
		return "p2p connected: 0"
	}
	pppstrs := ""
	for _, v := range hn.msghandler.GetAllPeers() {
		pppstrs += v.Describe() + ", "
	}
	pppstrs = strings.Trim(pppstrs, ", ")
	str := fmt.Sprintf(
		"p2p connected: %d, peers: %s",
		hn.msghandler.PeerLen(),
		pppstrs,
	)
	return str

}
