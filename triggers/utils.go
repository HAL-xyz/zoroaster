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

// TODO refactor this ugly mess at some point

// check if `s` is a valid dynamic array int/uint > 32 bits in multiple of 8 bits,
// e.g. uint128[], plus int[] and uint[] which are aliases for u/int256
func isValidDynamicBigIntArray(s string) bool {
	r := regexp.MustCompile(`u?int\d{0,}\[]$`)
	if r.MatchString(s) {
		ss := strings.Split(s, "[")
		return isValidBigInt(ss[0])
	}
	return false
}

// check if `s` is a valid static array of int/uint > 32 bits in multiple of 8 bits,
// e.g. uint128[4], plus int[N] and uint[N] which are aliases for u/int256
func isValidBigIntArray(s string) bool {
	r := regexp.MustCompile(`u?int\d{0,}\[\d+]$`)
	if r.MatchString(s) {
		ss := strings.Split(s, "[")
		return isValidBigInt(ss[0])
	}
	return false
}

// checks if `s` is a valid int/uint > 32 bits in multiples of 8 bits,
// e.g. uint256
func isValidBigInt(s string) bool {
	supportedBigInts := makeBigIntSet()
	_, ok := supportedBigInts[s]
	return ok
}

// the set of all valid int/uint > 32 bits in multiples of 8 bits
func makeBigIntSet() map[string]bool {
	set := make(map[string]bool)
	key := ""
	for i := 40; i <= 256; i += 8 {
		key = "int" + strconv.Itoa(i)
		set[key] = true
		key = "uint" + strconv.Itoa(i)
		set[key] = true
	}
	set["int"] = true  // alias for int256
	set["uint"] = true // alias for uint256
	return set
}

// check if `s` is a valid dynamic array of int/uint <= 32 bits
func isValidDynamicIntArray(s string) bool {
	r := regexp.MustCompile(`u?int\d{0,}\[]$`)
	if r.MatchString(s) {
		ss := strings.Split(s, "[")
		return isValidInt(ss[0])
	}
	return false
}

// check if `s` is a valid static array of int/uint <= 32 bits in multiple of 8 bits,
func isValidIntArray(s string) bool {
	r := regexp.MustCompile(`u?int\d{0,}\[\d+]$`)
	if r.MatchString(s) {
		ss := strings.Split(s, "[")
		return isValidInt(ss[0])
	}
	return false
}

// check if `s` is a valid int/uint <= 32 bits in multiples of 8 bits
func isValidInt(s string) bool {
	set := map[string]bool{
		"int8":   true,
		"int16":  true,
		"int24":  true,
		"int32":  true,
		"uint8":  true,
		"uint16": true,
		"uint24": true,
		"uint32": true,
	}
	return set[s]
}
