package utils

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

// encodes a [][]byte into a []string of hex values
func ByteArraysToHex(array [][]byte) []string {
	out := make([]string, len(array))
	for i := 0; i < len(array); i++ {
		out[i] = hex.EncodeToString(array[i])
	}
	return out
}

func MakeBigInt(s string) *big.Int {
	ret := new(big.Int)
	ret.SetString(s, 10)
	return ret
}

func MakeBigIntFromHex(s string) *big.Int {
	s = strings.Replace(s, "0x", "", 1)
	ret := new(big.Int)
	ret.SetString(s, 16)
	return ret
}

func StripCtlAndExtFromUTF8(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 32 && r < 127 {
			return r
		}
		return -1
	}, s)
}

func SplitStringByLength(s string, length int) []string {
	regexs := fmt.Sprintf(`(\S{%d})`, length)
	re := regexp.MustCompile(regexs)
	return re.FindAllString(s, -1)
}

func RemoveUntil(s string, until rune) string {
	if idx := strings.IndexRune(s, until); idx >= 0 {
		return s[idx:]
	}
	return s
}

func RemoveCharacters(input string, characters string) string {
	filter := func(r rune) rune {
		if strings.IndexRune(characters, r) < 0 {
			return r
		}
		return -1
	}
	return strings.Map(filter, input)
}

func GetOnlyNumbers(s string) string {
	re := regexp.MustCompile("[0-9]+")
	return re.FindString(s)
}

func GetValuesFromMap(m map[string]json.RawMessage) []json.RawMessage {
	v := make([]json.RawMessage, len(m), len(m))
	idx := 0
	for _, value := range m {
		v[idx] = value
		idx++
	}
	return v
}

func Reverse(s string) string {
	// Get Unicode code points.
	n := 0
	rune := make([]rune, len(s))
	for _, r := range s {
		rune[n] = r
		n++
	}
	rune = rune[0:n]
	// Reverse
	for i := 0; i < n/2; i++ {
		rune[i], rune[n-1-i] = rune[n-1-i], rune[i]
	}
	// Convert back to UTF-8.
	return string(rune)
}
