package p2p

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"github.com/hacash/core/inicnf"
	"github.com/hacash/core/sys"
	"os"
	"path"
)

func readIDFromDisk(data_dir string) []byte {
	var p2pid []byte = nil
	idjsonobj := struct {
		id string
	}{}
	idfile, e1 := os.OpenFile(path.Join(data_dir, "id.json"), os.O_RDWR|os.O_CREATE, 0777)
	if e1 == nil {
		idjsonstr := make([]byte, 16*4)
		rn, e := idfile.Read(idjsonstr)
		if e == nil {
			if json.Unmarshal(idjsonstr[0:rn], idjsonobj) == nil {
				id, e := hex.DecodeString(idjsonobj.id)
				if e == nil && len(id) >= 16 {
					p2pid = id[0:16]
				}
			}
		}
		// save
		if p2pid == nil {
			p2pid = make([]byte, 16)
			rand.Read(p2pid) // random id
			// save to disk
			idjsonobj.id = hex.EncodeToString(p2pid)
			jsonbts, e := json.Marshal(idjsonobj)
			if e == nil {
				idfile.Write(jsonbts)
			}
		}
		idfile.Close()
	}
	// return
	return p2pid
}

// new p2p
func NewP2PManagerByIniCnf(cnffile *inicnf.File) (*P2PManager, error) {
	data_dir := path.Join(sys.CnfMustDataDir(cnffile.Section("").Key("data_dir").String()), "node")
	p2pid := readIDFromDisk(data_dir)
	// create cnf
	cnf := NewP2PManagerConfig(p2pid)
	peercnf := NewPeerManagerConfig()

	return NewP2PManager(cnf, peercnf)
}
