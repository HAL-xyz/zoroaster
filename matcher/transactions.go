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
		log.Info("New block: #", block.Number)

		triggers, err := aws.LoadTriggersFromDB(zconf.TriggersDB.TableTriggers)
		if err != nil {
			log.Fatal(err)
		}
		for _, tg := range triggers {
			matchingZTxs := trigger.MatchTrigger(tg, block)
			for _, ztx := range matchingZTxs {
				log.Debugf("\tTrigger %d matched transaction https://etherscan.io/tx/%s", tg.TriggerId, ztx.Tx.Hash)
				m := trigger.Match{tg, ztx, 0}
				matchId := aws.LogMatch(zconf.TriggersDB.TableMatches, m)
				m.MatchId = matchId
				matchesChan <- &m
			}
		}
		aws.SetLastBlockProcessed(zconf.TriggersDB.TableStats, block.Number)
		log.Infof("\tProcessed %d triggers in %s", len(triggers), time.Since(start))
	}
}
