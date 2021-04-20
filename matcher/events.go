package matcher

import (
	"github.com/HAL-xyz/ethrpc"
	"github.com/HAL-xyz/zoroaster/db"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/HAL-xyz/zoroaster/trigger"
	"github.com/sirupsen/logrus"
	"time"
)

func EventMatcher(
	blocksChan chan *ethrpc.Block,
	matchesChan chan trigger.IMatch,
	idb db.IDB,
	tokenApi tokenapi.ITokenAPI) {

	for {
		block := <-blocksChan
		tokenApi.GetRPCCli().ResetCounterAndLogStats(block.Number - 1)
		tokenApi.LogFiatStatsAndReset(block.Number - 1)
		start := time.Now()

		triggers, err := idb.LoadTriggersFromDB(trigger.WaE)
		if err != nil {
			logrus.Fatal(err)
		}

		logs, err := getLogsForBlock(tokenApi.GetRPCCli(), block.Hash)
		if err != nil {
			logrus.Fatalf("cannot fetch logs for block %d: %s\n", block.Number, err)
		}
		// fmt.Println(utils.GimmePrettyJson(logs))

		for _, tg := range triggers {
			matchingEvents := trigger.MatchEvent(tg, logs, block.Transactions, tokenApi)
			for _, match := range matchingEvents {
				match.BlockTimestamp = block.Timestamp
				if err = idb.LogMatch(match); err != nil {
					logrus.Fatal(err)
				}
				logrus.Debug("\tlogged one event with id ", match.MatchUUID)
				matchesChan <- match
			}
		}
		if err = idb.SetLastBlockProcessed(block.Number, trigger.WaE); err != nil {
			logrus.Fatal(err)
		}
		logrus.Infof("Events: Processed %d triggers in %s from block %d", len(triggers), time.Since(start), block.Number)
	}
}

func getLogsForBlock(client tokenapi.IEthRpc, blockHash string) ([]ethrpc.Log, error) {
	logs, err := client.EthGetLogsByHash(blockHash)
	if err != nil {
		return nil, err
	}
	return logs, nil
}
