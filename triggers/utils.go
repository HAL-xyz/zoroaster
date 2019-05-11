package trigger

import (
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

func makeBigInt(s string) *big.Int {
	ret := new(big.Int)
	ret.SetString(s, 10)
	return ret
}

// check if `s` is a valid static array of int/uint in multiple of 8 bits,
// e.g. uint128[4], plus int[N] and uint[N] which are aliases for u/int256
func isValidBigIntArray(s string) bool {
	r := regexp.MustCompile(`u?int\d{0,}\[\d+]`)
	if r.MatchString(s) {
		ss := strings.Split(s, "[")
		return isValidBigInt(ss[0])
	}
	return false
}

// checks if `s` is a valid int/uint > 64 bits in multiples of 8 bits,
// e.g. uint256
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
	set["int"] = true  // alias for int256
	set["uint"] = true // alias for uint256
	return set
}
