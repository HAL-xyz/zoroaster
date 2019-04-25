package trigger

import (
	"github.com/INFURA/go-libs/jsonrpc_client"
	"github.com/ethereum/go-ethereum/common"
	"log"
)

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

			// extract params
			contractArg := funcArgs[f.ParameterName]
			if contractArg == nil {
				log.Println("Cannot find params in the function")
				return false
			}

			// cast
			if f.ParameterType == "Address" {
				triggerAddress := common.HexToAddress(v.Attribute)
				if triggerAddress == contractArg {
					return true
				}
			} else {
				log.Println("Parameter type not supported", f.ParameterType)
			}
		}
	}
	return false
}
