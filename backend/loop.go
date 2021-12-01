package backend

import "github.com/hacash/core/interfacev2"

func (h *Backend) loop() {

	for {
		select {
		case tx := <-h.addTxToPoolSuccessCh:
			go h.broadcastNewTxSubmit(tx.(interfacev2.Transaction))

		case block := <-h.discoverNewBlockSuccessCh:
			//fmt.Println("block := <- h.discoverNewBlockSuccessCh:")
			blkmark := block.OriginMark()
			if blkmark == "discover" || blkmark == "mining" {
				go h.broadcastNewBlockDiscover(block.(interfacev2.Block))
			}

		}
	}

}
