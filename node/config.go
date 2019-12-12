package node

import "github.com/hacash/core/sys"

type HacashNodeConfig struct {
	cnffile *sys.Inicnf

	Datadir string
}

func NewEmptyHacashNodeConfig() *HacashNodeConfig {
	return &HacashNodeConfig{}
}

func NewHacashNodeConfig(cnffile *sys.Inicnf) *HacashNodeConfig {
	cnf := NewEmptyHacashNodeConfig()

	cnf.cnffile = cnffile

	return cnf
}
