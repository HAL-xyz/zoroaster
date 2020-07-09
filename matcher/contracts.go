package matcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/HAL-xyz/zoroaster/aws"
	"github.com/HAL-xyz/zoroaster/rpc"
	"github.com/HAL-xyz/zoroaster/trigger"
	"github.com/HAL-xyz/zoroaster/utils"
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func ContractMatcher(
	blocksChan chan *ethrpc.Block,
	matchesChan chan trigger.IMatch,
	getModifiedAccounts func(prevBlock, currBlock int, nodeURI string) ([]string, error),
	idb aws.IDB,
	client rpc.IEthRpc,
	useGetModAccount bool) {

	for {
		block := <-blocksChan
		client.ResetCounterAndLogStats(block.Number - 1)
		start := time.Now()
		log.Info("CN: new -> ", block.Number)

		triggers, err := idb.LoadTriggersFromDB(trigger.WaC)
		if err != nil {
			log.Fatal(err)
		}

		cnMatches := matchContractsForBlock(block.Number, block.Timestamp, block.Hash, getModifiedAccounts, idb, client, useGetModAccount)
		for _, m := range cnMatches {
			matchUUID, err := idb.LogMatch(*m)
			if err != nil {
				log.Fatal(err)
			}
			m.MatchUUID = matchUUID
			log.Debug("\tlogged one match with id ", matchUUID)
			matchesChan <- *m
		}
		err = idb.SetLastBlockProcessed(block.Number, trigger.WaC)
		if err != nil {
			log.Fatal(err)
		}
		err = idb.LogAnalytics(trigger.WaC, block.Number, len(triggers), block.Timestamp, start, time.Now())
		if err != nil {
			log.Warn("cannot log analytics: ", err)
		}
	}
}

func matchContractsForBlock(
	blockNo int,
	blockTimestamp int,
	blockHash string,
	getModAccounts func(prevBlock, currBlock int, nodeURI string) ([]string, error),
	idb aws.IDB,
	client rpc.IEthRpc,
	useGetModAccounts bool) []*trigger.CnMatch {

	start := time.Now()

	allTriggers, err := idb.LoadTriggersFromDB(trigger.WaC)
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("\tuseGetModAccount flag set to: ", useGetModAccounts)
	log.Debug("\ttriggers from IDB: ", len(allTriggers))

	var triggersToCheck []*trigger.Trigger

	if useGetModAccounts {
		log.Debug("\t...getting modified accounts...")
		modAccounts, modAccountErr := getModAccounts(blockNo-1, blockNo, client.URL())
		if modAccountErr != nil {
			log.Warn("\t", modAccountErr)
		}
		log.Debug("\tmodified accounts: ", len(modAccounts))

		// if getModifiedAccounts fails, we need to check all our WaC triggers
		if modAccountErr != nil {
			triggersToCheck = allTriggers
		} else {
			for i, t := range allTriggers {
				if utils.IsIn(strings.ToLower(t.ContractAdd), modAccounts) {
					triggersToCheck = append(triggersToCheck, allTriggers[i])
				}
			}
		}
	} else { // useGetModAccounts set to false
		triggersToCheck = allTriggers
	}

	log.Debug("\ttriggers to check: ", len(triggersToCheck))

	var cnMatches []*trigger.CnMatch
	for _, tg := range triggersToCheck {
		isMatch, matchedValues, allValues := trigger.MatchContract(client, tg, blockNo)
		if isMatch {
			match := &trigger.CnMatch{
				Trigger:        tg,
				MatchUUID:      "", // this will be set by Postgres once we persist
				BlockNumber:    blockNo,
				BlockHash:      blockHash,
				MatchedValues:  matchedValues,
				AllValues:      allValues,
				BlockTimestamp: blockTimestamp,
			}
			cnMatches = append(cnMatches, match)
			log.Debugf("\tCN: Trigger %s matched on block %d\n", tg.TriggerUUID, blockNo)
		}
	}
	matchesToActUpon := getMatchesToActUpon(idb, cnMatches)

	updateStatusForMatchingTriggers(idb, cnMatches)
	updateStatusForNonMatchingTriggers(idb, cnMatches, triggersToCheck)

	log.Infof("\tCN: Processed %d triggers in %s from block %d", len(triggersToCheck), time.Since(start), blockNo)
	return matchesToActUpon
}

// we only act on a match if it matches AND the triggered flag was set to false
func getMatchesToActUpon(idb aws.IDB, cnMatches []*trigger.CnMatch) []*trigger.CnMatch {
	var matchingTriggersUUIDs []string
	for _, m := range cnMatches {
		matchingTriggersUUIDs = append(matchingTriggersUUIDs, m.Trigger.TriggerUUID)
	}

	triggerUUIDsToActUpon, err := idb.GetSilentButMatchingTriggers(matchingTriggersUUIDs)
	if err != nil {
		log.Error(err)
	}

	var matchesToActUpon []*trigger.CnMatch
	for _, m := range cnMatches {
		if utils.IsIn(m.Trigger.TriggerUUID, triggerUUIDsToActUpon) {
			matchesToActUpon = append(matchesToActUpon, m)
		}
	}
	return matchesToActUpon
}

// set triggered flag to true for all matching 'false' triggers
func updateStatusForMatchingTriggers(idb aws.IDB, matches []*trigger.CnMatch) {
	var matchingTriggersIds []string
	for _, m := range matches {
		matchingTriggersIds = append(matchingTriggersIds, m.Trigger.TriggerUUID)
	}
	idb.UpdateMatchingTriggers(matchingTriggersIds)
}

// set triggered flag to false for all non-matching 'true' triggers
func updateStatusForNonMatchingTriggers(idb aws.IDB, matches []*trigger.CnMatch, allTriggers []*trigger.Trigger) {
	setAll := make(map[string]struct{})
	setMatches := make(map[string]struct{})

	for _, t := range allTriggers {
		setAll[t.TriggerUUID] = struct{}{}
	}
	for _, m := range matches {
		setMatches[m.Trigger.TriggerUUID] = struct{}{}
	}

	nonMatchingTriggersIds := utils.GetSliceFromIntSet(utils.SetDifference(setAll, setMatches))

	idb.UpdateNonMatchingTriggers(nonMatchingTriggersIds)
}

func GetModifiedAccounts(blockMinusOneNo, blockNo int, nodeURI string) ([]string, error) {

	type ethRequest struct {
		ID      int    `json:"id"`
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  []int  `json:"params"`
	}

	p := []int{blockMinusOneNo, blockNo}

	request := ethRequest{
		ID:      1,
		JSONRPC: "2.0",
		Method:  "debug_getModifiedAccountsByNumber",
		Params:  p,
	}

	body, err := json.Marshal(request)

	if err != nil {
		return nil, nil
	}

	response, err := http.Post(nodeURI, "application/json", bytes.NewBuffer(body))

	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("%s - request was: %s", err, string(body))
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("%s - request was: %s", err, string(body))
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("%s - request was: %s", err, string(body))
	}

	// result be like
	// {"jsonrpc":"2.0","id":1,"result":["0x31b93ca83b5ad17582e886c400667c6f698b8ccd",...]}

	type ethResponse struct {
		Result []string `json:"result"`
		Error  struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	var ethResp ethResponse

	err = json.Unmarshal(data, &ethResp)
	if err != nil {
		return nil, fmt.Errorf("%s - request was: %s", err, string(body))
	}

	if ethResp.Error.Message != "" {
		return nil, fmt.Errorf("%s - request was: %s", err, string(body))
	}

	return ethResp.Result, nil
}
