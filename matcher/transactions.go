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

		triggers, err := idb.LoadTriggersFromDB(trigger.WaT)
		if err != nil {
			log.Fatal(err)
		}
		for _, tg := range triggers {
			matchingZTxs := trigger.MatchTransaction(tg, block)
			for _, ztx := range matchingZTxs {
				log.Debugf("\tTX: Trigger %s matched transaction https://etherscan.io/tx/%s", tg.TriggerUUID, ztx.Tx.Hash)
				m := trigger.TxMatch{"", tg, ztx}
				matchUUID := idb.LogMatch(m)
				m.MatchUUID = matchUUID
				matchesChan <- &m
			}
		}
		err = idb.SetLastBlockProcessed(block.Number, trigger.WaT)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("\tTX: Processed %d triggers in %s from block %d", len(triggers), time.Since(start), block.Number)
	}
}
