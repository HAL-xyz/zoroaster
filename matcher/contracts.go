package matcher

import (
	"github.com/HAL-xyz/ethrpc"
	"github.com/HAL-xyz/zoroaster/aws"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/HAL-xyz/zoroaster/trigger"
	"github.com/HAL-xyz/zoroaster/utils"
	log "github.com/sirupsen/logrus"
	"time"
)

func ContractMatcher(
	blocksChan chan *ethrpc.Block,
	matchesChan chan trigger.IMatch,
	idb aws.IDB,
	tokenApi tokenapi.ITokenAPI,
) {

	for {
		block := <-blocksChan
		tokenApi.GetRPCCli().ResetCounterAndLogStats(block.Number - 1)
		start := time.Now()

		triggers, err := idb.LoadTriggersFromDB(trigger.WaC)
		if err != nil {
			log.Fatal(err)
		}

		cnMatches := matchContractsForBlock(block.Number, block.Timestamp, block.Hash, idb, tokenApi)
		for _, m := range cnMatches {
			if err = idb.LogMatch(m); err != nil {
				log.Fatal(err)
			}
			log.Debug("\tlogged one match with id ", m.MatchUUID)
			matchesChan <- m
		}
		if err = idb.SetLastBlockProcessed(block.Number, trigger.WaC); err != nil {
			log.Fatal(err)
		}
		err = idb.LogAnalytics(trigger.WaC, block.Number, len(triggers), block.Timestamp, start, time.Now())
		if err != nil {
			log.Warn("cannot log analytics: ", err)
		}
	}
}

func matchContractsForBlock(blockNo, blockTimestamp int, blockHash string, idb aws.IDB, tokenApi tokenapi.ITokenAPI) []*trigger.CnMatch {

	start := time.Now()

	allTriggers, err := idb.LoadTriggersFromDB(trigger.WaC)
	if err != nil {
		log.Fatal(err)
	}

	var cnMatches []*trigger.CnMatch
	var triggersWithErrorsUUIDs []string
	for _, tg := range allTriggers {
		match, err := trigger.MatchContract(tokenApi, tg, blockNo)
		if err != nil {
			log.Infof("WaC error for trigger %s: %s", tg.TriggerUUID, err.Error())
			triggersWithErrorsUUIDs = append(triggersWithErrorsUUIDs, tg.TriggerUUID)
			continue
		}
		if match != nil {
			match.BlockTimestamp, match.BlockNumber, match.BlockHash = blockTimestamp, blockNo, blockHash
			cnMatches = append(cnMatches, match)
			log.Debugf("\tCN: Trigger %s matched on block %d\n", tg.TriggerUUID, blockNo)
		}
	}
	matchesToActUpon := getMatchesToActUpon(idb, cnMatches)

	updateStatusForMatchingTriggers(idb, cnMatches)
	updateStatusForNonMatchingTriggers(idb, cnMatches, allTriggers, triggersWithErrorsUUIDs)

	log.Infof("\tCN: Processed %d triggers in %s from block %d", len(allTriggers), time.Since(start), blockNo)
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

// set triggered flag to false for all non-matching 'true' triggers, but excluding triggers with errors
func updateStatusForNonMatchingTriggers(idb aws.IDB, matches []*trigger.CnMatch, allTriggers []*trigger.Trigger, triggersWithErrors []string) {
	setAll := make(map[string]struct{})
	setMatches := make(map[string]struct{})
	setErrors := make(map[string]struct{})

	for _, t := range allTriggers {
		setAll[t.TriggerUUID] = struct{}{}
	}
	for _, m := range matches {
		setMatches[m.Trigger.TriggerUUID] = struct{}{}
	}
	for _, e := range triggersWithErrors {
		setErrors[e] = struct{}{}
	}
	allTriggersWithoutErrors := utils.SetDifference(setAll, setErrors)
	nonMatchingTriggersIds := utils.SetDifference(allTriggersWithoutErrors, setMatches)
	idb.UpdateNonMatchingTriggers(utils.GetSliceFromIntSet(nonMatchingTriggersIds))
}
