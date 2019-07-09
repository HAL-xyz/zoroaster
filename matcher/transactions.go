package matcher

import (
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"time"
	"zoroaster/aws"
	"zoroaster/config"
	"zoroaster/trigger"
)

func TxMatcher(
	blocksChan chan *ethrpc.Block,
	matchesChan chan *trigger.Match,
	zconf *config.ZConfiguration,
	idb aws.IDB) {

	for {
		block := <-blocksChan
		start := time.Now()
		log.Info("TX: new -> ", block.Number)

		triggers, err := idb.LoadTriggersFromDB(zconf.TriggersDB.TableTriggers, "WatchTransactions")
		if err != nil {
			log.Fatal(err)
		}
		for _, tg := range triggers {
			matchingZTxs := trigger.MatchTrigger(tg, block)
			for _, ztx := range matchingZTxs {
				log.Debugf("\tTX: Trigger %d matched transaction https://etherscan.io/tx/%s", tg.TriggerId, ztx.Tx.Hash)
				m := trigger.Match{tg, ztx, 0}
				matchId := idb.LogMatch(zconf.TriggersDB.TableMatches, m)
				m.MatchId = matchId
				matchesChan <- &m
			}
		}
		idb.SetLastBlockProcessed(zconf.TriggersDB.TableStats, block.Number, "wat")
		log.Infof("\tTX: Processed %d triggers in %s from block %d", len(triggers), time.Since(start), block.Number)
	}
}

func ContractMatcher(
	blocksChan chan int,
	matchesChan chan *trigger.Match,
	zconf *config.ZConfiguration,
	getModifiedAccounts func(prevBlock, currBlock int) []string,
	idb aws.IDB,
	client *ethrpc.EthRPC) {

	for {
		blockNo := <-blocksChan
		log.Info("CN: new -> ", blockNo)

		MatchContractsForBlock(blockNo, getModifiedAccounts, zconf, idb, client)
		// TODO: log matches, put matches to chan

		idb.SetLastBlockProcessed(zconf.TriggersDB.TableStats, blockNo, "wac")
	}
}

func MatchContractsForBlock(
	blockNo int,
	getModAccounts func(prevBlock, currBlock int) []string,
	zconf *config.ZConfiguration,
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

	triggers, err := idb.LoadTriggersFromDB(zconf.TriggersDB.TableTriggers, "WatchContracts")
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
	log.Debug("\tmatching triggers: ", len(wacTriggers))

	var cnMatches []*trigger.CnMatch
	for _, tg := range wacTriggers {
		contractValue := trigger.MatchContract(client, tg, blockNo)
		if contractValue != "" {
			cnMatches = append(cnMatches, &trigger.CnMatch{blockNo, tg.TriggerId, contractValue})
			log.Debugf("`\tCN: Trigger %d matched on block %d\n", tg.TriggerId, blockNo)
		}
	}
	log.Infof("\tCN: Processed %d triggers in %s from block %d", len(wacTriggers), time.Since(start), blockNo)
	return cnMatches
}

func isIn(a string, list []string) bool {
	for _, x := range list {
		if x == a {
			return true
		}
	}
	return false
}
