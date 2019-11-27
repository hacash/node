package p2p

import "bytes"

// change table and return drop one
func UpdateRelationshipTable(base []byte, table [][]byte, tablemaxsize int, newtwo []byte) ([][]byte, []byte) {
	newtable := InsertIntoRelationshipTable(base, table, newtwo)
	var dropone []byte = nil
	if len(newtable) > tablemaxsize {
		dropone = newtable[len(newtable)-1]
	}
	return newtable, dropone
}

//
func CalculateRelationshipTable(base []byte, othernodes [][]byte, maxnum int) [][]byte {
	tables := [][]byte{}
	for i := 0; i < len(othernodes); i++ {
		tables = InsertIntoRelationshipTable(base, tables, othernodes[i])
		if maxnum > 0 && len(tables) > maxnum {
			tables = tables[0:maxnum]
		}
	}
	return tables
}

//
func InsertIntoRelationshipTable(base []byte, table [][]byte, newtwo []byte) [][]byte {
	length := len(table)
	if length == 0 {
		table = append([][]byte{}, newtwo)
		return table
	}
	if length == 1 {
		cpres := CompareRelationshipByHash(base, newtwo, table[0])
		if cpres == 1 {
			tmp := append([][]byte{}, newtwo)
			tmp = append(tmp, table...)
			return tmp
		} else {
			tmp := append([][]byte{}, table...)
			tmp = append(tmp, newtwo)
			return tmp
		}
	}
	for i := length - 1; i >= 0; i-- {
		cpr1 := CompareRelationshipByHash(base, table[i], newtwo)
		//fmt.Println("ti i", i, table[i], cpr1)
		if i == 0 {
			if cpr1 == -1 {
				tmp := append([][]byte{}, newtwo)
				tmp = append(tmp, table...)
				return tmp
			} else {
				tmp := append([][]byte{}, table[0])
				tmp = append(tmp, newtwo)
				tmp = append(tmp, table[1:]...)
				return tmp
			}
		} else if i == length-1 {
			if cpr1 == 1 {
				tmp := append([][]byte{}, table...)
				tmp = append(tmp, newtwo)
				return tmp
			}
		} else {
			cpr2 := CompareRelationshipByHash(base, table[i+1], newtwo)
			if cpr1 > -1 && cpr2 < 1 {
				// insert
				ttl := append([][]byte{}, table[:i+1]...)
				ttr := append([][]byte{}, table[i+1:]...)
				tmp := append(ttl, newtwo)
				tmp = append(tmp, ttr...)
				return tmp
			}
		}
	}
	//fmt.Println("ttt aaa bbb lll eee")
	return table
}

//
func CompareRelationshipByHash(base []byte, one []byte, two []byte) int {

	compareLen := len(base)
	if compareLen > len(one) {
		compareLen = len(one)
	}
	if compareLen > len(two) {
		compareLen = len(two)
	}
	for i := 0; i < compareLen; i++ {
		ar1 := calculateByteDistance(base[i], one[i])
		ar2 := calculateByteDistance(base[i], two[i])
		if ar1 < ar2 {
			return 1
		} else if ar1 > ar2 {
			return -1
		}
	}
	return 0
}

//
func calculateByteDistance(a byte, b byte) uint8 {
	if a == b {
		return 0
	}
	if a < b {
		a, b = b, a
	}
	res := a - b
	if res > 128 {
		return uint8(256 - int(res))
	} else {
		return res
	}
}

//////////////////////////////////////////////////////////////////////////

func deleteBytesFromList(list [][]byte, del []byte) [][]byte {
	var deli = -1
	for i, v := range list {
		if bytes.Compare(v, del) == 0 {
			deli = i
			break
		}
	}
	if deli == -1 {
		return list
	} else if deli == 0 {
		return list[1:]
	} else if deli == len(list)-1 {
		return list[0 : len(list)-1]
	} else if deli > 0 {
		tmp1 := list[0:deli]
		tmp2 := list[deli+1:]
		return append(tmp1, tmp2...)
	}
	return list
}
