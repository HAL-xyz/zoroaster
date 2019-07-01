package trigger

import (
	"encoding/hex"
	"math/big"
)

// encodes a [][]byte into a []string of hex values
func ByteArraysToHex(array [][]byte) []string {
	out := make([]string, len(array))
	for i := 0; i < len(array); i++ {
		out[i] = hex.EncodeToString(array[i])
	}
	return out
}

func makeBigInt(s string) *big.Int {
	ret := new(big.Int)
	ret.SetString(s, 10)
	return ret
}
