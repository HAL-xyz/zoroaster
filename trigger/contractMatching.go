package trigger

import (
	"encoding/json"
	"fmt"
	"github.com/HAL-xyz/ethrpc"
	"github.com/HAL-xyz/zoroaster/abidec"
	"github.com/HAL-xyz/zoroaster/rpc"
	"github.com/HAL-xyz/zoroaster/utils"
	log "github.com/sirupsen/logrus"
	"strings"
)

func MatchContract(client rpc.IEthRpc, tg *Trigger, blockNo int) (*CnMatch, error) {

	methodId, err := EncodeMethod(tg.FunctionName, tg.ContractABI, tg.Inputs)
	if err != nil {
		return nil, fmt.Errorf("trigger %s: cannot encode method: %s", tg.TriggerUUID, err)
	}
	rawData, err := MakeEthRpcCall(client, tg.ContractAdd, methodId, blockNo)
	if err != nil {
		return nil, fmt.Errorf("rpc call failed for trigger %s with error : %s", tg.TriggerUUID, err)
	}

	//log.Debug("result from call is -> ", rawData)

	allValuesLs, err := abidec.DecodeParamsIntoList(strings.TrimPrefix(rawData, "0x"), tg.ContractABI, tg.FunctionName)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	matchingValues := make([]string, 0)
	for _, expectedOutput := range tg.Outputs {
		if expectedOutput.ReturnIndex < len(allValuesLs) {
			rawParam := getRawParam(allValuesLs[expectedOutput.ReturnIndex])
			cond := expectedOutput.Condition.(ConditionOutput)
			yes, matchedValue := ValidateParam(rawParam, expectedOutput.ReturnType, cond.Attribute, cond.Predicate, expectedOutput.Index, expectedOutput.Component)
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

func MakeEthRpcCall(client rpc.IEthRpc, cntAddress, data string, blockNumber int) (string, error) {

	params := ethrpc.T{
		To: cntAddress,
		// the from field is a random hardcoded address
		// because the ethrpc library for now doesn't support
		// an empty from field :(
		From: "0x2e34c46ad2f08a66bc9ff2e9fe5918590551e958",
		Data: data,
	}

	hexBlockNo := fmt.Sprintf("0x%x", blockNumber)

	return client.EthCall(params, hexBlockNo)
}
