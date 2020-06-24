package trigger

import (
	"encoding/json"
	"fmt"
	"github.com/HAL-xyz/zoroaster/abidec"
	"github.com/HAL-xyz/zoroaster/utils"
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"strings"
)

func MatchContract(
	client *ethrpc.EthRPC,
	tg *Trigger,
	blockNo int) (isMatch bool, allMatchingValues []string, allReturnedValues []interface{}) {

	methodId, err := encodeMethod(tg.FunctionName, tg.ContractABI, tg.Inputs)
	if err != nil {
		log.Debugf("trigger %s: cannot encode method: %s", tg.TriggerUUID, err)
		return false, nil, nil
	}
	rawData, err := MakeEthRpcCall(client, tg.ContractAdd, methodId, blockNo)
	if err != nil {
		log.Debug("rpc call failed: ", err)
		return false, nil, nil
	}

	//log.Debug("result from call is -> ", rawData)

	allValuesLs, err := abidec.DecodeParamsIntoList(strings.TrimPrefix(rawData, "0x"), tg.ContractABI, tg.FunctionName)
	if err != nil {
		log.Debug(err)
	}

	matchingValues := make([]string, 0)
	for _, expectedOutput := range tg.Outputs {
		if expectedOutput.ReturnIndex < len(allValuesLs) {
			rawParam := getRawParam(allValuesLs[expectedOutput.ReturnIndex])
			cond := expectedOutput.Condition.(ConditionOutput)
			yes, matchedValue := ValidateParam(rawParam, expectedOutput.ReturnType, cond.Attribute, cond.Predicate, expectedOutput.Index)
			if yes {
				matchingValues = append(matchingValues, fmt.Sprintf("%v", matchedValue))
			}
		}
	}
	return len(matchingValues) == len(tg.Outputs), matchingValues, utils.SprintfInterfaces(allValuesLs)
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

func MakeEthRpcCall(client *ethrpc.EthRPC, cntAddress, data string, blockNumber int) (string, error) {

	params := ethrpc.T{
		To: cntAddress,
		// the from field is a random hardcoded address
		// because the ethrpc library for now doesn't support
		// an empty from field :(
		From: "0x2e34c46ad2f08a66bc9ff2e9fe5918590551e958",
		Data: data,
	}

	hexBlockNo := fmt.Sprintf("0x%x", blockNumber)

	ret, err := client.EthCall(params, hexBlockNo)
	if err != nil {
		return "", err
	}
	return ret, nil
}
