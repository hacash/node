package node

import (
	"github.com/hacash/core/inicnf"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/sys"
	"github.com/hacash/mint/blockchain"
	"github.com/hacash/node/p2p"
)

type HacashNode struct {
	config *HacashNodeConfig

	p2p *p2p.P2PManager

	blockchain interfaces.BlockChain
}

func NewHacashNodeByIniCnf(cnffile *inicnf.File) *HacashNode {
	cnf := NewHacashNodeConfig()

	data_dir := cnffile.Section("").Key("data_dir").String()
	cnf.Datadir = sys.CnfMustDataDir(data_dir)

	hacashnode := newHacashNode(cnf)

	// p2p
	p2pmng, err := p2p.NewP2PManagerByIniCnf(cnffile)
	if err != nil {
		panic(err)
	}
	hacashnode.p2p = p2pmng

	// blockchain
	bc, err2 := blockchain.NewBlockChainByIniCnf(cnffile)
	if err2 != nil {
		panic(err2)
	}
	hacashnode.blockchain = bc

	return hacashnode
}

func newHacashNode(cnf *HacashNodeConfig) *HacashNode {

	hacash := &HacashNode{
		config: cnf,
	}

	return hacash
}

// Start
func (hn *HacashNode) Start() {

	hn.p2p.Start()

	go hn.loop()

}
