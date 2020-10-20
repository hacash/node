package p2pv2

import (
	"encoding/hex"
	"fmt"
	"net"
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
		// 发送握手信号
		e1 := sendTcpMsgHandshakeSignal(conn)
		if e1 != nil {
			conn.Close()
			continue
		}
		// 处理消息
		go p.handleNewConn(conn, nil)
	}

}
