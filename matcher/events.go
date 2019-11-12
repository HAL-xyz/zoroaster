package matcher

import (
	"github.com/onrik/ethrpc"
	"github.com/sirupsen/logrus"
	"time"
	"zoroaster/aws"
	"zoroaster/trigger"
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
				matchUUID := idb.LogMatch(match)
				match.MatchUUID = matchUUID
				matchesChan <- match
			}
		}
		err = idb.SetLastBlockProcessed(block.Number, trigger.WaE)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Infof("\tEvents: Processed %d triggers in %s from block %d", len(triggers), time.Since(start), block.Number)
	}
}
