package matcher

import (
	"github.com/HAL-xyz/ethrpc"
	"github.com/HAL-xyz/zoroaster/aws"
	"github.com/HAL-xyz/zoroaster/rpc"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/HAL-xyz/zoroaster/trigger"
	"github.com/HAL-xyz/zoroaster/utils"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

func EventMatcher(
	blocksChan chan *ethrpc.Block,
	matchesChan chan trigger.IMatch,
	idb aws.IDB,
	tokenApi tokenapi.ITokenAPI) {

	for {
		block := <-blocksChan
		tokenApi.GetRPCCli().ResetCounterAndLogStats(block.Number - 1)
		start := time.Now()

		triggers, err := idb.LoadTriggersFromDB(trigger.WaE)
		if err != nil {
			logrus.Fatal(err)
		}

		logs, err := getLogsForBlock(tokenApi.GetRPCCli(), block.Hash, triggers)
		if err != nil {
			logrus.Fatalf("cannot fetch logs for block %d: %s\n", block.Number, err)
		}
		// fmt.Println(utils.GimmePrettyJson(logs))

		for _, tg := range triggers {
			matchingEvents := trigger.MatchEvent(tg, logs, tokenApi)
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

func getLogsForBlock(client rpc.IEthRpc, blockHash string, triggers []*trigger.Trigger) ([]ethrpc.Log, error) {
	filter := ethrpc.FilterParams{
		BlockHash: blockHash,
	}
	logs, err := client.EthGetLogs(filter)
	if err != nil {
		return nil, err
	}
	var relevantLogs []ethrpc.Log
	uniqueTgAddresses := getUniqueTriggerAddresses(triggers)
	for i, log := range logs {
		if utils.IsIn(strings.ToLower(log.Address), uniqueTgAddresses) {
			relevantLogs = append(relevantLogs, logs[i])
		}
	}
	logrus.Debugf("fetched %d total logs and %d relevant logs\n", len(logs), len(relevantLogs))
	return relevantLogs, nil
}

func getUniqueTriggerAddresses(tgs []*trigger.Trigger) []string {
	var ads = make([]string, len(tgs))
	for i, tg := range tgs {
		ads[i] = strings.ToLower(tg.ContractAdd)
	}
	return utils.Uniques(ads)
}
