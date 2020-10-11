package p2pv2

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"
)

func Test_t3(t *testing.T) {

	idtables := []PeerID{
		PeerID([]byte{1, 2, 3, 4}),
		PeerID([]byte{5, 6, 7, 8}),
		PeerID([]byte{1, 1, 1, 1}),
		PeerID([]byte{9, 9, 9, 9}),
	}

	fmt.Println(removePeerIDFromTable(idtables, []byte{0, 0, 0, 0}))
	fmt.Println(removePeerIDFromTable(idtables, []byte{1, 2, 3, 4}))
	fmt.Println(removePeerIDFromTable(idtables, []byte{5, 6, 7, 8}))
	fmt.Println(removePeerIDFromTable(idtables, []byte{1, 1, 1, 1}))
	fmt.Println(removePeerIDFromTable(idtables, []byte{9, 9, 9, 9}))
	fmt.Println(removePeerIDFromTable(idtables, []byte{4, 3, 2, 1}))

}

func Test_t2(t *testing.T) {

	idtables := []PeerID{}
	oripeerid := bytes.Repeat([]byte{128}, 32)

	for i := 0; i < 1600; i++ {
		fmt.Println(i, "------------------------------------------")
		tmppeerid := bytes.Repeat([]byte{0}, 32)
		rand.Read(tmppeerid)
		istOk, istId, newidtables, dropid := insertUpdateTopologyDistancePeerIDTable(oripeerid, tmppeerid, idtables, 8)
		if istOk {
			fmt.Println(istOk, istId, tmppeerid)
			for k, v := range newidtables {
				fmt.Println("    ", k, v)
			}
			fmt.Println("dropid:", dropid)
		}
		idtables = newidtables
	}

	//for k, v := range idtables {
	//	fmt.Println("        ", k, v)
	//}

	//insertUpdateTopologyDistancePeerIDTable()

}

func Test_t1(t *testing.T) {

	fmt.Println(compareTopologyDistanceForPeerID([]byte{0, 0}, []byte{1, 0}, []byte{2, 4}))

}

func Test_t0(t *testing.T) {

	fmt.Println(calculateOneByteTopologyDistanceValue(1, 0))
	fmt.Println(calculateOneByteTopologyDistanceValue(255, 0))
	fmt.Println(calculateOneByteTopologyDistanceValue(0, 1))
	fmt.Println(calculateOneByteTopologyDistanceValue(0, 255))
	fmt.Println(calculateOneByteTopologyDistanceValue(250, 2))
	fmt.Println(calculateOneByteTopologyDistanceValue(200, 160))
	fmt.Println(calculateOneByteTopologyDistanceValue(200, 220))
	fmt.Println(calculateOneByteTopologyDistanceValue(1, 240))

	fmt.Println(calculateOneByteTopologyDistanceValue(128, 128))
	fmt.Println(calculateOneByteTopologyDistanceValue(128, 127))
	fmt.Println(calculateOneByteTopologyDistanceValue(128, 129))

}
