package p2p

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"github.com/hacash/core/sys"
	"os"
	"path"
)

type P2PManagerConfig struct {
	Datadir             string
	TCPListenPort       int
	UDPListenPort       int
	Name                string
	ID                  []byte
	lookupConnectMaxLen int
}

func NewP2PManagerConfigByID(id []byte) *P2PManagerConfig {
	if len(id) != 16 {
		panic("P2PManagerConfig ID len must be 16.")
	}
	cnf := &P2PManagerConfig{
		Name:                "hnode_" + hex.EncodeToString(id),
		ID:                  id,
		TCPListenPort:       3337,
		UDPListenPort:       3336,
		lookupConnectMaxLen: 128,
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
