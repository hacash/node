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
				// 此处只广播挖矿产生的新区块，discover 收到的区块将在 discover 线程当时自动广播
				go h.broadcastNewBlockDiscover(block)
			}

		}
	}

}
