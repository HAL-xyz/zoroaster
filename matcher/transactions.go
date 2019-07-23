package matcher

import (
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"time"
	"zoroaster/aws"
	"zoroaster/trigger"
)

func TxMatcher(blocksChan chan *ethrpc.Block, matchesChan chan interface{}, idb aws.IDB) {

	for {
		block := <-blocksChan
		start := time.Now()
		log.Info("TX: new -> ", block.Number)

		triggers, err := idb.LoadTriggersFromDB("WatchTransactions")
		if err != nil {
			log.Fatal(err)
		}
		for _, tg := range triggers {
			matchingZTxs := trigger.MatchTrigger(tg, block)
			for _, ztx := range matchingZTxs {
				log.Debugf("\tTX: Trigger %d matched transaction https://etherscan.io/tx/%s", tg.TriggerId, ztx.Tx.Hash)
				m := trigger.TxMatch{0, tg, ztx}
				matchId := idb.LogTxMatch(m)
				m.MatchId = matchId
				matchesChan <- &m
			}
		}
		idb.SetLastBlockProcessed(block.Number, "wat")
		log.Infof("\tTX: Processed %d triggers in %s from block %d", len(triggers), time.Since(start), block.Number)
	}
}

func ContractMatcher(
	blocksChan chan int,
	matchesChan chan interface{},
	getModifiedAccounts func(prevBlock, currBlock int) []string,
	idb aws.IDB,
	client *ethrpc.EthRPC) {

	for {
		blockNo := <-blocksChan
		log.Info("CN: new -> ", blockNo)

		cnMatches := MatchContractsForBlock(blockNo, getModifiedAccounts, idb, client)
		for _, m := range cnMatches {
			matchId := idb.LogCnMatch(*m)
			m.MatchId = matchId
			log.Debug("\tlogged one match with id ", matchId)
			matchesChan <- m
		}
		idb.SetLastBlockProcessed(blockNo, "wac")
	}
}

func MatchContractsForBlock(
	blockNo int,
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
		contractValue := trigger.MatchContract(client, tg, blockNo)
		if contractValue != "" {
			cnMatches = append(cnMatches, &trigger.CnMatch{0, blockNo, tg.TriggerId, tg.UserId, contractValue})
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
