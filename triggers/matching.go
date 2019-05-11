package trigger

import (
	"github.com/INFURA/go-libs/jsonrpc_client"
	"github.com/ethereum/go-ethereum/common"
	"log"
	"math/big"
	"regexp"
)

func MatchTrigger(trigger *Trigger, block *jsonrpc_client.Block) []*Trigger {
	var matchedTriggers []*Trigger
	for _, trans := range block.Transactions {
		_, ok := ValidateTrigger(trigger, &trans)
		if ok {
			matchedTriggers = append(matchedTriggers, trigger)
		}
	}
	return matchedTriggers
}

func ValidateTrigger(trigger *Trigger, transaction *jsonrpc_client.Transaction) (*Trigger, bool) {
	match := true
	for _, f := range trigger.Filters {
		filterMatch := ValidateFilter(transaction, &f, &trigger.ContractABI)
		match = match && filterMatch // a Trigger matches if all filters match
	}
	if match {
		return trigger, true
	} else {
		return nil, false
	}
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
		// when analyzing the function parameters, first thing we want
		// to make sure we are matching against the right transaction
		if !(f.ToContract == *ts.To) {
			return false
		}
		// decode function arguments
		funcArgs, err := DecodeInputData(ts.Input, *abi)
		if err != nil {
			log.Println("Cannot decode input data: ", err)
			return false
		}
		// extract param
		contractArg := funcArgs[f.ParameterName]
		if contractArg == nil {
			log.Printf("Cannot find param %s in the contract", f.ParameterName)
			return false
		}

		// cast single int/uint{64-256}
		if isValidBigInt(f.ParameterType) {
			ctVal := contractArg.(*big.Int)
			return validatePredBigInt(v.Predicate, ctVal, makeBigInt(v.Attribute))
		}
		// cast static array of int/uint{64-256}
		if isValidBigIntArray(f.ParameterType) {
			ctVals := DecodeBigIntArray(contractArg)
			return validatePredBigIntArray(v.Predicate, ctVals, makeBigInt(v.Attribute))
		}
		// cast static arrays of bytes1[] to bytes32[]
		var bytesArrayRx = regexp.MustCompile(`bytes\d{1,2}\[]`)
		if bytesArrayRx.MatchString(f.ParameterType) {
			arg := Decode2DBytesArray(contractArg)
			return validatePredStringArray(v.Predicate, MultArrayToHex(arg), v.Attribute)
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
		case "uint256[]": // TODO support any big int dynamic array
			return validatePredBigIntArray(v.Predicate, contractArg.([]*big.Int), makeBigInt(v.Attribute))
		default:
			log.Println("Parameter type not supported", f.ParameterType)
		}
	default:
		log.Fatalf("filter not supported of type %T", f.Condition)
	}
	return false
}
