package trigger

import (
	"encoding/json"
	"fmt"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/HAL-xyz/zoroaster/utils"
	log "github.com/sirupsen/logrus"
	"strings"
)

func MatchContract(tokenApi tokenapi.ITokenAPI, tg *Trigger, blockNo int) (*CnMatch, error) {

	tokenApiInputs := make([]tokenapi.Input, len(tg.Inputs))
	for i, e := range tg.Inputs {
		tokenApiInputs[i] = tokenapi.Input{
			ParameterType:  e.ParameterType,
			ParameterValue: e.ParameterValue,
		}
	}

	methodId, err := tokenApi.GetRPCCli().EncodeMethod(tg.FunctionName, tg.ContractABI, tokenApiInputs)
	if err != nil {
		return nil, fmt.Errorf("cannot encode method: %s", err)
	}
	rawData, err := tokenApi.GetRPCCli().MakeEthRpcCall(tg.ContractAdd, methodId, blockNo)
	if err != nil {
		return nil, fmt.Errorf("rpc call failed with error : %s", err)
	}

	//log.Debug("result from call is -> ", rawData)

	allValuesLs, err := utils.DecodeParamsIntoList(strings.TrimPrefix(rawData, "0x"), tg.ContractABI, tg.FunctionName)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	matchingValues := make([]string, 0)
	for _, expectedOutput := range tg.Outputs {
		if expectedOutput.ReturnIndex < len(allValuesLs) {
			rawParam := getRawParam(allValuesLs[expectedOutput.ReturnIndex])
			cond := expectedOutput.Condition.(ConditionOutput)
			yes, matchedValue := ValidateParam(rawParam, expectedOutput.ReturnType, expectedOutput.ReturnCurrency, cond.Attribute, cond.AttributeCurrency, cond.Predicate, expectedOutput.Index, expectedOutput.Component, tokenApi)
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
			AllValues:     utils.SprintfInterfaces(allValuesLs),
		}, nil
	}
	return nil, nil
}

func getRawParam(param interface{}) []byte {
	jsnBytes, _ := json.Marshal(param)
	var rawParamOut json.RawMessage
	err := json.Unmarshal(jsnBytes, &rawParamOut)
	if err != nil {
		log.Debug(err)
	}
	return rawParamOut
}
