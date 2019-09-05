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
	blockNo int) (isMatch bool, allMatchingValues []string, allReturnedValues string) {

	methodId, err := encodeMethod(tg.MethodName, tg.ContractABI, tg.Inputs)
	if err != nil {
		log.Debug("cannot encode method: ", err)
		return false, nil, ""
	}
	contractReturnValue, err := makeEthRpcCall(client, tg.ContractAdd, methodId, blockNo)
	if err != nil {
		log.Debug("rpc call failed: ", err)
		return false, nil, ""
	}
	log.Debug("result from call is -> ", contractReturnValue)

	var allValues string
	matchingValues := make([]string, len(tg.Outputs))
	for i, expectedOutput := range tg.Outputs {
		outputMatch, allVals := validateContractReturnValue(contractReturnValue, tg.ContractABI, tg.MethodName, expectedOutput)
		matchingValues[i] = outputMatch
		if allVals != "" {
			allValues = allVals
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
// - in case of a match: a tuple (value_matched, all_values)
// - in case of no match, error or whatever: a tuple ("", "")
func validateContractReturnValue(
	contractValue string,
	abi string,
	methodName string,
	expectedOutput Output) (string, string) {

	cnReturnType := expectedOutput.ReturnType
	cond := expectedOutput.Condition.(ConditionOutput)
	index := expectedOutput.Index

	contractValue = strings.TrimPrefix(contractValue, "0x")

	rawJsParamsMap, err := abidec.DecodeParamsToJsonMap(contractValue, abi, methodName)
	if err != nil {
		log.Debug(err)
		return "", ""
	}
	rawJsParamsList := utils.GetValuesFromMap(rawJsParamsMap)

	// in case of multiple return values, like (int128, []uint8, string)
	// we want to select the right param from the list, as well as the right type
	var rawParam json.RawMessage
	if len(rawJsParamsMap) > 1 && index != nil && *index < len(rawJsParamsMap) {
		rawParam = rawJsParamsList[*index]
		allTypes := strings.Split(cnReturnType, ",")
		indexedType := allTypes[*index]
		cnReturnType = utils.RemoveCharacters(indexedType, "() ")
	} else {
		rawParam = rawJsParamsList[0]
	}
	// Yes this whole templating thing is beyond shit but hey.
	if ValidateParam(rawParam, cnReturnType, cond.Attribute, cond.Predicate, index) {
		out := make([]string, len(rawJsParamsList))
		for i, elem := range rawJsParamsList {
			out[i] = fmt.Sprintf("%s"+"#END#", elem)
		}
		// remove the last #END# from the string.
		s := utils.Reverse(fmt.Sprintf("%s", out))
		s = strings.Replace(s, "#DNE#", "", 1)
		s = utils.Reverse(s)
		return cond.Attribute, s
	}
	return "", ""
}

func makeEthRpcCall(client *ethrpc.EthRPC, cntAddress, data string, blockNumber int) (string, error) {

	params := ethrpc.T{
		To:   cntAddress,
		From: cntAddress,
		Data: data,
	}

	hexBlockNo := fmt.Sprintf("0x%x", blockNumber)

	ret, err := client.EthCall(params, hexBlockNo)
	if err != nil {
		return "", err
	}
	return ret, nil
}
