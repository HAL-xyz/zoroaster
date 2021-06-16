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

		logs, err := getLogsForBlock(tokenApi.GetRPCCli(), block.Hash, block.Number, 3, nil)
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

func getLogsForBlock(client tokenapi.IEthRpc, blockHash string, blockNo, retries int, lastErr error) ([]ethrpc.Log, error) {
	if retries == 0 {
		return nil, lastErr
	}
	logs, err := client.EthGetLogsByHash(blockHash)
	if err != nil {
		logrus.Warnf("cannot fetch logs for block #%d, err: %v - retrying %d times", blockNo, err, retries)
		time.Sleep(500 * time.Millisecond)
		return getLogsForBlock(client, blockHash, blockNo, retries-1, err)
	}
	return logs, nil
}
