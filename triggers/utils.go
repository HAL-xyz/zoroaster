package trigger

import (
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

const dyArrayIntRx = `u?int\d{0,}\[]$`
const stArrayIntRx = `u?int\d{0,}\[\d+]$`
const dyArrayBytesRx = `bytes\d{1,2}\[]$`

func makeBigInt(s string) *big.Int {
	ret := new(big.Int)
	ret.SetString(s, 10)
	return ret
}

func isValidArray(s string, rg string, validate func(string) bool) bool {
	r := regexp.MustCompile(rg)
	if r.MatchString(s) {
		ss := strings.Split(s, "[")
		return validate(ss[0])
	}
	return false
}

// checks if `s` is a valid int/uint > 32 bits in multiples of 8 bits,
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

// check if `s` is a valid u/int <= 32 bits in multiples of 8 bits
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

func isValidByte(s string) bool {
	set := map[string]bool{
		"bytes":   true,
		"bytes1":  true,
		"bytes2":  true,
		"bytes3":  true,
		"bytes4":  true,
		"bytes5":  true,
		"bytes6":  true,
		"bytes7":  true,
		"bytes8":  true,
		"bytes9":  true,
		"bytes10": true,
		"bytes11": true,
		"bytes12": true,
		"bytes13": true,
		"bytes14": true,
		"bytes15": true,
		"bytes16": true,
		"bytes17": true,
		"bytes18": true,
		"bytes19": true,
		"bytes20": true,
		"bytes21": true,
		"bytes22": true,
		"bytes23": true,
		"bytes24": true,
		"bytes25": true,
		"bytes26": true,
		"bytes27": true,
		"bytes28": true,
		"bytes29": true,
		"bytes30": true,
		"bytes31": true,
		"bytes32": true,
	}
	return set[s]
}
