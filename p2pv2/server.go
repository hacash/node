package p2pv2

import (
	"encoding/hex"
	"fmt"
	"net"
	"time"
)

func (p *P2P) listen(port int) error {

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{net.IPv4zero, port, ""})
	//laddr := net.TCPAddr{net.IPv4zero, p2p_other.config.TCPListenPort, ""}
	//listener, err := reuseport.Listen("tcp", laddr.String())
	if err != nil {
		return err
	}

	fmt.Printf("[P2P] Start node %s id:%s listen port %d.\n", p.Config.Name, hex.EncodeToString(p.Config.ID), port)

	go func() {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println(err)
				break
			}
			//fmt.Println(conn.RemoteAddr())
			// Perform handshake
			e1 := doTcpMsgHandshakeSignalIfErrorClose(conn, time.Second*10)
			if e1 != nil {
				continue
			}
			// Processing messages
			go p.handleNewConn(conn, nil)
		}
	}()

	return nil

}
