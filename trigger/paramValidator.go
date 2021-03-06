package trigger

import (
	"bytes"
	"encoding/json"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/HAL-xyz/zoroaster/utils"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

// According to https://solidity.readthedocs.io/en/v0.5.3/abi-spec.html#abi-events
//
// For all fixed-length Solidity types, the EVENT_INDEXED_ARGS array contains
// the 32-byte encoded value directly.
// However, for types of dynamic length, which include string, bytes, and arrays,
// EVENT_INDEXED_ARGS will contain the Keccak hash of the packed encoded value
// rather than the encoded value directly.
//
// This means that when validating strings, bytes and arrays
// we simply compare their Keccak hash

func ValidateTopicParam(topicParam, paramType, paramCurrency string, condition ConditionEvent, tokenApi tokenapi.ITokenAPI) (bool, string) {
	attribute := condition.Attribute

	// bool
	if paramType == "bool" {
		if topicParam == "0x0000000000000000000000000000000000000000000000000000000000000001" {
			return "true" == strings.ToLower(attribute), topicParam
		} else {
			return "false" == strings.ToLower(attribute), topicParam
		}
	}

	// bool[], bool[N] - Keccak hash, only Eq supported
	if strings.HasPrefix(paramType, "bool[") {
		return strings.ToLower(topicParam) == strings.ToLower(attribute), topicParam
	}

	// string, string[], string[N] - Keccak hash, only Eq supported
	if strings.HasPrefix(paramType, "string") {
		return strings.ToLower(topicParam) == strings.ToLower(attribute), topicParam
	}

	// bytes1...32, bytes
	if strings.HasPrefix(paramType, "bytes") {
		return compareBytesWithParameter(common.Hex2Bytes(strings.TrimPrefix(topicParam, "0x")), attribute), topicParam
	}

	// address
	if paramType == "address" {
		return utils.NormalizeAddress(topicParam) == utils.NormalizeAddress(attribute), topicParam
	}

	// address[], address[N] - Keccak hash, only Eq supported
	if strings.HasPrefix(paramType, "address[") {
		return strings.ToLower(topicParam) == strings.ToLower(attribute), topicParam
	}

	// int
	intRgx := regexp.MustCompile(`u?int\d*$`)
	if intRgx.MatchString(paramType) {
		if paramCurrency != "" {
			convertedValue, err := convertToCurrency(tokenApi, paramCurrency, condition.AttributeCurrency, utils.MakeBigInt(topicParam))
			if err != nil {
				return false, ""
			}
			return validatePredBigFloat(condition.Predicate, convertedValue, utils.MakeBigFloat(attribute)), convertedValue.String()
		}

		return validatePredBigInt(condition.Predicate, utils.MakeBigInt(topicParam), utils.MakeBigInt(attribute)), topicParam
	}

	// int[N] and int[] - Keccak hash, only Eq supported
	arrayIntRgx := regexp.MustCompile(`u?int\d*\[\d*]$`)
	if arrayIntRgx.MatchString(paramType) {
		return strings.ToLower(topicParam) == strings.ToLower(attribute), topicParam
	}

	log.Debug("topic parameter type not supported: ", paramType)
	return false, ""
}

func ValidateParam(
	ifcParam interface{},
	parameterType, parameterCurrency, attribute, attributeCurrency string,
	predicate Predicate,
	index *int,
	component Component,
	tokenApi tokenapi.ITokenAPI) (bool, interface{}) {

	var err error
	rawParam, _ := json.Marshal(ifcParam)

	// tuple
	if parameterType == "tuple" {
		var param map[string]json.RawMessage
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		return ValidateParam(param[component.Name], component.Type, parameterCurrency, attribute, attributeCurrency, predicate, index, Component{}, tokenApi)
	}
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
		param = strings.ReplaceAll(param, "\x00", "")
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
	// bool[], bool[N]
	if strings.HasPrefix(parameterType, "bool[") {
		var param []bool
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false, nil
		}
		if index != nil && *index < len(param) && predicate == Eq {
			return validatePredBool(predicate, param[*index], attribute), param[*index]
		}
		return validatePredBoolArray(predicate, param, attribute, index), param
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
		if parameterCurrency != "" {
			convertedValue, err := convertToCurrency(tokenApi, parameterCurrency, attributeCurrency, param)
			if err != nil {
				return false, nil
			}
			return validatePredBigFloat(predicate, convertedValue, utils.MakeBigFloat(attribute)), param
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
			if parameterCurrency != "" {
				convertedValue, err := convertToCurrency(tokenApi, parameterCurrency, attributeCurrency, param[*index])
				if err != nil {
					return false, nil
				}
				return validatePredBigFloat(predicate, convertedValue, utils.MakeBigFloat(attribute)), param
			}
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

	log.Debug("data parameter type not supported: ", parameterType)
	return false, nil
}

func compareBytesWithParameter(b []byte, p string) bool {
	b = bytes.TrimRight(b, "\x00")

	p = strings.TrimPrefix(p, "0x")
	param := bytes.TrimRight(common.Hex2Bytes(p), "\x00")

	return common.Bytes2Hex(param) == common.Bytes2Hex(b)
}

func convertToCurrency(tokenApi tokenapi.ITokenAPI, parameterCurrency, attributeCurrency string, param *big.Int) (*big.Float, error) {
	exchangeRate, err := tokenApi.GetExchangeRate(parameterCurrency, attributeCurrency)
	if err != nil {
		return new(big.Float), err
	}
	decimals := tokenApi.Decimals(parameterCurrency)
	scaledValue := utils.MakeBigFloat(tokenApi.FromWei(param, decimals))
	convertedValue := scaledValue.Mul(scaledValue, utils.MakeBigFloat(exchangeRate))
	return convertedValue, nil
}
