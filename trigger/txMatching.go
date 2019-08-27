package trigger

import (
	"encoding/hex"
	"encoding/json"
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"zoroaster/utils"
)

func MatchTrigger(trigger *Trigger, block *ethrpc.Block) []*ZTransaction {
	ztxs := make([]*ZTransaction, 0)
	for i, tx := range block.Transactions {
		if validateTrigger(trigger, &tx) {
			// we discard errors here bc not every match will have input data
			var fnArgs *string
			fnArgsData, _ := decodeInputData(tx.Input, trigger.ContractABI)
			if fnArgsData != nil {
				fnArgsBytes, _ := json.Marshal(fnArgsData)
				fnArgsString := string(fnArgsBytes)
				fnArgs = &fnArgsString
			}
			fnName, _ := decodeInputMethod(&tx.Input, &trigger.ContractABI)

			zt := ZTransaction{
				BlockTimestamp: block.Timestamp,
				DecodedFnArgs:  fnArgs,
				DecodedFnName:  fnName,
				Tx:             &block.Transactions[i],
			}
			ztxs = append(ztxs, &zt)
		}
	}
	return ztxs
}

func validateTrigger(tg *Trigger, transaction *ethrpc.Transaction) bool {
	match := true
	for _, f := range tg.Filters {
		filterMatch := validateFilter(transaction, &f, tg.ContractAdd, &tg.ContractABI, tg.TriggerId)
		match = match && filterMatch // a Trigger matches if all filters match
	}
	return match
}

func validateFilter(ts *ethrpc.Transaction, f *Filter, cnt string, abi *string, tgId int) bool {
	cxtLog := log.WithFields(log.Fields{
		"trigger_id": tgId,
		"tx_hash":    ts.Hash,
	})
	defer func() {
		if r := recover(); r != nil {
			cxtLog.Errorf("panic: %s", r)
		}
	}()

	switch v := f.Condition.(type) {
	case ConditionFrom:
		return strings.ToLower(v.Attribute) == ts.From
	case ConditionTo:
		return strings.ToLower(v.Attribute) == ts.To
	case ConditionNonce:
		return validatePredInt(v.Predicate, ts.Nonce, v.Attribute)
	case ConditionValue:
		return validatePredBigInt(v.Predicate, &ts.Value, v.Attribute)
	case ConditionGas:
		return validatePredInt(v.Predicate, ts.Gas, v.Attribute)
	case ConditionGasPrice:
		return validatePredBigInt(v.Predicate, &ts.GasPrice, v.Attribute)
	case ConditionFunctionParam:
		// check transaction and ABI
		if !isValidContractAbi(abi, cnt, ts.To, tgId) {
			return false
		}
		// check FunctionName matches the transaction's method
		ok, err := matchesMethodName(abi, ts.Input, f.FunctionName)
		if err != nil {
			cxtLog.Debugf("cannot decode input method %v\n", err)
			return false
		}
		if !ok {
			return false // tx called a different method name
		}
		// decode input data
		decodedData, err := decodeInputDataToJsonMap(ts.Input, *abi)
		if err != nil {
			cxtLog.Debugf("cannot decode input data: %v\n", err)
			return false
		}
		// extract parameter
		rawParam, ok := decodedData[f.ParameterName]
		if !ok {
			cxtLog.Debugf("cannot find param %s in contract %s\n", f.ParameterName, ts.To)
			return false
		}
		// uint8
		if f.ParameterType == "uint8[]" {
			var param []uint8
			if err = json.Unmarshal(rawParam, &param); err != nil {
				cxtLog.Debug(err)
				return false
			}
			tgVal, err := strconv.Atoi(v.Attribute)
			if err == nil {
				return validatePredUIntArray(v.Predicate, param, tgVal, f.Index)
			}
		}
		// single address or string
		if f.ParameterType == "address" || f.ParameterType == "string" {
			var param string
			if err = json.Unmarshal(rawParam, &param); err != nil {
				cxtLog.Debug(err)
				return false
			}
			return strings.ToLower(param) == strings.ToLower(v.Attribute)
		}
		// bool
		if f.ParameterType == "bool" {
			var param bool
			if err = json.Unmarshal(rawParam, &param); err != nil {
				cxtLog.Debug(err)
				return false
			}
			return validatePredBool(v.Predicate, param, v.Attribute)
		}
		// address[]
		addressesRgx := regexp.MustCompile(`address\[\d*]$`)
		if addressesRgx.MatchString(f.ParameterType) {
			var param []string
			if err = json.Unmarshal(rawParam, &param); err != nil {
				cxtLog.Debug(err)
				return false
			}
			return validatePredStringArray(v.Predicate, param, v.Attribute, f.Index)
		}
		// string[]
		stringsRgx := regexp.MustCompile(`string\[\d*]$`)
		if stringsRgx.MatchString(f.ParameterType) {
			var param []string
			if err = json.Unmarshal(rawParam, &param); err != nil {
				cxtLog.Debug(err)
				return false
			}
			return validatePredStringArray(v.Predicate, param, v.Attribute, f.Index)
		}
		// int
		intRgx := regexp.MustCompile(`u?int\d*$`)
		if intRgx.MatchString(f.ParameterType) {
			var param *big.Int
			if err = json.Unmarshal(rawParam, &param); err != nil {
				cxtLog.Debug(err)
				return false
			}
			return validatePredBigInt(v.Predicate, param, utils.MakeBigInt(v.Attribute))
		}
		// int[]
		arrayIntRgx := regexp.MustCompile(`u?int\d*\[\d*]$`)
		if arrayIntRgx.MatchString(f.ParameterType) {
			var param []*big.Int
			if err = json.Unmarshal(rawParam, &param); err != nil {
				cxtLog.Debug(err)
				return false
			}
			return validatePredBigIntArray(v.Predicate, param, utils.MakeBigInt(v.Attribute), f.Index)
		}
		// byte[][]
		arrayByteRgx := regexp.MustCompile(`bytes\d*\[\d*]$`)
		if arrayByteRgx.MatchString(f.ParameterType) {
			var param [][]byte
			if err = json.Unmarshal(rawParam, &param); err != nil {
				cxtLog.Debug(err)
				return false
			}
			return validatePredStringArray(v.Predicate, utils.ByteArraysToHex(param), v.Attribute, f.Index)
		}
		if f.ParameterType == "bytes" {
			var param []uint8
			if err = json.Unmarshal(rawParam, &param); err != nil {
				cxtLog.Debug(err)
				return false
			}
			return hex.EncodeToString(param) == v.Attribute
		}
		cxtLog.Debug("parameter type not supported: ", f.ParameterType)
	case ConditionFunctionCalled:
		if !isValidContractAbi(abi, cnt, ts.To, tgId) {
			return false
		}
		ok, err := matchesMethodName(abi, ts.Input, f.FunctionName)
		if err != nil {
			cxtLog.Debugf("cannot decode input method %v\n", err)
			return false
		}
		return ok
	default:
		cxtLog.Debugf("filter not supported of type %T\n", f.Condition)
	}
	return false
}

func isValidContractAbi(abi *string, cntAddress string, txTo string, tgId int) bool {
	if len(*abi) == 0 {
		log.Debugf("(trigger %d) no ABI provided\n", tgId)
		return false
	}
	// make sure we are matching against the right transaction
	if strings.ToLower(cntAddress) != txTo {
		return false
	}
	return true
}

// check the trigger's FunctionName value matches the transaction's method
func matchesMethodName(abi *string, inputData string, funcName string) (bool, error) {
	methodName, err := decodeInputMethod(&inputData, abi)
	if err != nil {
		return false, err
	}
	return *methodName == funcName, nil
}
