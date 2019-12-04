package p2p

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
)

func (p2p *P2PManager) startListenUDP() {

	udpconn, err := net.ListenUDP("udp", &net.UDPAddr{net.IPv4zero, p2p.config.UDPListenPort, ""})
	if err != nil {
		fmt.Println("startListenUDP error:", err)
		os.Exit(1)
	}

	for {
		data := make([]byte, 1025)
		rn, rmtaddr, err := udpconn.ReadFromUDP(data)
		//fmt.Println("udpconn.ReadFromUDP(data)", rmtaddr.String(), data[0:rn])
		if err != nil {
			continue
		}
		if rn > 1024 || rn < 2 {
			continue
		}
		msgtype := binary.BigEndian.Uint16(data[0:2])
		msgbody := data[2:rn]

		go p2p.handleUDPMessage(rmtaddr, msgtype, msgbody)
	}
}
