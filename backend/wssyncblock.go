package backend

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/node/websocket"
	"strconv"
)

// download block data form ws api
func (h *Backend) DownloadBlocksDataFromWebSocketApi(ws_url string, start_height uint64) (uint64, error) {

	// websocket
	// ws_url = "ws://127.0.0.1:3338/websocket"
	wsConn, e2 := websocket.Dial(ws_url, "ws", "http://127.0.0.1/")
	if e2 != nil {
		fmt.Println(e2)
		return 0, e2
	}

	start_block_height := start_height // 1

	datasbuf := bytes.NewBuffer([]byte{})
	tagetdataslength := -1

	rdata := make([]byte, 5000)
	for {
		if tagetdataslength == -1 {
			fmt.Print("download blocks from ["+ws_url+"] start height: ", start_block_height, " ... ")
			wsConn.Write([]byte("getblocks " + strconv.FormatUint(start_block_height, 10)))
		}

		rn, e := wsConn.Read(rdata)
		if e != nil {
			fmt.Println(e)
			return 0, e
		}
		//fmt.Println("rn", rn)
		data := rdata[0:rn]
		if rn == 9 && bytes.Compare(data, []byte("endblocks")) == 0 {
			fmt.Println("got endblocks.")
			break
		}
		datasbuf.Write(data)
		if datasbuf.Len() < 4 {
			fmt.Println("datasbuf.Len() < 4, continue")
			continue
		}
		if tagetdataslength == -1 {
			tagetdataslength = int(binary.BigEndian.Uint32(data[0:4]))
		}
		if datasbuf.Len() == tagetdataslength+4 {
			datas := datasbuf.Bytes()
			fmt.Print("got success, inserting ... ")
			start_block_height, e = newBlocksDataArrive(h.blockchain, datas[4:])
			//fmt.Println("start_block_height", start_block_height)
			if e != nil {
				fmt.Println(e)
				return 0, e
			}
			fmt.Println("OK.")
			tagetdataslength = -1
			datasbuf = bytes.NewBuffer([]byte{})
		}

	}

	fmt.Println("end of download blocks.")

	return start_block_height, nil
}

func newBlocksDataArrive(blockchain interfaces.BlockChain, datas []byte) (uint64, error) {

	start_block_height := uint64(0)

	seek := uint32(0)
	for {
		if int(seek)+1 > len(datas) {
			break
		}
		//fmt.Println(seek, datas[seek:seek + 80])
		newblock, sk, e := blocks.ParseBlock(datas, seek)
		if e != nil {
			fmt.Println(e)
			return 0, e
		}
		//fmt.Println(newblock.GetHeight())
		seek = sk
		// do store
		err := blockchain.InsertBlock(newblock)
		if err != nil {
			return 0, err
		}
		start_block_height = newblock.GetHeight() + 1

		/************* test *************/
		if start_block_height == 5 {
			//return 0, fmt.Errorf("ok 5 blocks")
		}
		/*********** test end **********/
	}
	// ok
	return start_block_height, nil
}
