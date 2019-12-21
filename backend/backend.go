package backend

import (
	"github.com/hacash/core/interfaces"
	"github.com/hacash/mint/blockchain"
	"github.com/hacash/node/p2p"
	"sync"
)

type Backend struct {
	config *BackendConfig

	p2p        *p2p.P2PManager
	msghandler interfaces.MsgCommunicator

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

	// insert block success
	bc.SubscribeValidatedBlockOnInsert(backend.discoverNewBlockSuccessCh)

	// return
	return backend, nil
}

// Start
func (hn *Backend) Start() {

	hn.p2p.Start()

	go hn.loop()

}

//
func (hn *Backend) GetBlockChain() interfaces.BlockChain {
	return hn.blockchain
}

// set
func (hn *Backend) SetTxPool(pool interfaces.TxPool) {
	hn.txpool = pool
	// add tx feed
	pool.SubscribeOnAddTxSuccess(hn.addTxToPoolSuccessCh)

}

/*
1 0 0 0 0 1 0 93 254 1 219 0 0 0 7 119 144 186 47 205 234 239 74 66 153 217 182 103 19 91 172 87 124 226 4 222 232 56 143 27 151 247 230 61 219 168 184 220 232 27 37 120 229 222 140 118 239 175 152 156 98 181 249 21 5 253 57 173 235 205 62 227 98 250 209 0 0 0 1 0 0 0 0 255 255 255 254 0 0 0 0 230 60 51 167 150 179 3 44 230 184 86 246 143 204 240 102 8 217 237 24 248 1 1 32 32 32 32 32 32 32 32 32 32 32 0 0 0 0 1 0
1 0 0 0 0 1 0 93 254 1 219 0 0 0 7 119 144 186 47 205 234 239 74 66 153 217 182 103 19 91 172 87 124 226 4 222 232 56 143 27 151 247 230 61 219 168 184 220 232 27 37 120 229 222 140 118 239 175 152 156 98 181 249 21 5 253 57 173 235 205 62 227 98 250 209 0 0 0 1 0 0 0 0 255 255 255 254 0 0 0 0 230 60 51 167 150 179 3 44 230 184 86 246 143 204 240 102 8 217 237 24 248 1 1 32 32 32 32 32 32 32 32 32 32 32 0 0 0 0 1 0


010000000001005dfe0346000000077790ba2fcdeaef4a4299d9b667135bac577ce204dee8388f1b97f7e63ddba8b8dce81b2578e5de8c76efaf989c62b5f91505fd39adebcd3ee362fad10000000100000000fffffffe00000000e63c33a796b3032ce6b856f68fccf06608d9ed18f801012020202020202020202020000000000100
010000000001005dfe0346000000077790ba2fcdeaef4a4299d9b667135bac577ce204dee8388f1b97f7e63ddba8b8dce81b2578e5de8c76efaf989c62b5f91505fd39adebcd3ee362fad10000000100000000fffffffe00000000e63c33a796b3032ce6b856f68fccf06608d9ed18f801012020202020202020202020000000000100


*/
