package backend

import (
	"github.com/hacash/core/sys"
)

type BackendConfig struct {
	cnffile *sys.Inicnf

	UseBlockChainV3 bool

	Datadir string
}

func NewEmptyBackendConfig() *BackendConfig {
	return &BackendConfig{}
}

func NewBackendConfig(cnffile *sys.Inicnf) *BackendConfig {
	cnf := NewEmptyBackendConfig()

	cnf.cnffile = cnffile

	cnf.Datadir = cnffile.MustDataDirWithVersion()

	sec := cnffile.Section("")
	cnf.UseBlockChainV3 = sec.Key("UseBlockChainV3").MustBool(false)

	return cnf
}
