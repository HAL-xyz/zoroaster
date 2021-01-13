package trigger

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/HAL-xyz/ethrpc"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/HAL-xyz/zoroaster/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

func MatchEvent(tg *Trigger, logs []ethrpc.Log, tokenApi tokenapi.ITokenAPI) []*EventMatch {
	abiObj, err := abi.JSON(strings.NewReader(tg.ContractABI))
	if err != nil {
		logrus.Debug(err)
		return []*EventMatch{}
	}

	// EventName must be the same for every Filter, so we just get the first one
	var eventName string
	for _, f := range tg.Filters {
		if f.FilterType == "CheckEventParameter" || f.FilterType == "CheckEventEmitted" {
			eventName = f.EventName
			break
		}
	}
	if eventName == "" {
		logrus.Debug("no valid Event Name found in trigger ", tg.TriggerUUID)
		return []*EventMatch{}
	}

	var eventMatches []*EventMatch
	for i, log := range logs {
		if utils.NormalizeAddress(log.Address) != utils.NormalizeAddress(tg.ContractAdd) {
			continue
		}
		if validateTriggerLog(&log, tg, &abiObj, eventName, tokenApi) || validateEmittedEvent(&log, tg, eventName) {
			tx, err := tokenApi.GetRPCCli().EthGetTransactionByHash(log.TransactionHash)
			if err != nil {
				logrus.Error("cannot fetch tx by hash: ", err)
				continue
			}
			if tg.hasBasicFilters() {
				if !validateBasicFiltersForEvent(tg, tx) {
					continue
				}
			}
			decodedData, _ := decodeDataField(log.Data, eventName, &abiObj)
			topicsMap := getTopicsMap(&abiObj, eventName, &log)
			ev := EventMatch{
				Tg:          tg,
				Log:         &logs[i],
				EventParams: makeEventParams(decodedData, topicsMap),
				TxTo:        tx.To,
				TxFrom:      tx.From,
			}
			eventMatches = append(eventMatches, &ev)
		}
	}
	return eventMatches
}

func validateBasicFiltersForEvent(tg *Trigger, tx *ethrpc.Transaction) bool {
	allFiltersMatch := true
	for _, f := range tg.Filters {
		switch v := f.Condition.(type) {
		case ConditionFrom:
			allFiltersMatch = allFiltersMatch && (utils.NormalizeAddress(v.Attribute) == utils.NormalizeAddress(tx.From))
		case ConditionTo:
			allFiltersMatch = allFiltersMatch && (utils.NormalizeAddress(v.Attribute) == utils.NormalizeAddress(tx.To))
		default:
			continue
		}
	}
	return allFiltersMatch
}

func validateEmittedEvent(evLog *ethrpc.Log, tg *Trigger, eventName string) bool {
	for _, f := range tg.Filters {
		if f.FilterType == "CheckEventEmitted" {
			eventSignature, _ := getEventSignature(tg.ContractABI, eventName)
			if evLog.Topics[0] == eventSignature {
				return true
			}
		}
	}
	return false
}

func validateTriggerLog(evLog *ethrpc.Log, tg *Trigger, abiObj *abi.ABI, eventName string, tokenApi tokenapi.ITokenAPI) bool {
	cxtLog := logrus.WithFields(logrus.Fields{
		"trigger_id": tg.TriggerUUID,
		"tx_hash":    evLog.TransactionHash,
		"block_no":   evLog.BlockNumber,
	})

	eventSignature, err := getEventSignature(tg.ContractABI, eventName)
	if err != nil {
		cxtLog.Debug(err)
		return false
	}

	match := true
	if evLog.Topics[0] == eventSignature {
		for _, f := range tg.Filters {
			filterMatch, err := validateFilterLog(evLog, &f, abiObj, eventName, tokenApi)
			if err != nil {
				cxtLog.Debug(err)
			}
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
	tokenApi tokenapi.ITokenAPI) (bool, error) {

	condition, ok := filter.Condition.(ConditionEvent)
	if !ok {
		// filter isn't of type CheckEventParameter, ignore it
		return true, nil
	}

	topicsMap := getTopicsMap(abiObj, eventName, evLog)
	decodedData, err := decodeDataField(evLog.Data, eventName, abiObj)

	// handle implicit currencies, where ParameterCurrency is an Event field name
	if filter.ParameterCurrency != "" && !strings.HasPrefix(filter.ParameterCurrency, "0x") {
		tv, ok := topicsMap[filter.ParameterCurrency]
		if ok {
			filter.ParameterCurrency = utils.NormalizeAddress(tv)
		}
		dv, ok := decodedData[filter.ParameterCurrency]
		if ok {
			filter.ParameterCurrency = utils.NormalizeAddress(fmt.Sprintf("%v", dv))
		}
	}

	// validate TOPICS
	param, ok := topicsMap[filter.ParameterName]
	if ok {
		isValid, _ := ValidateTopicParam(param, filter.ParameterType, filter.ParameterCurrency, condition, tokenApi)
		return isValid, nil
	}

	// validate DATA field
	if err != nil {
		return false, err
	}
	dataParam, ok := decodedData[filter.ParameterName]
	if ok {
		jsn, err := json.Marshal(dataParam)
		if err != nil {
			return false, err
		}
		isValid, _ := ValidateParam(jsn, filter.ParameterType, filter.ParameterCurrency, condition.Attribute, condition.AttributeCurrency, condition.Predicate, filter.Index, Component{}, tokenApi)
		return isValid, nil
	}

	return false, nil // parameter name not found in topics nor in data
}

// a topicsMap is a map where
// topic_name_1 -> value
// topic_name_2 -> value
// topic_name_3 -> value
// topic_name_0 is skipped being the event signature
//
// this is needed bc evLog.Topics is simply a []string of the topic values,
// and we want to produce a map (topic_name -> value) looping through the
// events Inputs names (i.e. the variables of the Event struct) and linking
// each name to the value in Topics
func getTopicsMap(abiObj *abi.ABI, eventName string, evLog *ethrpc.Log) map[string]string {
	finalMap := make(map[string]string)
	myEvent := abiObj.Events[eventName]
	intRgx := regexp.MustCompile(`u?int\d*$`)

	var i = 1 // topic_name_0 is the event signature so we start from 1
	for _, input := range myEvent.Inputs {
		if input.Indexed {
			if intRgx.MatchString(input.Type.String()) {
				finalMap[input.Name] = utils.MakeBigIntFromHex(evLog.Topics[i]).String()
			} else {
				finalMap[input.Name] = evLog.Topics[i]
			}
			i += 1
		}
	}
	return finalMap
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
	for _, event := range abiObj.Events {
		if event.Name == eventName {
			return event.ID.Hex(), nil
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

func makeEventParams(data map[string]interface{}, topics map[string]string) map[string]interface{} {

	paramsMap := make(map[string]interface{}, len(data)+len(topics))
	for k, v := range topics {
		paramsMap[k] = utils.NormalizeAddress(v)
	}
	for k, v := range data {
		paramsMap[k] = utils.SprintfInterfaces([]interface{}{v})[0]
	}
	return paramsMap
}
