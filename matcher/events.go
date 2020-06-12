package matcher

import (
	"github.com/HAL-xyz/zoroaster/aws"
	"github.com/HAL-xyz/zoroaster/trigger"
	"github.com/onrik/ethrpc"
	"github.com/sirupsen/logrus"
	"time"
)

func EventMatcher(
	blocksChan chan *ethrpc.Block,
	matchesChan chan trigger.IMatch,
	idb aws.IDB,
	rpcCli trigger.IEthRpc) {

	for {
		block := <-blocksChan
		start := time.Now()
		logrus.Info("Events: new -> ", block.Number)

		triggers, err := idb.LoadTriggersFromDB(trigger.WaE)
		if err != nil {
			logrus.Fatal(err)
		}
		for _, tg := range triggers {
			matchingEvents := trigger.MatchEvent(rpcCli, tg, block.Number, block.Timestamp)
			for _, match := range matchingEvents {
				matchUUID, err := idb.LogMatch(match)
				if err != nil {
					logrus.Fatal(err)
				}
				match.MatchUUID = matchUUID
				logrus.Debug("\tlogged one event with id ", matchUUID)
				matchesChan <- *match
			}
		}
		err = idb.SetLastBlockProcessed(block.Number, trigger.WaE)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Infof("\tEvents: Processed %d triggers in %s from block %d", len(triggers), time.Since(start), block.Number)

		err = idb.LogAnalytics(trigger.WaE, block.Number, len(triggers), block.Timestamp, start, time.Now())
		if err != nil {
			logrus.Warn("cannot log analytics: ", err)
		}
	}
}
