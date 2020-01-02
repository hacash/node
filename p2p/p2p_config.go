package p2p

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/hacash/core/sys"
	"net"
	"os"
	"path"
)

type P2PManagerConfig struct {
	Datadir       string
	TCPListenPort int
	UDPListenPort int
	Name          string
	ID            []byte

	StaticHnodeAddrs []*net.TCPAddr // IP:port ...

	lookupConnectMaxLen int
}

func NewP2PManagerConfigByID(id []byte) *P2PManagerConfig {
	if len(id) != 16 {
		panic("P2PManagerConfig ID len must be 16.")
	}
	cnf := &P2PManagerConfig{
		Name:                "hnode_" + string([]byte(hex.EncodeToString(id))[0:14]),
		ID:                  id,
		TCPListenPort:       3337,
		UDPListenPort:       3336,
		lookupConnectMaxLen: 128,
		StaticHnodeAddrs:    make([]*net.TCPAddr, 0),
	}
	return cnf
}

// new p2p
func NewP2PManagerConfig(cnffile *sys.Inicnf) *P2PManagerConfig {
	data_dir := path.Join(cnffile.MustDataDir(), "node")
	p2pid := readIDFromDisk(data_dir)
	if p2pid == nil {
		p2pid = make([]byte, 16)
		rand.Read(p2pid) // random id
		saveIDToDisk(data_dir, p2pid)
	}
	// create cnf
	cnf := NewP2PManagerConfigByID(p2pid)
	cnf.Datadir = data_dir
	ini_section_p2p := cnffile.Section("p2p")
	// port
	ini_listen_port := ini_section_p2p.Key("listen_port").MustInt(3337)
	ini_udp_port := ini_section_p2p.Key("udp_port").MustInt(0)
	if ini_udp_port == 0 {
		ini_udp_port = ini_listen_port - 1
	}
	cnf.TCPListenPort = ini_listen_port
	cnf.UDPListenPort = ini_udp_port
	// name
	p2pname := ini_section_p2p.Key("name").MustString("")
	if p2pname != "" {
		cnf.Name = p2pname
	}
	// static node url bootnodes
	boot_nodes := cnffile.StringValueList("p2p", "boot_nodes")
	for _, one := range boot_nodes {
		if tcp, err := net.ResolveTCPAddr("tcp", one); err == nil {
			cnf.StaticHnodeAddrs = append(cnf.StaticHnodeAddrs, tcp)
		}else{
			fmt.Println("[P2P Config Error]", err.Error())
			os.Exit(0)
		}
	}
	// ok
	return cnf
}

func saveIDToDisk(data_dir string, p2pid []byte) {
	os.MkdirAll(data_dir, os.ModePerm)
	idfile, e1 := os.OpenFile(path.Join(data_dir, "id.json"), os.O_RDWR|os.O_CREATE, 0777)
	if e1 == nil {
		idjsonobj := struct {
			ID string
		}{}
		idjsonobj.ID = hex.EncodeToString(p2pid)
		jsonbts, e := json.Marshal(idjsonobj)
		if e == nil {
			idfile.Write(jsonbts)
		}
		idfile.Close()
	}
}

func readIDFromDisk(data_dir string) []byte {
	var p2pid []byte = nil
	idjsonobj := struct {
		ID string `json:"ID"`
	}{}
	idfile, e1 := os.OpenFile(path.Join(data_dir, "id.json"), os.O_RDWR|os.O_CREATE, 0777)
	if e1 == nil {
		idjsonstr := make([]byte, 16*4)
		rn, e := idfile.Read(idjsonstr)
		if e == nil {
			e := json.Unmarshal(idjsonstr[0:rn], &idjsonobj)
			if e == nil {
				id, e := hex.DecodeString(idjsonobj.ID)
				if e == nil && len(id) >= 16 {
					p2pid = id[0:16]
				}
			}
		}
		idfile.Close()
	}
	// return
	return p2pid
}
