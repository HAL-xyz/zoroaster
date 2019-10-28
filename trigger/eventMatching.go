package trigger

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/onrik/ethrpc"
	"github.com/sirupsen/logrus"
	"strings"
)

func MatchEvent(client IEthRpc, tg *Trigger, blockNo int) []*EventMatch {

	logs, err := getLogsForBlock(client, blockNo, tg.ContractAdd)
	if err != nil {
		logrus.Debug(err)
		return []*EventMatch{}
	}

	abiObj, err := abi.JSON(strings.NewReader(tg.ContractABI))
	if err != nil {
		logrus.Debug(err)
		return []*EventMatch{}
	}
	// EventName must be the same for every Filter, so we just get the first one
	if len(tg.Filters) < 1 {
		return []*EventMatch{}
	}
	eventName := tg.Filters[0].EventName

	namedTopicsMap := getNamedTopics(abiObj, eventName)
	eventSignature, err := getEventSignature(tg.ContractABI, eventName)
	if err != nil {
		logrus.Debug(err)
		return []*EventMatch{}
	}

	var eventMatches []*EventMatch
	for _, log := range logs {
		if validateTriggerLog(&log, tg, &abiObj, eventName, eventSignature, namedTopicsMap) {
			decodedData, _ := decodeDataField(log.Data, eventName, &abiObj)
			ev := EventMatch{
				tg:          tg,
				log:         &log,
				decodedData: decodedData,
			}
			eventMatches = append(eventMatches, &ev)
		}
	}
	return eventMatches
}

func validateTriggerLog(
	evLog *ethrpc.Log,
	tg *Trigger,
	abiObj *abi.ABI,
	eventName string, // TODO this should be moved inside the filter
	eventSignature string,
	namedTopicsMap map[int]string,
) bool {

	match := true
	if evLog.Topics[0] == eventSignature {
		for _, f := range tg.Filters {
			filterMatch := validateFilterLog(evLog, &f, abiObj, eventName, namedTopicsMap)
			match = match && filterMatch // a Trigger matches if all filters match
		}
	} else {
		return false
	}
	return match
}

func validateFilterLog(
	evLog *ethrpc.Log,
	filter *Filter,
	abiObj *abi.ABI,
	eventName string,
	namedTopicsMap map[int]string) bool {

	condition := filter.Condition.(ConditionEvent)

	// validate TOPICS field

	// a topicsMap is a map where
	// named_topic_{1,2,3} -> value
	// named_topic_0 is skipped being the event signature
	topicsMap := make(map[string]string)
	for i, t := range evLog.Topics {
		if i > 0 { // topic 0 is always the event signature
			topicsMap[namedTopicsMap[i]] = t
		}
	}

	param, ok := topicsMap[filter.ParameterName]
	if ok {
		jsn, err := json.Marshal(param)
		if err != nil {
			logrus.Debug(err)
			return false
		}
		return ValidateParam(jsn, filter.ParameterType, condition.Attribute, condition.Predicate, filter.Index)
	}

	// validate DATA field
	decodedData, err := decodeDataField(evLog.Data, eventName, abiObj)
	if err != nil {
		logrus.Debug(err)
		return false
	}

	dataParam, ok := decodedData[filter.ParameterName]
	if ok {
		jsn, err := json.Marshal(dataParam)
		if err != nil {
			logrus.Debug(err)
			return false
		}
		return ValidateParam(jsn, filter.ParameterType, condition.Attribute, condition.Predicate, filter.Index)
	}

	// parameter name not found in topics nor in data
	return false
}

func getLogsForBlock(client IEthRpc, blockNo int, address string) ([]ethrpc.Log, error) {
	fromBlock := fmt.Sprintf("0x%x", blockNo)
	toBlock := fmt.Sprintf("0x%x", blockNo)

	filter := ethrpc.FilterParams{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Address:   []string{address},
		// TODO: perhaps address should be an array, so I only make one RPC call?
		// this implies that MatchEvent is against []*Trigger and not a single *Trigger
	}
	logs, err := client.EthGetLogs(filter)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// a map where
// 1: first indexed input (topic 1)
// 2: second indexed input (topic 2)
// 3: third indexed input (topic 3)
// topic 0 is skipped being the event signature
func getNamedTopics(abiObj abi.ABI, eventName string) map[int]string {
	namedTopicsMap := make(map[int]string)
	myEvent := abiObj.Events[eventName]
	for i, input := range myEvent.Inputs {
		if input.Indexed {
			namedTopicsMap[i+1] = input.Name
		}
	}
	return namedTopicsMap
}

func getEventSignature(cntABI string, eventName string) (string, error) {

	// let's find out the Event specified by our trigger
	// this is equivalent to:
	// eventSignature := []byte("ItemSet(bytes32,bytes32)")
	// hash := crypto.Keccak256Hash(eventSignature)
	// but slightly easier because I don't have to make up the
	// string-form of the event signature ItemSet(bytes32,bytes32)

	abiObj, err := abi.JSON(strings.NewReader(cntABI))
	if err != nil {
		return "", err
	}

	var eventSignature string
	for _, event := range abiObj.Events {
		if event.Name == eventName {
			eventSignature = event.Id().Hex()
			return eventSignature, nil
		}
	}
	return "", fmt.Errorf("cannot find event %s\n", eventName)
}

func decodeDataField(rawData, eventName string, abiObj *abi.ABI) (map[string]interface{}, error) {

	decodedData, err := hex.DecodeString(rawData[2:])
	if err != nil {
		return nil, err
	}
	getMap := map[string]interface{}{}
	err = abiObj.UnpackIntoMap(getMap, eventName, decodedData)
	if err != nil {
		return nil, err
	}
	return getMap, nil
}
