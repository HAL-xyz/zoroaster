package trigger

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

func MatchTrigger(trigger *Trigger, block *ethrpc.Block) []*ZTransaction {
	ztxs := make([]*ZTransaction, 0)
	for i, tx := range block.Transactions {
		if ValidateTrigger(trigger, &tx) {
			// we discard errors here bc not every match will have input data
			var fnArgs *string
			fnArgsData, _ := DecodeInputData(tx.Input, trigger.ContractABI)
			if fnArgsData != nil {
				fnArgsBytes, _ := json.Marshal(fnArgsData)
				fnArgsString := string(fnArgsBytes)
				fnArgs = &fnArgsString
			}
			fnName, _ := DecodeInputMethod(&tx.Input, &trigger.ContractABI)

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

func ValidateTrigger(tg *Trigger, transaction *ethrpc.Transaction) bool {
	match := true
	for _, f := range tg.Filters {
		filterMatch := ValidateFilter(transaction, &f, tg.ContractAdd, &tg.ContractABI, tg.TriggerId)
		match = match && filterMatch // a Trigger matches if all filters match
	}
	return match
}

func ValidateFilter(ts *ethrpc.Transaction, f *Filter, cnt string, abi *string, tgId int) bool {
	cxtLog := log.WithFields(log.Fields{
		"trigger_id": tgId,
		"tx_hash":    ts.Hash,
	})
	defer func() {
		if r := recover(); r != nil {
			cxtLog.Debugf("panic: %s", r)
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
		// decode function arguments
		funcArgs, err := DecodeInputData(ts.Input, *abi)
		if err != nil {
			cxtLog.Debugf("cannot decode input data: %v\n", err)
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
		// extract params
		contractArg := funcArgs[f.ParameterName]
		if contractArg == nil {
			cxtLog.Debugf("cannot find param %s in contract %s\n", f.ParameterName, ts.To)
			return false
		}
		// single int/uint{40-256}
		if isValidBigInt(f.ParameterType) {
			return validatePredBigInt(v.Predicate, contractArg.(*big.Int), makeBigInt(v.Attribute))
		}
		// static array of int/uint{40-256}
		if isValidArray(f.ParameterType, stArrayIntRx, isValidBigInt) {
			ctVals := DecodeBigIntArray(contractArg)
			return validatePredBigIntArray(v.Predicate, ctVals, makeBigInt(v.Attribute), f.Index)
		}
		// dynamic array of int/uint{40-256}
		if isValidArray(f.ParameterType, dyArrayIntRx, isValidBigInt) {
			return validatePredBigIntArray(v.Predicate, contractArg.([]*big.Int), makeBigInt(v.Attribute), f.Index)
		}
		// single int/uint{8-32}
		if isValidInt(f.ParameterType) {
			tgVal, err := strconv.Atoi(v.Attribute)
			if err == nil {
				return validatePredInt(v.Predicate, int(contractArg.(int32)), tgVal)
			}
		}
		// static array of int/uint{8-32}
		if isValidArray(f.ParameterType, stArrayIntRx, isValidInt) {
			ctVals := DecodeIntArray(contractArg)
			tgVal, err := strconv.Atoi(v.Attribute)
			if err == nil {
				return validatePredIntArray(v.Predicate, ctVals, tgVal, f.Index)
			}
		}
		// dynamic array of int/uint{8-32}
		if isValidArray(f.ParameterType, dyArrayIntRx, isValidInt) {
			tgVal, err := strconv.Atoi(v.Attribute)
			if err == nil {
				return validatePredIntArray(v.Predicate, contractArg.([]int32), tgVal, f.Index)
			}
		}
		// static arrays of bytes1[] to bytes32[]
		if isValidArray(f.ParameterType, dyArrayBytesRx, isValidByte) {
			ctVals := Decode2DBytesArray(contractArg)
			return validatePredStringArray(v.Predicate, MultArrayToHex(ctVals), v.Attribute, f.Index)
		}
		// static arrays of address
		var addressArrayRx = regexp.MustCompile(`address\[\d+]`)
		if addressArrayRx.MatchString(f.ParameterType) {
			ctVals := DecodeAddressArray(contractArg)
			return validatePredStringArray(v.Predicate, ctVals, v.Attribute, f.Index)
		}
		// other types
		switch f.ParameterType {
		case "bool":
			return validatePredBool(v.Predicate, contractArg.(bool), v.Attribute)
		case "address":
			return contractArg.(common.Address).String() == common.HexToAddress(v.Attribute).String()
		case "address[]":
			byteAddresses := contractArg.([]common.Address)
			addresses := make([]string, len(byteAddresses))
			for i, a := range byteAddresses {
				addresses[i] = a.String()
			}
			return validatePredStringArray(v.Predicate, addresses, v.Attribute, f.Index)
		case "string[]":
			return validatePredStringArray(v.Predicate, contractArg.([]string), v.Attribute, f.Index)
		default:
			cxtLog.Debugf("parameter type not supported %s\n", f.ParameterType)
		}
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
	methodName, err := DecodeInputMethod(&inputData, abi)
	if err != nil {
		return false, err
	}
	return *methodName == funcName, nil
}
