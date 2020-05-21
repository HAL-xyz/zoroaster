package utils

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"reflect"
	"regexp"
	"strings"
)

func ComposeStringFns(fns ...func(string) string) func(string) string {

	return func(s string) string {
		for _, f := range fns {
			s = f(s)
		}
		return s
	}
}

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

// sometimes addresses (40 hex chars) are padded with 0s to
// make them 64 chars long. In this case we want to strip them.
func NormalizeAddress(add string) string {
	if len(add) == 66 && strings.HasPrefix(add, "0x000000000000000000000000") {
		standardAdd := strings.Replace(add, "0x000000000000000000000000", "0x", 1)
		return strings.ToLower(standardAdd)
	}
	return strings.ToLower(add)
}

// given "[10, 20, 30]" returns []string{"10", "20", "30"}
func GetValsFromStringifiedArray(a string) []string {
	a = RemoveCharacters(a, "[] ")
	return strings.Split(a, ",")
}

// given an []interfaces{} returns and []interfaces{} where
// all the objects are the sprintf'd version of the original ones.
// just using fmt.Sprintf here isn't enough because for some types
// we want to do some custom formatting:
//
// []byte => hex representation
// []address and address => normalized address
// []int{1,2,3} => []string{"1", "2", "3"}
//
// note that the order *IS* important: bytes must go first
func SprintfInterfaces(ls []interface{}) []interface{} {
	out := make([]interface{}, len(ls))
	for i, e := range ls {
		// bytes
		v, ok := e.([]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v)
			continue
		}
		v32, ok := e.([32]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v32[:])
			continue
		}
		v31, ok := e.([31]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v31[:])
			continue
		}
		v30, ok := e.([30]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v30[:])
			continue
		}
		v29, ok := e.([29]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v29[:])
			continue
		}
		v28, ok := e.([28]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v28[:])
			continue
		}
		v27, ok := e.([27]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v27[:])
			continue
		}
		v26, ok := e.([26]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v26[:])
			continue
		}
		v25, ok := e.([25]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v25[:])
			continue
		}
		v24, ok := e.([24]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v24[:])
			continue
		}
		v23, ok := e.([23]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v23[:])
			continue
		}
		v22, ok := e.([22]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v22[:])
			continue
		}
		v21, ok := e.([21]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v21[:])
			continue
		}
		v20, ok := e.([20]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v20[:])
			continue
		}
		v19, ok := e.([19]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v19[:])
			continue
		}
		v18, ok := e.([18]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v18[:])
			continue
		}
		v17, ok := e.([17]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v17[:])
			continue
		}
		v16, ok := e.([16]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v16[:])
			continue
		}
		v15, ok := e.([15]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v15[:])
			continue
		}
		v14, ok := e.([14]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v14[:])
			continue
		}
		v13, ok := e.([13]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v13[:])
			continue
		}
		v12, ok := e.([12]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v12[:])
			continue
		}
		v11, ok := e.([11]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v11[:])
			continue
		}
		v10, ok := e.([10]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v10[:])
			continue
		}
		v9, ok := e.([9]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v9[:])
			continue
		}
		v8, ok := e.([8]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v8[:])
			continue
		}
		v7, ok := e.([7]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v7[:])
			continue
		}
		v6, ok := e.([6]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v6[:])
			continue
		}
		v5, ok := e.([5]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v5[:])
			continue
		}
		v4, ok := e.([4]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v4[:])
			continue
		}
		v3, ok := e.([3]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v3[:])
			continue
		}
		v2, ok := e.([2]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v2[:])
			continue
		}
		v1, ok := e.([1]uint8)
		if ok {
			out[i] = common.Bytes2Hex(v1[:])
			continue
		}
		// common.Address
		add, ok := e.(common.Address)
		if ok {
			out[i] = NormalizeAddress(add.String())
			continue
		}
		// []common.Address
		out[i], ok = decodeAddressArray(e)
		if ok {
			continue
		}
		// []Int - any kind
		out[i], ok = decodeIntArray(e)
		if ok {
			continue
		}
		// default case, where good old Sprintf is enough
		out[i] = fmt.Sprintf("%v", e)
		continue
	}
	return out
}

// []interface{}{[]common.Address{...}} => []interface{}{[]string{...}
// where the output string is a "normalized" version of the address
func decodeAddressArray(array interface{}) ([]string, bool) {
	worked := false
	var out []string
	a := reflect.ValueOf(array)

	if a.Kind() == reflect.Array || a.Kind() == reflect.Slice {
		out = make([]string, a.Len())
		for i := 0; i < a.Len(); i++ {
			idxval := reflect.ValueOf(array).Index(i)
			aidxval, ok := idxval.Interface().(common.Address)
			if ok {
				out[i] = NormalizeAddress(aidxval.String())
				worked = ok
			}
		}
	}
	return out, worked
}

// []interface{}{1,2,3} => []interface{}{[]string{"1","2","3"}
// works with every type of int
func decodeIntArray(array interface{}) ([]string, bool) {
	worked := false
	var out []string
	a := reflect.ValueOf(array)

	if a.Kind() == reflect.Array || a.Kind() == reflect.Slice {
		worked = true
		out = make([]string, a.Len())
		for i := 0; i < a.Len(); i++ {
			idxval := reflect.ValueOf(array).Index(i)
			out[i] = fmt.Sprintf("%v", idxval.Interface())
		}
	}
	return out, worked // only return worked=true on arrays and slices
}
