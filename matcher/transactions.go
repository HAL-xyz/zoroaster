package matcher

import (
	"github.com/HAL-xyz/ethrpc"
	"github.com/HAL-xyz/zoroaster/db"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/HAL-xyz/zoroaster/trigger"
	log "github.com/sirupsen/logrus"
	"time"
)

func TxMatcher(blocksChan chan *ethrpc.Block, matchesChan chan trigger.IMatch, idb db.IDB, api tokenapi.ITokenAPI) {

	for {
		block := <-blocksChan
		api.GetRPCCli().ResetCounterAndLogStats(block.Number - 1)
		api.LogFiatStatsAndReset(block.Number - 1)
		start := time.Now()

		triggers, err := idb.LoadTriggersFromDB(trigger.WaT)
		if err != nil {
			log.Fatal(err)
		}
		for _, tg := range triggers {
			matchingTxs := trigger.MatchTransaction(tg, block, api)
			for _, m := range matchingTxs {
				if err = idb.LogMatch(m); err != nil {
					log.Fatal(err)
				}
				matchesChan <- m
			}
		}
		if err = idb.SetLastBlockProcessed(block.Number, trigger.WaT); err != nil {
			log.Fatal(err)
		}
		log.Infof("TX: Processed %d triggers in %s from block %d", len(triggers), time.Since(start), block.Number)
	}
}
