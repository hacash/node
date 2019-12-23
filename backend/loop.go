package backend

func (h *Backend) loop() {

	for {
		select {
		case tx := <-h.addTxToPoolSuccessCh:
			go h.broadcastNewTxSubmit(tx)

		case block := <-h.discoverNewBlockSuccessCh:
			//fmt.Println("block := <- h.discoverNewBlockSuccessCh:")
			blkmark := block.OriginMark()
			if blkmark == "discover" || blkmark == "mining" {
				go h.broadcastNewBlockDiscover(block)
			}

		}
	}

}
