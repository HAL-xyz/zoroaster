package trigger

import (
	"encoding/hex"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"zoroaster/utils"
)

func ValidateParam(rawParam []byte, parameterType string, attribute string, predicate Predicate, index *int) bool {

	var err error

	// uint8
	if parameterType == "uint8[]" {
		var param []uint8
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false
		}
		tgVal, err := strconv.Atoi(attribute)
		if err == nil {
			return validatePredUIntArray(predicate, param, tgVal, index)
		}
	}
	// address
	if parameterType == "address" {
		var param string
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false
		}
		return utils.NormalizeAddress(param) == strings.ToLower(attribute)
	}
	// string
	if parameterType == "string" {
		var param string
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false
		}
		return strings.ToLower(param) == strings.ToLower(attribute)
	}
	// bool
	if parameterType == "bool" {
		var param bool
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false
		}
		return validatePredBool(predicate, param, attribute)
	}
	// address[]
	addressesRgx := regexp.MustCompile(`address\[\d*]$`)
	if addressesRgx.MatchString(parameterType) {
		var param []string
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false
		}
		return validatePredStringArray(predicate, param, attribute, index)
	}
	// string[]
	stringsRgx := regexp.MustCompile(`string\[\d*]$`)
	if stringsRgx.MatchString(parameterType) {
		var param []string
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false
		}
		return validatePredStringArray(predicate, param, attribute, index)
	}
	// int
	intRgx := regexp.MustCompile(`u?int\d*$`)
	if intRgx.MatchString(parameterType) {
		var param *big.Int
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false
		}
		return validatePredBigInt(predicate, param, utils.MakeBigInt(attribute))
	}
	// int[]
	arrayIntRgx := regexp.MustCompile(`u?int\d*\[\d*]$`)
	if arrayIntRgx.MatchString(parameterType) {
		var param []*big.Int
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false
		}
		return validatePredBigIntArray(predicate, param, utils.MakeBigInt(attribute), index)
	}
	// byte[][]
	arrayByteRgx := regexp.MustCompile(`bytes\d*\[\d*]$`)
	if arrayByteRgx.MatchString(parameterType) {
		var param [][]byte
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false
		}
		return validatePredStringArray(predicate, utils.ByteArraysToHex(param), attribute, index)
	}
	if parameterType == "bytes" {
		var param []uint8
		if err = json.Unmarshal(rawParam, &param); err != nil {
			log.Debug(err)
			return false
		}
		return hex.EncodeToString(param) == attribute
	}
	log.Debug("parameter type not supported: ", parameterType)
	return false
}
