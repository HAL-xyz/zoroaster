package matcher

import (
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"time"
	"zoroaster/aws"
	"zoroaster/trigger"
)

func TxMatcher(blocksChan chan *ethrpc.Block, matchesChan chan trigger.IMatch, idb aws.IDB) {

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
				matchId := idb.LogMatch(m)
				m.MatchId = matchId
				matchesChan <- &m
			}
		}
		idb.SetLastBlockProcessed(block.Number, "wat")
		log.Infof("\tTX: Processed %d triggers in %s from block %d", len(triggers), time.Since(start), block.Number)
	}
}
