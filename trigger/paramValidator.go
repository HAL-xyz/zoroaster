package trigger

import (
	"bytes"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"zoroaster/utils"
)

func ValidateParam(rawParam []byte, parameterType string, attribute string, predicate Predicate, index *int) (bool, interface{}) {

	var err error

	// uint8
	if parameterType == "uint8[]" {
		var param []uint8
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		tgVal, err := strconv.Atoi(attribute)
		if err == nil {
			return validatePredUIntArray(predicate, param, tgVal, index), param
		}
	}
	// address
	if parameterType == "address" {
		var param string
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return utils.NormalizeAddress(param) == utils.NormalizeAddress(attribute), param
	}
	// string
	if parameterType == "string" {
		var param string
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return strings.ToLower(param) == strings.ToLower(attribute), param
	}
	// bool
	if parameterType == "bool" {
		var param bool
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return validatePredBool(predicate, param, attribute), param
	}
	// address[]
	addressesRgx := regexp.MustCompile(`address\[\d*]$`)
	if addressesRgx.MatchString(parameterType) {
		var param []string
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		if index != nil && *index < len(param) && predicate == Eq {
			return validatePredStringArray(predicate, param, attribute, index), "0x" + param[*index]
		}
		return validatePredStringArray(predicate, param, attribute, index), param
	}
	// string[]
	stringsRgx := regexp.MustCompile(`string\[\d*]$`)
	if stringsRgx.MatchString(parameterType) {
		var param []string
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		if index != nil && *index < len(param) && predicate == Eq {
			return validatePredStringArray(predicate, param, attribute, index), param[*index]
		}
		return validatePredStringArray(predicate, param, attribute, index), param
	}
	// int
	intRgx := regexp.MustCompile(`u?int\d*$`)
	if intRgx.MatchString(parameterType) {
		var param *big.Int
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return validatePredBigInt(predicate, param, utils.MakeBigInt(attribute)), param
	}
	// u?int[] && u?int[N]
	arrayIntRgx := regexp.MustCompile(`u?int\d*\[\d*]$`)
	if arrayIntRgx.MatchString(parameterType) {
		var param []*big.Int
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		if index != nil && *index < len(param) && predicate == Eq {
			return validatePredBigIntArray(predicate, param, utils.MakeBigInt(attribute), index), param[*index]
		}
		return validatePredBigIntArray(predicate, param, utils.MakeBigInt(attribute), index), param
	}
	// byte[][]
	arrayByteRgx := regexp.MustCompile(`bytes\d*\[\d*]$`)
	if arrayByteRgx.MatchString(parameterType) {
		var param [][]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return validatePredStringArray(predicate, utils.ByteArraysToHex(param), attribute, index), param
	}
	if parameterType == "bytes" {
		var param []uint8
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param, attribute), "0x" + common.Bytes2Hex(param)
	}
	if parameterType == "bytes32" {
		var param [32]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes31" {
		var param [31]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes30" {
		var param [30]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes29" {
		var param [29]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes28" {
		var param [28]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes27" {
		var param [27]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes26" {
		var param [26]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes25" {
		var param [25]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes24" {
		var param [24]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes23" {
		var param [23]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes22" {
		var param [22]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes21" {
		var param [21]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes20" {
		var param [20]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes19" {
		var param [19]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes18" {
		var param [18]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes17" {
		var param [17]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes16" {
		var param [16]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes15" {
		var param [15]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes14" {
		var param [14]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes13" {
		var param [13]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes12" {
		var param [12]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes1" {
		var param [11]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes10" {
		var param [10]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes9" {
		var param [9]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes8" {
		var param [8]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes7" {
		var param [7]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes6" {
		var param [6]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes5" {
		var param [5]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes4" {
		var param [4]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes3" {
		var param [3]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes2" {
		var param [2]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}
	if parameterType == "bytes1" || parameterType == "byte" {
		var param [1]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return compareBytesWithParameter(param[:], attribute), "0x" + common.Bytes2Hex(param[:])
	}

	log.Debug("parameter type not supported: ", parameterType)
	return false, nil
}

func compareBytesWithParameter(b []byte, p string) bool {
	b = bytes.TrimRight(b, "\x00")

	p = strings.TrimPrefix(p, "0x")
	param := bytes.TrimRight(common.Hex2Bytes(p), "\x00")

	return common.Bytes2Hex(param) == common.Bytes2Hex(b)
}
