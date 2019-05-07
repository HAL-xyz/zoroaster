package trigger

import (
	"encoding/hex"
	"reflect"
)

// decodes a static bytes array casted as an interface back to a slice
func DecodeBytesArray(array interface{}, size int) []byte {
	out := make([]byte, size)
	for i := 0; i < size; i++ {
		idxval := reflect.ValueOf(array).Index(i)
		uidxval := uint8(idxval.Uint())
		out[i] = uidxval
	}
	return out
}

// decodes a static [][]bytes casted as an interface back to a [][]bytes (slice)
// this is used e.g. for Solidity's bytesN[], which corresponds to uint8[][N]
func Decode2DBytesArray(array interface{}) [][]byte {

	outer := reflect.ValueOf(array)
	outres := make([][]byte, outer.Len())

	if outer.Kind() == reflect.Slice || outer.Kind() == reflect.Array {
		for i := 0; i < outer.Len(); i++ {
			inner := outer.Index(i)
			if inner.Kind() == reflect.Array {
				outres[i] = make([]byte, inner.Len())
				for j := 0; j < inner.Len(); j++ {
					item := inner.Index(j)
					uidxval := uint8(item.Uint())
					outres[i][j] = uidxval
				}
			}
		}
	}
	return outres
}

// decodes a [][]byte into a []string of hex values
func MultArrayToHex(array [][]byte) []string {
	out := make([]string, len(array))
	for i := 0; i < len(array); i++ {
		out[i] = hex.EncodeToString(array[i])
	}
	return out
}
