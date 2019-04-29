package trigger

import (
	"github.com/INFURA/go-libs/jsonrpc_client"
	"github.com/ethereum/go-ethereum/common"
	"log"
	"math/big"
)

func MatchTrigger(trigger Trigger, block *jsonrpc_client.Block) []*Trigger {
	var matchedTriggers []*Trigger
	for _, trans := range block.Transactions {
		_, ok := ValidateTrigger(trigger, trans)
		if ok {
			matchedTriggers = append(matchedTriggers, &trigger)
		}
	}
	return matchedTriggers
}

// TODO profile memory usage for this; perhaps take a *Trigger instead
func ValidateTrigger(trigger Trigger, transaction jsonrpc_client.Transaction) (*Trigger, bool) {
	match := true
	for _, f := range trigger.Filters {
		filterMatch := ValidateFilter(transaction, f, trigger.ContractABI)
		match = match && filterMatch // a Trigger matches if all filters match
	}
	if match {
		return &trigger, true
	} else {
		return nil, false
	}
}

// TODO return errors instead of logging?
// TODO implement all Conditions and FunctionParamConditions
func ValidateFilter(ts jsonrpc_client.Transaction, f Filter, abi string) bool {

	switch v := f.Condition.(type) {
	case ConditionTo:
		return v.Attribute == *ts.To
	case ConditionNonce:
		switch v.Predicate {
		case Eq:
			return v.Attribute == ts.Nonce
		case BiggerThan:
			return ts.Nonce > v.Attribute
		case SmallerThan:
			return ts.Nonce < v.Attribute
		}
	case FunctionParamCondition:

		// check smart contract TO
		if f.ToContract == *ts.To {

			// decode function arguments
			funcArgs, err := DecodeInputData(ts.Input, abi)
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

			// cast
			switch f.ParameterType {
			case "address":
				triggerAddress := common.HexToAddress(v.Attribute)
				if triggerAddress == contractArg {
					return true
				}
			case "uint256":
				contractValue := contractArg.(*big.Int)
				triggerValue := new(big.Int)
				triggerValue.SetString(v.Attribute, 10)
				switch v.Predicate {
				case Eq:
					return contractValue.Cmp(triggerValue) == 0
				case SmallerThan:
					return contractValue.Cmp(triggerValue) == -1
				case BiggerThan:
					return contractValue.Cmp(triggerValue) == 1
				}
			default:
				log.Println("Parameter type not supported", f.ParameterType)
			}
		}
	}
	return false
}
