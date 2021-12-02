package backend

import (
	"github.com/hacash/core/sys"
)

type BackendConfig struct {
	cnffile *sys.Inicnf

	UseBlockChainV2 bool

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
	cnf.UseBlockChainV2 = sec.Key("UseBlockChainV2").MustBool(false)

	return cnf
}
