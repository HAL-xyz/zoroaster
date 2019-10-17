package matcher

import (
	"fmt"
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"time"
	"zoroaster/aws"
	"zoroaster/trigger"
	"zoroaster/utils"
)

func ContractMatcher(
	blocksChan chan *ethrpc.Block,
	matchesChan chan trigger.IMatch,
	getModifiedAccounts func(prevBlock, currBlock int, nodeURI string) []string,
	idb aws.IDB,
	client *ethrpc.EthRPC) {

	for {
		block := <-blocksChan
		log.Info("CN: new -> ", block.Number)

		cnMatches := matchContractsForBlock(block.Number, block.Timestamp, block.Hash, getModifiedAccounts, idb, client)
		for _, m := range cnMatches {
			matchId := idb.LogMatch(*m)
			m.MatchUUID = matchId
			log.Debug("\tlogged one match with id ", matchId)
			matchesChan <- m
		}
		idb.SetLastBlockProcessed(block.Number, "wac")
	}
}

func matchContractsForBlock(
	blockNo int,
	blockTimestamp int,
	blockHash string,
	getModAccounts func(prevBlock, currBlock int, nodeURI string) []string,
	idb aws.IDB,
	client *ethrpc.EthRPC) []*trigger.CnMatch {

	start := time.Now()

	log.Debug("\t...getting modified accounts...")
	modAccounts := getModAccounts(blockNo-1, blockNo, client.URL())
	for len(modAccounts) == 0 {
		log.Warn("\tdidn't get any modified accounts, retrying in a few seconds")
		time.Sleep(10 * time.Second)
		modAccounts = getModAccounts(blockNo-1, blockNo, client.URL())
	}
	log.Debug("\tmodified accounts: ", len(modAccounts))

	allTriggers, err := idb.LoadTriggersFromDB("WatchContracts")
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("\ttriggers from IDB: ", len(allTriggers))

	var triggersToCheck []*trigger.Trigger
	for i, t := range allTriggers {
		if utils.IsIn(t.ContractAdd, modAccounts) {
			triggersToCheck = append(triggersToCheck, allTriggers[i])
		}
	}
	log.Debug("\ttriggers pointing to a modified account: ", len(triggersToCheck))

	var cnMatches []*trigger.CnMatch
	for _, tg := range triggersToCheck {
		isMatch, matchedValues, allValues := trigger.MatchContract(client, tg, blockNo)
		if isMatch {
			match := &trigger.CnMatch{
				Trigger:        tg,
				MatchUUID:      "uuid", // this will be set by Postgres once we persist
				BlockNo:        blockNo,
				BlockHash:      blockHash,
				MatchedValues:  fmt.Sprint(matchedValues),
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
		matchingTriggersUUIDs = append(matchingTriggersUUIDs, m.MatchUUID)
	}

	triggerUUIDsToActUpon := idb.GetSilentButMatchingTriggers(matchingTriggersUUIDs)

	var matchesToActUpon []*trigger.CnMatch
	for _, m := range cnMatches {
		if utils.IsIn(m.MatchUUID, triggerUUIDsToActUpon) {
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
