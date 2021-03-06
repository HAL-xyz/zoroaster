package trigger

import (
	"encoding/hex"
	"fmt"
	"github.com/HAL-xyz/ethrpc"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/HAL-xyz/zoroaster/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

func MatchEvent(tg *Trigger, logs []ethrpc.Log, txs []ethrpc.Transaction, tokenApi tokenapi.ITokenAPI) []*EventMatch {
	if tg.eventName() == "" {
		logrus.Debug("no valid Event Name found in trigger ", tg.TriggerUUID)
		return []*EventMatch{}
	}

	// loading ABIs is expensive, so we want to do it as little as possible
	abiObj, err := tg.getABIObj()
	if err != nil {
		logrus.Debug(err)
		return []*EventMatch{}
	}

	var eventMatches []*EventMatch
	for i, log := range logs {
		if !isRelevantLog(log.Address, tg.ContractAdd, tokenApi) {
			continue
		}
		if validateTriggerLog(&log, tg, tokenApi, abiObj) || validateEmittedEvent(&log, tg, abiObj) {
			tx := getTxByHash(log.TransactionHash, txs)
			if !validateBasicFiltersForEvent(tg, &tx) {
				continue
			}

			// make a new EventMatch
			decodedData, _ := decodeDataField(log.Data, tg.eventName(), &abiObj)
			topicsMap := getTopicsMap(&abiObj, tg.eventName(), &log)
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
	if !tg.hasBasicFilters() {
		return true
	}
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

func validateEmittedEvent(evLog *ethrpc.Log, tg *Trigger, abiObj abi.ABI) bool {

	for _, f := range tg.Filters {
		if f.FilterType == "CheckEventEmitted" {
			eventSignature, _ := getEventSignature(abiObj, tg.eventName())
			if evLog.Topics[0] == eventSignature {
				return true
			}
		}
	}
	return false
}

func validateTriggerLog(evLog *ethrpc.Log, tg *Trigger, tokenApi tokenapi.ITokenAPI, abiObj abi.ABI) bool {

	eventSignature, err := getEventSignature(abiObj, tg.eventName())
	if err != nil {
		logrus.Debug(err)
		return false
	}

	match := true
	if evLog.Topics[0] == eventSignature {
		for i := range tg.Filters {
			filterMatch, err := validateFilterLog(evLog, tg.Filters[i], &abiObj, tg.eventName(), tokenApi)
			if err != nil {
				logrus.Debug(err)
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
	filter Filter,
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
			add, ok := dv.(common.Address)
			if ok {
				filter.ParameterCurrency = utils.NormalizeAddress(add.String())
			}
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
		isValid, _ := ValidateParam(dataParam, filter.ParameterType, filter.ParameterCurrency, condition.Attribute, condition.AttributeCurrency, condition.Predicate, filter.Index, Component{}, tokenApi)
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

func getEventSignature(abiObj abi.ABI, eventName string) (string, error) {
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

func getTxByHash(hash string, txs []ethrpc.Transaction) ethrpc.Transaction {
	tx := ethrpc.Transaction{}
	for _, t := range txs {
		if t.Hash == hash {
			tx = t
			break
		}
	}
	return tx
}

// isRelevantLog decides if the inspected logAdd is to be considered a match, given the tgAdd.
// there are two cases:
// 1 - there is an exact match; in this case, it is obviously relevant
// 2 - the tgAdd is set to the keyword `all_erc20_tokens`; in this case, the inspected logAdd will be considered
// relevant as long as it is the address of any erc20 token we know.
func isRelevantLog(logAdd, tgAdd string, api tokenapi.ITokenAPI) bool {
	if tgAdd == "all_erc20_tokens" {
		_, ok := api.GetAllERC20TokensMap()[utils.NormalizeAddress(logAdd)]
		return ok
	}
	return utils.NormalizeAddress(logAdd) == utils.NormalizeAddress(tgAdd)
}

// global because it's expensive to compute
var intRgx = regexp.MustCompile(`u?int\d*$`)
