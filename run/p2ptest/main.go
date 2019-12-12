package main

import (
	"fmt"
	"github.com/hacash/node/p2p"
	"github.com/hacash/node/p2p_other"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"
)

/**

go build -o test1 run/p2ptest/main.go && ./test1
go build -o test2 run/p2ptest/main.go && ./test2
go build -o test3 run/p2ptest/main.go && ./test3


*/

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	fmt.Println("hacash p2p_other test")

	//startServiceUDP()
	//startClientUDP()

	//startServiceTcp()
	//startClientTcp()

	//p2pcnf := p2p_other.NewP2PManagerConfigByID()
	//pm, _ := p2p_other.NewP2PManager(p2pcnf)
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

	//go start_netpass_server()
	//go start_netpass_client(9982)
	//go start_netpass_client(9983)

	//startnode(7001)
	//startnode(7003)
	startnode(7005)

	s := <-c
	fmt.Println("Got signal:", s)

}

func startnode(port int) {

	p2pcnf := p2p.NewP2PManagerConfigByID()
	p2pcnf.TCPListenPort = port
	p2pcnf.UDPListenPort = p2pcnf.TCPListenPort + 1
	pmcnf := p2p.NewEmptyPeerManagerConfig()
	pm, _ := p2p.NewP2PManager(p2pcnf, pmcnf)
	pm.Start()
	// connect test
	if port != 7001 {
		//rmtaddr, _ := net.ResolveTCPAddr("tcp", "182.92.163.225:7001")
		rmtaddr, _ := net.ResolveTCPAddr("tcp", "39.96.212.167:7001")
		//rmtaddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:7001")
		go pm.TryConnectToPeer(nil, rmtaddr)
		//go pm.TryConnectToNode(nil, &net.TCPAddr{net.IPv4zero, 7001, ""})
	}

}

var tag string

const HAND_SHAKE_MSG = "im nat pass msg"

func start_netpass_client(port int) {
	// 当前进程标记字符串,便于显示
	tag = os.Args[1]
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: port} // 注意端口必须固定
	dstAddr := &net.UDPAddr{IP: net.ParseIP("182.92.163.225"), Port: 9981}
	//dstAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9981}
	conn, err := net.ListenUDP("udp", srcAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	if _, err = conn.WriteTo([]byte("hello, I'm new peer:"+tag), dstAddr); err != nil {
		log.Panic(err)
	}
	data := make([]byte, 1024)
	n, remoteAddr, err := conn.ReadFromUDP(data)
	if err != nil {
		fmt.Printf("error during read: %s", err)
	}
	//conn.Close()
	anotherPeer, _ := net.ResolveUDPAddr("udp", string(data[:n]))
	fmt.Printf("local:%s server:%s another:%s\n", srcAddr, remoteAddr, anotherPeer.String())

	// 开始打洞
	//bidirectionHole(srcAddr, anotherPeer)
	bidirectionHole_v2(conn, anotherPeer)
}

func bidirectionHole_v2(conn *net.UDPConn, anotherPeer *net.UDPAddr) {
	// 向另一个peer发送一条udp消息(对方peer的nat设备会丢弃该消息,非法来源),用意是在自身的nat设备打开一条可进入的通道,这样对方peer就可以发过来udp消息
	if _, err := conn.WriteToUDP([]byte(HAND_SHAKE_MSG), anotherPeer); err != nil {
		log.Println("send handshake:", err)
	}
	go func() {
		for {
			time.Sleep(5 * time.Second)
			fmt.Println("try WriteToUDP ...")
			if _, err := conn.WriteToUDP([]byte("from ["+tag+"]"), anotherPeer); err != nil {
				log.Println("send msg fail", err)
			}
		}
	}()
	for {
		data := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(data)
		if err != nil {
			log.Printf("error during read: %s\n", err)
		} else {
			log.Printf("收到数据:%s\n", data[:n])
		}
	}
}

func bidirectionHole(srcAddr *net.UDPAddr, anotherAddr *net.UDPAddr) {
	conn, err := net.DialUDP("udp", srcAddr, anotherAddr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	// 向另一个peer发送一条udp消息(对方peer的nat设备会丢弃该消息,非法来源),用意是在自身的nat设备打开一条可进入的通道,这样对方peer就可以发过来udp消息
	if _, err = conn.Write([]byte(HAND_SHAKE_MSG)); err != nil {
		log.Println("send handshake:", err)
	}
	go func() {
		for {
			time.Sleep(5 * time.Second)
			if _, err = conn.Write([]byte("from [" + tag + "]")); err != nil {
				log.Println("send msg fail", err)
			}
		}
	}()
	for {
		data := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(data)
		if err != nil {
			log.Printf("error during read: %s\n", err)
		} else {
			log.Printf("收到数据:%s\n", data[:n])
		}
	}
}
func start_netpass_server() {

	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 9981})
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Printf("本地地址: <%s> \n", listener.LocalAddr().String())
	peers := make([]net.UDPAddr, 0, 2)
	data := make([]byte, 1024)
	for {
		n, remoteAddr, err := listener.ReadFromUDP(data)
		if err != nil {
			fmt.Printf("error during read: %s", err)
		}
		log.Printf("<%s> %s\n", remoteAddr.String(), data[:n])
		peers = append(peers, *remoteAddr)
		if len(peers) == 2 {
			log.Printf("进行UDP打洞,建立 %s <--> %s 的连接\n", peers[0].String(), peers[1].String())
			listener.WriteToUDP([]byte(peers[1].String()), &peers[0])
			listener.WriteToUDP([]byte(peers[0].String()), &peers[1])
			time.Sleep(time.Second * 188)
			log.Println("中转服务器退出,仍不影响peers间通信")
			return
		}
	}

}

func startnode_old(port int) {

	p2pcnf := p2p_other.NewP2PManagerConfig()
	p2pcnf.TcpListenPort = port
	p2pcnf.UdpListenPort = p2pcnf.TcpListenPort + 1
	pmcnf := p2p_other.NewPeerManagerConfig()
	pm, _ := p2p_other.NewP2PManager(p2pcnf, pmcnf)
	pm.Start()
	// connect test
	if port != 7001 {
		rmtaddr, _ := net.ResolveTCPAddr("tcp", "182.92.163.225:7001")
		//rmtaddr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:7001")
		go pm.TryConnectToNode(nil, rmtaddr)
		//go pm.TryConnectToNode(nil, &net.TCPAddr{net.IPv4zero, 7001, ""})
	}

}

func test_tcp_udp() {

	// UDP call to out of NAT
	socket, err := net.DialUDP("udp",
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
	socket, err := net.DialUDP("udp", nil, &net.UDPAddr{
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
