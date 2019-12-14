package backend

import "github.com/hacash/core/sys"

type BackendConfig struct {
	cnffile *sys.Inicnf

	Datadir string
}

func NewEmptyBackendConfig() *BackendConfig {
	return &BackendConfig{}
}

func NewBackendConfig(cnffile *sys.Inicnf) *BackendConfig {
	cnf := NewEmptyBackendConfig()

	cnf.cnffile = cnffile

	return cnf
}
