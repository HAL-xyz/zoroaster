package trigger

import (
	"strconv"
)

// checks if `s` is a valid int/uint > 64 bits in multiples of 8 bits
func isValidBigInt(s string) bool {
	supportedBigInts := makeBigIntsSet()
	_, ok := supportedBigInts[s]
	return ok
}

// the set of all valid int/uint > 64 bits in multiples of 8 bits
func makeBigIntsSet() map[string]bool {
	set := make(map[string]bool)
	key := ""
	for i := 64; i <= 256; i += 8 {
		key = "int" + strconv.Itoa(i)
		set[key] = true
		key = "uint" + strconv.Itoa(i)
		set[key] = true
	}
	return set
}
