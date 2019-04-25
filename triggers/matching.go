package trigger

import (
	"github.com/INFURA/go-libs/jsonrpc_client"
	"github.com/ethereum/go-ethereum/common"
	"log"
)

// TODO ABI
// TODO AND logic
// TODO return type
// TODO tests
func process(trigger Trigger, block jsonrpc_client.Block) {
	for _, ts := range block.Transactions {
		for _, f := range trigger.Filters {
			ValidateFilter(ts, f, trigger.ContractABI)
		}
	}
}

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
	// TODO extract to func?
	// TODO use typed errors?
	case FunctionParamCondition:
		// check smart contract TO
		if f.ToContract == *ts.To {

			// decode function arguments
			funcArgs := DecodeInputData(ts.Input, abi)

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
				log.Print("Parameter type not supported", f.ParameterType)
			}
		}
	}
	return false
}
