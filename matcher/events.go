package matcher

import (
	"fmt"
	"github.com/HAL-xyz/ethrpc"
	"github.com/HAL-xyz/zoroaster/aws"
	"github.com/HAL-xyz/zoroaster/rpc"
	"github.com/HAL-xyz/zoroaster/trigger"
	"github.com/HAL-xyz/zoroaster/utils"
	"github.com/sirupsen/logrus"
	"time"
)

func EventMatcher(
	blocksChan chan *ethrpc.Block,
	matchesChan chan trigger.IMatch,
	idb aws.IDB,
	rpcCli rpc.IEthRpc) {

	for {
		block := <-blocksChan
		rpcCli.ResetCounterAndLogStats(block.Number - 1)
		start := time.Now()
		logrus.Info("Events: new -> ", block.Number)

		triggers, err := idb.LoadTriggersFromDB(trigger.WaE)
		if err != nil {
			logrus.Fatal(err)
		}

		logs, err := getLogsForBlock(rpcCli, block.Hash, getUniqueTriggerAddresses(triggers))
		logrus.Debugf("fetched %d logs\n", len(logs))
		if err != nil {
			logrus.Fatalf("cannot fetch logs for block %d: %s\n", block.Number, err)
		}
		// fmt.Println(utils.GimmePrettyJson(logs))

		for _, tg := range triggers {
			matchingEvents := trigger.MatchEvent(tg, block.Timestamp, logs)
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

func getLogsForBlock(client rpc.IEthRpc, blockHash string, addresses []string) ([]ethrpc.Log, error) {
	filter := ethrpc.FilterParams{
		BlockHash: blockHash,
		Address:   addresses,
	}
	return client.EthGetLogs(filter)
}

func getUniqueTriggerAddresses(tgs []*trigger.Trigger) []string {
	var ads = make([]string, len(tgs))
	for i, tg := range tgs {
		ads[i] = tg.ContractAdd
	}
	return utils.Uniques(ads)
}
