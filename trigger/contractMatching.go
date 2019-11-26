package trigger

import (
	"encoding/json"
	"fmt"
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"strings"
	"zoroaster/abidec"
	"zoroaster/utils"
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
	contractReturnValue, err := makeEthRpcCall(client, tg.ContractAdd, methodId, blockNo)
	if err != nil {
		log.Debug("rpc call failed: ", err)
		return false, nil, nil
	}
	log.Debug("result from call is -> ", contractReturnValue)

	var allValues []interface{}
	matchingValues := make([]string, len(tg.Outputs))
	for i, expectedOutput := range tg.Outputs {
		outputMatch, allVals := validateContractReturnValue(contractReturnValue, tg.ContractABI, tg.FunctionName, expectedOutput)
		matchingValues[i] = outputMatch
		if allVals != nil {
			allValues = allVals // always the same if not empty
		}
	}
	// a trigger matches if all the Outputs are a match (i.e. non-empty strings)
	for _, o := range matchingValues {
		if o == "" {
			return false, matchingValues, allValues
		}
	}
	return true, matchingValues, allValues
}

// returns:
// - in case of a match: a tuple (value_matched, []all_values casted as interface{})
// - in case of no match, error or whatever: a tuple ("", nil)
func validateContractReturnValue(
	rawData string,
	abi string,
	methodName string,
	expectedOutput Output) (string, []interface{}) {

	cnReturnType := expectedOutput.ReturnType
	cond := expectedOutput.Condition.(ConditionOutput)
	index := expectedOutput.Index
	rawData = strings.TrimPrefix(rawData, "0x")

	allValuesLs, err := abidec.DecodeParamsIntoList(rawData, abi, methodName)
	if err != nil {
		log.Debug(err)
		return "", nil
	}

	rawParam, returnType := getRawParamAndReturnType(cnReturnType, index, allValuesLs)
	if ValidateParam(rawParam, returnType, cond.Attribute, cond.Predicate, index) {
		return cond.Attribute, allValuesLs
	}
	return "", nil
}

// in case of multiple return values, like (int128, []uint8, string)
// we want to select the right param from the list, as well as the right type
func getRawParamAndReturnType(
	cnReturnType string,
	index *int,
	allParams []interface{}) ([]byte, string) {

	var rawParam interface{}
	if len(allParams) > 1 && index != nil && *index < len(allParams) {
		rawParam = allParams[*index]
		indexedType := strings.Split(cnReturnType, ",")[*index]
		cnReturnType = utils.RemoveCharacters(indexedType, "() ")
	} else {
		rawParam = allParams[0]
	}
	jsnBytes, _ := json.Marshal(rawParam)
	var rawParamOut json.RawMessage
	_ = json.Unmarshal(jsnBytes, &rawParamOut)
	return rawParamOut, cnReturnType
}

func makeEthRpcCall(client *ethrpc.EthRPC, cntAddress, data string, blockNumber int) (string, error) {

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
