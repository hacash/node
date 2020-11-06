package p2pv2

import (
	"encoding/hex"
	"fmt"
	"net"
	"time"
)

func (p *P2P) listen(port int) {

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{net.IPv4zero, port, ""})
	//laddr := net.TCPAddr{net.IPv4zero, p2p_other.config.TCPListenPort, ""}
	//listener, err := reuseport.Listen("tcp", laddr.String())
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	fmt.Printf("[P2P] Start node %s id:%s listen port %d.\n", p.Config.Name, hex.EncodeToString(p.Config.ID), port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			break
		}
		//fmt.Println(conn.RemoteAddr())
		// 执行握手
		e1 := doTcpMsgHandshakeSignalIfErrorClose(conn, time.Second*10)
		if e1 != nil {
			continue
		}
		// 处理消息
		go p.handleNewConn(conn, nil)
	}

}
