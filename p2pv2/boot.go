package p2pv2

import (
	"fmt"
)

// Static Boot Nodes
func (p *P2P) tryConnectToStaticBootNodes() {
	if len(p.Config.StaticHnodeAddrs) == 0 {
		fmt.Println("[P2P] !!!Warning!!! Not configurate or find any boot nodes! p2p may cannot synchronize data.")
		return
	}
	// try
	for _, tcp := range p.Config.StaticHnodeAddrs {
		_, err := p.ConnectNodeInitiative(tcp)
		if err != nil {
			fmt.Println("[P2P Error] Try notifyConnect To Static Boot Node:", err)
		}
	}
}
