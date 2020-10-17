package p2pv2

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/hacash/core/sys"
	"net"
	"os"
	"path"
	//"path"
)

type P2PConfig struct {
	Datadir          string
	Name             string
	ID               []byte
	StaticHnodeAddrs []*net.TCPAddr // IP:port ...
	TCPListenPort    int

	// 骨干/超级/公网 节点 连接表大小
	BackboneNodeTableSizeMax int `json:"backbone_node_table_size_max"`
	// 叶子/普通/私网 节点 连接表大小
	OrdinaryNodeTableSizeMax int `json:"ordinary_node_table_size_max"`
	// 临时/新增 节点 连接表大小
	UnfamiliarNodeTableSizeMax int `json:"unfamiliar_node_table_size_max"`
}

func NewEmptyP2PConfig() *P2PConfig {
	return &P2PConfig{
		Name:                       "",
		ID:                         nil,
		TCPListenPort:              3331,
		BackboneNodeTableSizeMax:   8,
		OrdinaryNodeTableSizeMax:   32,
		UnfamiliarNodeTableSizeMax: 128,
	}
}

// new p2p
func NewP2PConfig(cnffile *sys.Inicnf) *P2PConfig {
	//data_dir := path.Join(cnffile.MustDataDir(), "node")
	//ini_section_p2p := cnffile.Section("p2p")

	cnf := NewEmptyP2PConfig()

	data_dir := path.Join(cnffile.MustDataDir(), "node")
	p2pid := readIDFromDisk(data_dir)
	if p2pid == nil {
		p2pid = make([]byte, 16)
		rand.Read(p2pid) // random id
		saveIDToDisk(data_dir, p2pid)
	}
	cnf.ID = p2pid

	// create cnf
	cnf.Datadir = data_dir
	ini_section_p2p := cnffile.Section("p2p")
	// port
	ini_listen_port := ini_section_p2p.Key("listen_port").MustInt(3331)
	cnf.TCPListenPort = ini_listen_port
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
		} else {
			fmt.Println("[P2P Config Error]", err.Error())
			os.Exit(0)
		}
	}
	//fmt.Println(cnf.StaticHnodeAddrs)
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
