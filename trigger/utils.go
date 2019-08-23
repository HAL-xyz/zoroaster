package trigger

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

// encodes a [][]byte into a []string of hex values
func byteArraysToHex(array [][]byte) []string {
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

func makeBigIntFromHex(s string) *big.Int {
	s = strings.Replace(s, "0x", "", 1)
	ret := new(big.Int)
	ret.SetString(s, 16)
	return ret
}

func stripCtlAndExtFromUTF8(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 32 && r < 127 {
			return r
		}
		return -1
	}, s)
}

func getOnlyNumbers(s string) string {
	re := regexp.MustCompile("[0-9]+")
	return re.FindString(s)
}

func splitStringByLength(s string, length int) []string {
	regexs := fmt.Sprintf(`(\S{%d})`, length)
	re := regexp.MustCompile(regexs)
	return re.FindAllString(s, -1)
}

func removeUntil(s string, until rune) string {
	if idx := strings.IndexRune(s, until); idx >= 0 {
		return s[idx:]
	}
	return s
}
