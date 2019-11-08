package utils

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
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

func IsIn(a string, list []string) bool {
	for _, x := range list {
		if x == a {
			return true
		}
	}
	return false
}

func AreEqualJSON(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 1 :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 2 :: %s", err.Error())
	}

	return reflect.DeepEqual(o1, o2), nil
}

func GimmePrettyJson(o interface{}) (string, error) {
	bytesObj, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, bytesObj, "", "  ")
	if err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}

func GetSliceFromIntSet(set map[string]struct{}) []string {
	out := make([]string, len(set))
	i := 0
	for k := range set {
		out[i] = k
		i++
	}
	return out
}

func SetDifference(s1 map[string]struct{}, s2 map[string]struct{}) map[string]struct{} {
	diff := make(map[string]struct{})
	for v := range s1 {
		_, ok := s2[v]
		if ok {
			continue
		}
		diff[v] = struct{}{}
	}
	return diff
}
