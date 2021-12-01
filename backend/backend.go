package backend

import (
	"fmt"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint/blockchain"
	"github.com/hacash/mint/blockchainv3"
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

	var e error = nil

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
	var blockchainObj interfaces.BlockChain = nil
	if config.UseBlockChainV3 {
		// use v3
		bccnf := blockchainv3.NewBlockChainConfig(config.cnffile)
		blockchainObj, e = blockchainv3.NewBlockChain(bccnf)
	} else {
		// use v2
		bccnf := blockchain.NewBlockChainConfig(config.cnffile)
		blockchainObj, e = blockchain.NewBlockChain(bccnf)
	}
	if e != nil {
		fmt.Println("blockchain.NewBlockChain Error", e)
		return nil, e
	}
	backend.blockchain = blockchainObj

	// insert block success
	blockchainObj.GetChainEngineKernel().SubscribeValidatedBlockOnInsert(backend.discoverNewBlockSuccessCh)

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
