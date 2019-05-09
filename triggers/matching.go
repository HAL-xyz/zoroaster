package trigger

import (
	"fmt"
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
		// check smart contract TO
		if f.ToContract == *ts.To {
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
			// cast static arrays of bytes1[] to bytes32[]
			var bytesArrayRx = regexp.MustCompile(`bytes\d{1,2}\[\]`)
			if bytesArrayRx.MatchString(f.ParameterType) {
				arg := Decode2DBytesArray(contractArg)
				contractValues := MultArrayToHex(arg)
				return validatePredStringArray(v.Predicate, contractValues, v.Attribute)
			}
			// cast static arrays of address
			var addressArrayRx = regexp.MustCompile(`address\[\d+]`)
			if addressArrayRx.MatchString(f.ParameterType) {
				contractValues := DecodeAddressArray(contractArg)
				return validatePredStringArray(v.Predicate, contractValues, v.Attribute)
			}
			// cast static arrays of uint256[8]
			var uintArrayRx = regexp.MustCompile(`uint256\[\d+\]`)
			if uintArrayRx.MatchString(f.ParameterType) {
				contractValues := DecodeUint256Array(contractArg)
				triggerValue := new(big.Int)
				triggerValue.SetString(v.Attribute, 10)
				return validatePredBigIntArray(v.Predicate, contractValues, triggerValue)
			}
			// cast other types
			switch f.ParameterType {
			case "bool":
				return validatePredBool(v.Predicate, contractArg.(bool), v.Attribute)
			case "address":
				triggerAddress := common.HexToAddress(v.Attribute)
				if triggerAddress == contractArg {
					return true
				}
			case "uint256":
				contractValue := contractArg.(*big.Int)
				triggerValue := new(big.Int)
				triggerValue.SetString(v.Attribute, 10)
				fmt.Println(contractValue)
				return validatePredBigInt(v.Predicate, contractValue, triggerValue)
			case "uint256[]":
				contractValues := contractArg.([]*big.Int)
				triggerValue := new(big.Int)
				triggerValue.SetString(v.Attribute, 10)
				return validatePredBigIntArray(v.Predicate, contractValues, triggerValue)
			default:
				log.Println("Parameter type not supported", f.ParameterType)
			}
		}
	default:
		log.Fatalf("filter not supported of type %T", f.Condition)
	}
	return false
}
