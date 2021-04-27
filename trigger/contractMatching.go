package trigger

import (
	"fmt"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/HAL-xyz/zoroaster/utils"
)

func MatchContract(api tokenapi.ITokenAPI, tg *Trigger, blockNo int) (*CnMatch, error) {

	result, err := api.EthCall(tg.ContractAdd, tg.FunctionName, tg.ContractABI, blockNo, tg.CallArgs()...)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	matchingValues := make([]string, 0)
	for _, expectedOutput := range tg.Outputs {
		if expectedOutput.ReturnIndex < len(result) {
			cond := expectedOutput.Condition.(ConditionOutput)
			yes, matchedValue := ValidateParam(result[expectedOutput.ReturnIndex], expectedOutput.ReturnType, expectedOutput.ReturnCurrency, cond.Attribute, cond.AttributeCurrency, cond.Predicate, expectedOutput.Index, expectedOutput.Component, api)
			if yes {
				matchingValues = append(matchingValues, fmt.Sprintf("%v", matchedValue))
			}
		}
	}

	if len(matchingValues) == len(tg.Outputs) { // all filters match
		return &CnMatch{
			MatchUUID:     "", // this will be set by Postgres once we persist
			Trigger:       tg,
			MatchedValues: matchingValues,
			AllValues:     utils.SprintfInterfaces(result),
		}, nil
	}
	return nil, nil
}
