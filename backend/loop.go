package backend

func (h *Backend) loop() {

	for {
		select {
		case tx := <-h.addTxToPoolSuccessCh:
			go h.broadcastNewTxSubmit(tx)

		case block := <-h.discoverNewBlockSuccessCh:
			//fmt.Println("block := <- h.discoverNewBlockSuccessCh:")
			blkmark := block.OriginMark()
			//if blkmark == "discover" || blkmark == "mining" {
			if blkmark == "mining" {
				// Only the new blocks generated by mining are broadcast here. The blocks received by the discover thread will be broadcast automatically at that time
				go h.broadcastNewBlockDiscover(block)
			}

		}
	}

}
