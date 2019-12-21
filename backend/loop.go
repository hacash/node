package backend

func (h *Backend) loop() {

	for {
		select {
		case tx := <-h.addTxToPoolSuccessCh:
			go h.broadcastNewTxSubmit(tx)

		case block := <-h.discoverNewBlockSuccessCh:
			//fmt.Println("block := <- h.discoverNewBlockSuccessCh:")
			go h.broadcastNewBlockDiscover(block)

		}
	}

}
