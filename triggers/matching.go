package trigger

import (
	"github.com/INFURA/go-libs/jsonrpc_client"
	"github.com/ethereum/go-ethereum/common"
	"log"
	"math/big"
	"regexp"
	"strconv"
)

func MatchTrigger(trigger *Trigger, block *jsonrpc_client.Block) int {
	matchedTriggers := 0
	for _, trans := range block.Transactions {
		if ValidateTrigger(trigger, &trans) {
			matchedTriggers += 1
		}
	}
	return matchedTriggers
}

func ValidateTrigger(trigger *Trigger, transaction *jsonrpc_client.Transaction) bool {
	match := true
	for _, f := range trigger.Filters {
		filterMatch := ValidateFilter(transaction, &f, &trigger.ContractABI)
		match = match && filterMatch // a Trigger matches if all filters match
	}
	return match
}

// TODO return errors instead of logging?
func ValidateFilter(ts *jsonrpc_client.Transaction, f *Filter, abi *string) bool {

	switch v := f.Condition.(type) {
	case ConditionFrom:
		return v.Attribute == ts.From
	case ConditionTo:
		return v.Attribute == *ts.To
	case ConditionNonce:
		return validatePredInt(v.Predicate, ts.Nonce, v.Attribute)
	case ConditionValue:
		return validatePredBigInt(v.Predicate, ts.Value, v.Attribute)
	case ConditionGas:
		return validatePredInt(v.Predicate, ts.Gas, v.Attribute)
	case ConditionGasPrice:
		return validatePredBigInt(v.Predicate, ts.GasPrice, v.Attribute)
	case FunctionParamCondition:
		if len(*abi) == 0 {
			log.Println("No ABI provided")
			return false
		}
		// make sure we are matching against the right transaction
		if !(f.ToContract == *ts.To) {
			return false
		}
		// decode function arguments
		funcArgs, err := DecodeInputData(ts.Input, *abi)
		if err != nil {
			log.Println("Cannot decode input data: ", err)
			return false
		}
		// extract params
		contractArg := funcArgs[f.ParameterName]
		if contractArg == nil {
			log.Printf("Cannot find param %s in the contract", f.ParameterName)
			return false
		}

		// cast single int/uint{40-256}
		if isValidBigInt(f.ParameterType) {
			return validatePredBigInt(v.Predicate, contractArg.(*big.Int), makeBigInt(v.Attribute))
		}
		// cast static array of int/uint{40-256}
		if isValidBigIntArray(f.ParameterType) {
			ctVals := DecodeBigIntArray(contractArg)
			return validatePredBigIntArray(v.Predicate, ctVals, makeBigInt(v.Attribute))
		}
		// cast dynamic array of int/uint{40-256}
		if isValidDynamicBigIntArray(f.ParameterType) {
			return validatePredBigIntArray(v.Predicate, contractArg.([]*big.Int), makeBigInt(v.Attribute))
		}
		// cast single int/uint{8-32}
		if isValidInt(f.ParameterType) {
			tgVal, err := strconv.Atoi(v.Attribute)
			if err == nil {
				return validatePredInt(v.Predicate, int(contractArg.(int32)), tgVal)
			}
		}
		// cast dynamic array of int/uint{8-32}
		if isValidDynamicIntArray(f.ParameterType) {
			tgVal, err := strconv.Atoi(v.Attribute)
			if err == nil {
				return validatePredIntArray(v.Predicate, contractArg.([]int32), tgVal)
			}
		}
		// cast static arrays of bytes1[] to bytes32[]
		var bytesArrayRx = regexp.MustCompile(`bytes\d{1,2}\[]`)
		if bytesArrayRx.MatchString(f.ParameterType) {
			ctVals := Decode2DBytesArray(contractArg)
			return validatePredStringArray(v.Predicate, MultArrayToHex(ctVals), v.Attribute)
		}
		// cast static arrays of address
		var addressArrayRx = regexp.MustCompile(`address\[\d+]`)
		if addressArrayRx.MatchString(f.ParameterType) {
			ctVals := DecodeAddressArray(contractArg)
			return validatePredStringArray(v.Predicate, ctVals, v.Attribute)
		}
		// cast other types
		switch f.ParameterType {
		case "bool":
			return validatePredBool(v.Predicate, contractArg.(bool), v.Attribute)
		case "address":
			return contractArg == common.HexToAddress(v.Attribute)
		default:
			log.Println("Parameter type not supported", f.ParameterType)
		}
	default:
		log.Fatalf("filter not supported of type %T", f.Condition)
	}
	return false
}
