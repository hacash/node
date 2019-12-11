package node

type HacashNodeConfig struct {
	Datadir string
}

func NewHacashNodeConfig() *HacashNodeConfig {
	return &HacashNodeConfig{}
}
