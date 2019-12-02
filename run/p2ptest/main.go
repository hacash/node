package main

import (
	"fmt"
	"github.com/hacash/node/p2p"
	"net"
	"os"
	"os/signal"
	"strings"
)

/**

go build -o test1 run/p2ptest/main.go && ./test1
go build -o test2 run/p2ptest/main.go && ./test2
go build -o test3 run/p2ptest/main.go && ./test3


*/

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	fmt.Println("hacash p2p test")

	//startServiceUDP()
	//startClientUDP()

	//startServiceTcp()
	//startClientTcp()

	//p2pcnf := p2p.NewP2PManagerConfig()
	//pm, _ := p2p.NewP2PManager(p2pcnf)
	//
	//pm.Start()
	//
	//addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:8181")
	//pm.ConnectToAddr(addr)

	//listener, err := net.ListenTCP("tcp", &net.TCPAddr{net.IPv4zero, 8182, ""})
	//if err != nil {
	//	fmt.Println("startListenTCP error:", err)
	//	os.Exit(1)
	//}
	//for {
	//	conn, err := listener.Accept()
	//	if err != nil {
	//		fmt.Println(err)
	//		break
	//	}
	//	fmt.Println("conn.LocalAddr", conn.LocalAddr())
	//	conn.Write([]byte("hello client!"))
	//	data := make([]byte, 1024)
	//	rn, _ := conn.Read(data)
	//	fmt.Println("ListenTCP", string(data[:rn]))
	//	conn.Close()
	//	break
	//}

	//tcpconn, err := net.DialTCP("tcp",
	//	&net.TCPAddr{net.IPv4zero, 8181, ""},
	//	&net.TCPAddr{net.IPv4zero, 8182, ""},
	//	)
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(0)
	//}
	//fmt.Println("tcpconn.LocalAddr", tcpconn.LocalAddr())
	//tcpconn.Write([]byte("hello server!"))
	//data := make([]byte, 1024)
	//rn, _ := tcpconn.Read(data)
	//fmt.Println("tcpconn.Read(data)", string(data[:rn]))
	//tcpconn.Close()

	//listener, _ := net.ListenTCP("tcp", &net.TCPAddr{net.IPv4zero, 8181, ""})
	//conn, _ := listener.Accept()
	//conn.Write([]byte("hello client!"))
	//go func() {for {
	//	data := make([]byte, 1024)
	//	rn, _ := conn.Read(data)
	//	fmt.Println(string(data[:rn]))
	//}
	//}()
	//listener.Close()
	//fmt.Println("listener.Close() ok")

	//tcpconn, _ := net.DialTCP("tcp", nil, &net.TCPAddr{net.IPv4zero, 8181, ""})
	//go func() {
	//	for {
	//		tcpconn.Write([]byte("hello server!"))
	//		<- time.Tick(time.Second * 3)
	//	}
	//}()

	//test_tcp_udp()

	//startnode(7001)
	//startnode(7003)
	startnode(7005)

	s := <-c
	fmt.Println("Got signal:", s)

}

func startnode(port int) {

	p2pcnf := p2p.NewP2PManagerConfig()
	p2pcnf.TcpListenPort = port
	p2pcnf.UdpListenPort = p2pcnf.TcpListenPort + 1
	pmcnf := p2p.NewPeerManagerConfig()
	pm, _ := p2p.NewP2PManager(p2pcnf, pmcnf)
	pm.Start()
	// connect test
	if port != 7001 {
		rmtaddr, _ := net.ResolveTCPAddr("tcp", "182.92.163.225:7001")
		go pm.TryConnectToNode(nil, rmtaddr)
		//go pm.TryConnectToNode(nil, &net.TCPAddr{net.IPv4zero, 7001, ""})
	}

}

func test_tcp_udp() {

	// UDP call to out of NAT
	socket, err := net.DialUDP("udp4",
		&net.UDPAddr{net.IPv4zero, 8181, ""},
		&net.UDPAddr{net.IPv4zero, 8182, ""},
	)
	if err != nil {
		fmt.Println("DialUDP error", err)
		os.Exit(1)
	}
	// send data
	n, err := socket.Write([]byte("hello!"))
	fmt.Println(n, err)
	socket.Close()

	fmt.Println("DialUDP socket.Close  ok!")

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{net.IPv4zero, 8181, ""})
	if err != nil {
		fmt.Println("startListen error", err)
		os.Exit(1)
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("conn, err := listener.Accept()", conn.LocalAddr().String(), conn.RemoteAddr().String())
		}
	}()

	fmt.Println("ListenTCP 8181 ok!")

	// UDP call to out of NAT
	socket2, err := net.DialTCP("tcp", nil,
		&net.TCPAddr{net.IPv4zero, 8181, ""},
	)
	if err != nil {
		fmt.Println("DialTCP error", err)
		os.Exit(1)
	}
	// send data
	wn, err := socket2.Write([]byte("hello server!"))
	fmt.Println(wn, err)
	socket.Close()

	fmt.Println("DialTCP socket.Close  ok!")

	// UDP call to out of NAT
	socket3, err := net.DialTCP("tcp", nil,
		&net.TCPAddr{net.IPv4zero, 8181, ""},
	)
	if err != nil {
		fmt.Println("DialTCP error", err)
		os.Exit(1)
	}
	// send data
	wn2, err := socket3.Write([]byte("hello server!"))
	fmt.Println(wn2, err)
	socket.Close()

	fmt.Println("DialTCP socket.Close  ok!")

}

func startClientTcp() {
	conn, err := net.Dial("tcp", "0.0.0.0:8181") // 255.255.255.255:65535
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	RemoteAddr := conn.RemoteAddr()
	fmt.Println("RemoteAddr", RemoteAddr)

	conn.Write([]byte("hello im client."))

	conn.Close()

}

func startServiceTcp() {

	listener, err := net.Listen("tcp", "0.0.0.0:8181")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		RemoteAddr := conn.RemoteAddr()
		fmt.Println("Connect Remote Addr", RemoteAddr)

		defer conn.Close()
		for {
			data := make([]byte, 2048)
			_, err = conn.Read(data)
			if err != nil {
				fmt.Println(err)
				break
			}

			strData := string(data)
			fmt.Println("Received:", strData)

			upper := strings.ToUpper(strData)
			_, err = conn.Write([]byte(upper))
			if err != nil {
				fmt.Println(err)
				break
			}

			fmt.Println("Send:", upper)
		}
	}

}

func startClientUDP() {
	// 创建连接
	socket, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 8181,
	})
	if err != nil {
		fmt.Println("连接失败!", err)
		return
	}
	defer socket.Close()

	// 发送数据
	senddata := []byte("hello server!")
	_, err = socket.Write(senddata)
	if err != nil {
		fmt.Println("发送数据失败!", err)
		return
	}

	// 接收数据
	data := make([]byte, 4096)
	read, remoteAddr, err := socket.ReadFromUDP(data)
	if err != nil {
		fmt.Println("读取数据失败!", err)
		return
	}
	fmt.Println(read, remoteAddr)
	fmt.Printf("%s\n", data)
}

func startServiceUDP() {

	// 创建监听
	socket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 8181,
	})
	if err != nil {
		fmt.Println("监听失败!", err)
		return
	}
	defer socket.Close()

	for {
		// 读取数据
		data := make([]byte, 4096)
		read, remoteAddr, err := socket.ReadFromUDP(data)
		if err != nil {
			fmt.Println("读取数据失败!", err)
			continue
		}
		fmt.Println(read, remoteAddr)
		fmt.Printf("%s\n\n", data)

		// 发送数据
		senddata := []byte("hello client!")
		_, err = socket.WriteToUDP(senddata, remoteAddr)
		if err != nil {
			return
			fmt.Println("发送数据失败!", err)
		}
	}

}
