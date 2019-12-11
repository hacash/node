package p2p

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"
)

func Test_t4(t *testing.T) {

	allids := [][]byte{}
	itnum := uint8(5)
	for i0 := uint8(0); i0 < itnum; i0++ {
		for i1 := uint8(0); i1 < itnum; i1++ {
			for i2 := uint8(0); i2 < itnum; i2++ {
				for i3 := uint8(0); i3 < itnum; i3++ {
					allids = append(allids, []byte{i0, i1, i2, i3})
				}
			}
		}
	}
	//fmt.Println(allids)

	self := allids[len(allids)-1]
	fmt.Println(self)
	tartables := CalculateRelationshipTable(self, allids, 8)
	fmt.Println(tartables)

}

func Test_t3(t *testing.T) {

	fmt.Println(calculateByteDistance(0, 10))
	fmt.Println(calculateByteDistance(0, 246))

}

func Test_t2(t *testing.T) {

	self := []byte{0, 0, 0, 0}
	table := [][]byte{
		[]byte{0, 0, 0, 1},
		[]byte{0, 4, 8, 16},
		[]byte{0, 100, 255, 255},
		[]byte{100, 0, 5, 5},
	}

	table = InsertIntoRelationshipTable(self, table, []byte{0, 1, 0, 0})
	table = InsertIntoRelationshipTable(self, table, []byte{0, 99, 0, 0})
	table = InsertIntoRelationshipTable(self, table, []byte{255, 0, 0, 0})

	fmt.Println(table)

}

func Test_t1(t *testing.T) {

	self_peerid := bytes.Repeat([]byte{100}, 32)
	intimate_peerids := make([][]byte, 0, 16)

	fmt.Println(self_peerid)

	for i := 0; i < 10000000; i++ {
		peerid := make([]byte, 32)
		rand.Read(peerid)

		intimate_peerids = InsertIntoRelationshipTable(self_peerid, intimate_peerids, peerid)
		if len(intimate_peerids) > 16 {
			intimate_peerids = intimate_peerids[0:16]
		}

	}

	fmt.Println("- - - - - - - -")

	for i := 0; i < len(intimate_peerids); i++ {
		fmt.Println(intimate_peerids[i])
	}

	//kkk := []byte{0,1,2,3,4}
	//fmt.Println(kkk[:2])
	//fmt.Println(kkk[2:])

}
