package matcher

import (
	"github.com/HAL-xyz/ethrpc"
	"github.com/HAL-xyz/zoroaster/db"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/HAL-xyz/zoroaster/trigger"
	"github.com/HAL-xyz/zoroaster/utils"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

func ContractMatcher(
	blocksChan chan *ethrpc.Block,
	matchesChan chan trigger.IMatch,
	idb db.IDB,
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
			log.Debug("logged one match with id ", m.MatchUUID)
			matchesChan <- m
		}
		if err = idb.SetLastBlockProcessed(block.Number, trigger.WaC); err != nil {
			log.Fatal(err)
		}
		log.Infof("CN: Processed %d triggers in %s from block %d", len(triggers), time.Since(start), block.Number)
	}
}

func matchContractsForBlock(blockNo, blockTimestamp int, blockHash string, idb db.IDB, tokenApi tokenapi.ITokenAPI) []*trigger.CnMatch {

	allTriggers, err := idb.LoadTriggersFromDB(trigger.WaC)
	if err != nil {
		log.Fatal(err)
	}

	const MAX = 3
	sem := make(chan int, MAX)
	mu := &sync.Mutex{}
	var wg sync.WaitGroup

	var cnMatches []*trigger.CnMatch
	var triggersWithErrorsUUIDs []string

	for _, trig := range allTriggers {
		sem <- 1
		wg.Add(1)
		go func(api tokenapi.ITokenAPI, tg *trigger.Trigger, bNo int) {
			defer wg.Done()
			match, err := trigger.MatchContract(api, tg, bNo)
			if err != nil {
				log.Debugf("WaC error for trigger %s: %s", tg.TriggerUUID, err.Error())
				mu.Lock()
				triggersWithErrorsUUIDs = append(triggersWithErrorsUUIDs, tg.TriggerUUID)
				mu.Unlock()
			}
			if match != nil {
				match.BlockTimestamp, match.BlockNumber, match.BlockHash = blockTimestamp, bNo, blockHash
				mu.Lock()
				cnMatches = append(cnMatches, match)
				mu.Unlock()
				log.Debugf("\tCN: Trigger %s matched on block %d\n", tg.TriggerUUID, bNo)
			}
			<-sem
		}(tokenApi, trig, blockNo)
	}
	wg.Wait()

	matchesToActUpon := getMatchesToActUpon(idb, cnMatches)

	updateStatusForMatchingTriggers(idb, cnMatches)
	updateStatusForNonMatchingTriggers(idb, cnMatches, allTriggers, triggersWithErrorsUUIDs)

	log.Infof("WAC potential matches: %d; errors: %d ", len(cnMatches), len(triggersWithErrorsUUIDs))
	return matchesToActUpon
}

// we only act on a match if it matches AND the triggered flag was set to false
func getMatchesToActUpon(idb db.IDB, cnMatches []*trigger.CnMatch) []*trigger.CnMatch {
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
func updateStatusForMatchingTriggers(idb db.IDB, matches []*trigger.CnMatch) {
	var matchingTriggersIds []string
	for _, m := range matches {
		matchingTriggersIds = append(matchingTriggersIds, m.Trigger.TriggerUUID)
	}
	idb.UpdateMatchingTriggers(matchingTriggersIds)
}

// set triggered flag to false for all non-matching 'true' triggers, but excluding triggers with errors
func updateStatusForNonMatchingTriggers(idb db.IDB, matches []*trigger.CnMatch, allTriggers []*trigger.Trigger, triggersWithErrors []string) {
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
