package trigger

import (
	"encoding/json"
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"strings"
	"zoroaster/abidec"
)

func MatchTrigger(trigger *Trigger, block *ethrpc.Block) []*ZTransaction {
	ztxs := make([]*ZTransaction, 0)
	for i, tx := range block.Transactions {
		if validateTrigger(trigger, &tx) {
			// we discard errors here bc not every match will have input data
			var fnArgs *string
			fnArgsData, _ := abidec.DecodeInputData(tx.Input, trigger.ContractABI)
			if fnArgsData != nil {
				fnArgsBytes, _ := json.Marshal(fnArgsData)
				fnArgsString := string(fnArgsBytes)
				fnArgs = &fnArgsString
			}
			fnName, _ := abidec.DecodeInputMethod(&tx.Input, &trigger.ContractABI)

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
		filterMatch := validateFilter(transaction, &f, tg.ContractAdd, &tg.ContractABI, tg.TriggerUUID)
		match = match && filterMatch // a Trigger matches if all filters match
	}
	return match
}

func validateFilter(ts *ethrpc.Transaction, f *Filter, cnt string, abi *string, tgUUID string) bool {
	cxtLog := log.WithFields(log.Fields{
		"trigger_id": tgUUID,
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
		if !isValidContractAbi(abi, cnt, ts.To, tgUUID) {
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
		decodedData, err := abidec.DecodeInputDataToJsonMap(ts.Input, *abi)
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
		return ValidateParam(rawParam, f.ParameterType, v.Attribute, v.Predicate, f.Index)
	case ConditionFunctionCalled:
		if !isValidContractAbi(abi, cnt, ts.To, tgUUID) {
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

func isValidContractAbi(abi *string, cntAddress string, txTo string, tgUUID string) bool {
	if len(*abi) == 0 {
		log.Debugf("(trigger %s) no ABI provided\n", tgUUID)
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
	methodName, err := abidec.DecodeInputMethod(&inputData, abi)
	if err != nil {
		return false, err
	}
	return *methodName == funcName, nil
}
