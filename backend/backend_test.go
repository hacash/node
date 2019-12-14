package backend

import (
	"fmt"
	"github.com/hacash/core/sys"
	"testing"
	"time"
)

func Test_t1(t *testing.T) {

	testcnffilename := "/home/shiqiujie/Desktop/Hacash/go/src/github.com/hacash/node/node/hacash.config.test.ini"

	cnffile, cnfe := sys.LoadInicnf(testcnffilename)
	if cnfe != nil {
		fmt.Println(cnfe)
		return
	}

	hcnf := NewBackendConfig(cnffile)
	hnode, e2 := NewBackend(hcnf)
	if e2 != nil {
		fmt.Println(e2)
		return
	}

	hnode.Start()

	<-time.Tick(time.Hour * 100)

}
