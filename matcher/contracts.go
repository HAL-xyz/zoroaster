package matcher

import (
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"time"
	"zoroaster/aws"
	"zoroaster/trigger"
)

func ContractMatcher(
	blocksChan chan *ethrpc.Block,
	matchesChan chan trigger.IMatch,
	getModifiedAccounts func(prevBlock, currBlock int) []string,
	idb aws.IDB,
	client *ethrpc.EthRPC) {

	for {
		block := <-blocksChan
		log.Info("CN: new -> ", block.Number)

		cnMatches := matchContractsForBlock(block.Number, block.Timestamp, getModifiedAccounts, idb, client)
		for _, m := range cnMatches {
			matchId := idb.LogCnMatch(*m)
			m.MatchId = matchId
			log.Debug("\tlogged one match with id ", matchId)
			matchesChan <- m
		}
		idb.SetLastBlockProcessed(block.Number, "wac")
	}
}

func matchContractsForBlock(
	blockNo int,
	blockTimestamp int,
	getModAccounts func(prevBlock, currBlock int) []string,
	idb aws.IDB,
	client *ethrpc.EthRPC) []*trigger.CnMatch {

	start := time.Now()

	log.Debug("\t...getting modified accounts...")
	modAccounts := getModAccounts(blockNo-1, blockNo)
	for len(modAccounts) == 0 {
		log.Warn("\tdidn't get any modified accounts, retrying in a few seconds")
		time.Sleep(10 * time.Second)
		modAccounts = getModAccounts(blockNo-1, blockNo)
	}
	log.Debug("\tmodified accounts: ", len(modAccounts))

	triggers, err := idb.LoadTriggersFromDB("WatchContracts")
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("\ttriggers from IDB: ", len(triggers))

	var wacTriggers []*trigger.Trigger
	for i, t := range triggers {
		if isIn(t.ContractAdd, modAccounts) {
			wacTriggers = append(wacTriggers, triggers[i])
		}
	}
	log.Debug("\ttriggers pointing to a modified account: ", len(wacTriggers))

	var cnMatches []*trigger.CnMatch
	for _, tg := range wacTriggers {
		matchedValue, allValues := trigger.MatchContract(client, tg, blockNo)
		if matchedValue != "" {
			match := &trigger.CnMatch{
				MatchId:        0,
				BlockNo:        blockNo,
				TgId:           tg.TriggerId,
				TgUserId:       tg.UserId,
				Value:          matchedValue,
				AllValues:      allValues,
				BlockTimestamp: blockTimestamp,
			}
			cnMatches = append(cnMatches, match)
			log.Debugf("\tCN: Trigger %d matched on block %d\n", tg.TriggerId, blockNo)
		}
	}

	updateStatusForMatchingTriggers(idb, cnMatches)
	updateStatusForNonMatchingTriggers(idb, cnMatches, wacTriggers)

	log.Infof("\tCN: Processed %d triggers in %s from block %d", len(wacTriggers), time.Since(start), blockNo)
	return cnMatches
}

// set triggered flag to true for all matching 'false' triggers
func updateStatusForMatchingTriggers(idb aws.IDB, matches []*trigger.CnMatch) {
	var matchingTriggersIds []int
	for _, m := range matches {
		matchingTriggersIds = append(matchingTriggersIds, m.TgId)
	}
	idb.UpdateMatchingTriggers(matchingTriggersIds)
}

// set triggered flag to false for all non-matching 'true' triggers
func updateStatusForNonMatchingTriggers(idb aws.IDB, matches []*trigger.CnMatch, allTriggers []*trigger.Trigger) {
	setAll := make(map[int]struct{})
	setMatches := make(map[int]struct{})

	for _, t := range allTriggers {
		setAll[t.TriggerId] = struct{}{}
	}
	for _, m := range matches {
		setMatches[m.TgId] = struct{}{}
	}

	nonMatchingTriggersIds := getSliceFromIntSet(setDifference(setAll, setMatches))

	idb.UpdateNonMatchingTriggers(nonMatchingTriggersIds)
}

func getSliceFromIntSet(set map[int]struct{}) []int {
	out := make([]int, len(set))
	i := 0
	for k := range set {
		out[i] = k
		i++
	}
	return out
}

func setDifference(s1 map[int]struct{}, s2 map[int]struct{}) map[int]struct{} {
	diff := make(map[int]struct{})
	for v := range s1 {
		_, ok := s2[v]
		if ok {
			continue
		}
		diff[v] = struct{}{}
	}
	return diff
}

func isIn(a string, list []string) bool {
	for _, x := range list {
		if x == a {
			return true
		}
	}
	return false
}
