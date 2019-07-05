package matcher

import (
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"time"
	"zoroaster/aws"
	"zoroaster/config"
	"zoroaster/trigger"
)

func TxMatcher(blocksChan chan *ethrpc.Block, matchesChan chan *trigger.Match, zconf *config.ZConfiguration) {
	for {
		block := <-blocksChan
		start := time.Now()
		log.Info("TxMatcher: Processing: #", block.Number)

		triggers, err := aws.LoadTriggersFromDB(zconf.TriggersDB.TableTriggers)
		if err != nil {
			log.Fatal(err)
		}
		for _, tg := range triggers {
			matchingZTxs := trigger.MatchTrigger(tg, block)
			for _, ztx := range matchingZTxs {
				log.Debugf("\tTxMatcher: Trigger %d matched transaction https://etherscan.io/tx/%s", tg.TriggerId, ztx.Tx.Hash)
				m := trigger.Match{tg, ztx, 0}
				matchId := aws.LogMatch(zconf.TriggersDB.TableMatches, m)
				m.MatchId = matchId
				matchesChan <- &m
			}
		}
		aws.SetLastBlockProcessed(zconf.TriggersDB.TableStats, block.Number)
		log.Infof("\tTxMatcher: Processed %d triggers in %s", len(triggers), time.Since(start))
	}
}

func ContractMatcher(
	blocksChan chan *ethrpc.Block,
	matchesChan chan *trigger.Match,
	triggersTable string,
	getModifiedAccounts func(prevBlock, currBlock int) []string,
	idb aws.IDB,
	client *ethrpc.EthRPC) {

	for {
		block := <-blocksChan
		//start := time.Now()
		log.Info("CntMatcher: New block: #", block.Number)

		MatchContractsForBlock(block.Number, getModifiedAccounts, triggersTable, idb, client)
	}
}

func MatchContractsForBlock(
	blockNo int,
	getModAccounts func(prevBlock, currBlock int) []string,
	tableName string,
	idb aws.IDB,
	client *ethrpc.EthRPC) {

	modAccounts := getModAccounts(blockNo-1, blockNo)
	log.Debug("modified accounts: ", len(modAccounts))

	triggers, err := idb.LoadTriggersFromDB(tableName)
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("triggers from IDB: ", len(triggers))

	var wacTriggers []*trigger.Trigger
	for i, t := range triggers {
		if t.TriggerType == "WatchContracts" {
			if isIn(t.ContractAdd, modAccounts) {
				wacTriggers = append(wacTriggers, triggers[i])
			}
		}
	}
	log.Debug("matching triggers: ", len(wacTriggers))

	for _, tg := range wacTriggers {
		if trigger.MatchContract(client, tg, blockNo) {
			log.Infof("CntMatcher: Trigger %d matched on block %d\n", tg.TriggerId, blockNo)
		}
	}
}

func isIn(a string, list []string) bool {
	for _, x := range list {
		if x == a {
			return true
		}
	}
	return false
}
