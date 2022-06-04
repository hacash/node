package p2pv2

import (
	"bytes"
)

// Remove a from the table
func removePeerIDFromTable(idtables []PeerID, rmid PeerID) (bool, []PeerID) {
	rmidx := -1
	for i, v := range idtables {
		if bytes.Compare(rmid, v) == 0 {
			rmidx = i
			break
		}
	}
	if rmidx == -1 {
		return false, idtables
	}
	// remove
	var newtable = []PeerID{}
	if rmidx == 0 {
		return true, append(newtable, idtables[1:]...)
	}
	tail := len(idtables) - 1
	if rmidx == tail {
		return true, append(newtable, idtables[0:tail]...)
	}
	//
	newtable = append(newtable, idtables[0:rmidx]...)
	newtable = append(newtable, idtables[rmidx+1:]...)
	return true, newtable
}

/**
 * 判断表单是否可以/会被更新
 */
func isCanUpdateTopologyDistancePeerIDTable(ori PeerID, insert PeerID, idtables []PeerID, tablesizemax int) bool {
	rlnum := len(idtables)
	if tablesizemax <= 0 {
		return false // Tablespace is zero and cannot be updated
	}
	if rlnum < tablesizemax {
		return true // Empty table or quantity not full, can be updated
	}
	// First judgment comparison in the table
	lastp := idtables[0]
	dsv := compareTopologyDistanceForPeerID(ori, insert, lastp)
	if dsv == 1 {
		// The insert relationship is closer and can be updated
		return true // Empty table or quantity not full, can be updated
	}
	// Insufficient kinship, unable to update
	return false
}

/**
 * 更新亲源拓扑表
 */
func insertUpdateTopologyDistancePeerIDTable(ori PeerID, insert PeerID, idtables []PeerID, tablesizemax int) (inserted bool, insertIdx int, tables []PeerID, dropid PeerID) {
	if tablesizemax <= 0 {
		panic("tablesize cannot be 0.")
	}
	inserted = false
	insertIdx = -1
	tables = nil
	dropid = nil
	// New addition
	tblen := len(idtables)
	if tblen == 0 {
		tables = []PeerID{insert}
		insertIdx = 0
		inserted = true
		return
	}
	// insert
	inserted = true
	if tblen == 1 {
		dsv := compareTopologyDistanceForPeerID(ori, insert, idtables[0])
		//fmt.Println("compareTopologyDistanceForPeerID: ", dsv)
		if 1 == dsv {
			tables = []PeerID{insert, idtables[0]}
			insertIdx = 0
		} else {
			tables = []PeerID{idtables[0], insert}
			insertIdx = 1
		}
	} else {
		insertIdx = 0
		for i := tblen - 1; i >= 0; i-- {
			dsv := compareTopologyDistanceForPeerID(ori, idtables[i], insert)
			//fmt.Println("compareTopologyDistanceForPeerID: ", dsv)
			if 1 == dsv {
				insertIdx = i + 1
				break
			}
		}
		//fmt.Println("insertIdx:", insertIdx)
		newidtbs := []PeerID{}
		newidtbs = append(newidtbs, idtables[0:insertIdx]...)
		newidtbs = append(newidtbs, insert)
		newidtbs = append(newidtbs, idtables[insertIdx:]...)
		tables = newidtbs
	}
	// drop ?
	var ntblen = len(tables)
	if ntblen > tablesizemax {
		dropid = tables[ntblen-1]
		tables = tables[0 : ntblen-1]
	}
	if bytes.Compare(dropid, insert) == 0 {
		// The one at the end
		inserted = false
		insertIdx = -1
	}
	// return
	return

}

/**
 * 比较两个节点id  哪个的拓扑距离与源节点更近
 * left  更近返回  1
 * right 更近返回 -1
 * 一样近则返回     0
 */
func compareTopologyDistanceForPeerID(ori, left, right []byte) int {
	tarlen := len(ori)
	if len(left) < tarlen {
		tarlen = len(left)
	}
	if len(right) < tarlen {
		tarlen = len(right)
	}

	//fmt.Println(ori, left, right)

	for i := 0; i < tarlen; i++ {
		ds1 := calculateOneByteTopologyDistanceValue(ori[i], left[i])
		ds2 := calculateOneByteTopologyDistanceValue(ori[i], right[i])
		if ds1 < ds2 {
			return 1
		} else if ds1 > ds2 {
			return -1
		}
		// diff next byte
	}
	return 0
}

// Calculate distance value
func calculateOneByteTopologyDistanceValue(dst, src byte) uint8 {
	if dst == src {
		return 0
	}
	var (
		bt uint8
		st uint8
	)
	if dst > src {
		bt, st = dst, src
	} else {
		bt, st = src, dst
	}
	// distance
	disv := int(bt - st)
	if disv > 128 {
		disv = 256 - disv
	}
	return uint8(disv)
}
