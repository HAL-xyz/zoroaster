package matcher

import (
	"fmt"
	"github.com/HAL-xyz/ethrpc"
	"github.com/HAL-xyz/zoroaster/db"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/HAL-xyz/zoroaster/trigger"
	"github.com/HAL-xyz/zoroaster/utils"
	"github.com/gorhill/cronexpr"
	"github.com/sirupsen/logrus"
	"time"
)

func CronScheduler(idb db.IDB, api tokenapi.ITokenAPI, matchesChan chan trigger.IMatch) {
	ticker := time.NewTicker(15 * time.Second)
	for range ticker.C {
		CronExecutor(idb, time.Now(), api, matchesChan)
	}
}

func CronExecutor(idb db.IDB, now time.Time, api tokenapi.ITokenAPI, matchesChan chan trigger.IMatch) {
	start := time.Now()
	allTriggers, err := idb.LoadTriggersFromDB(trigger.CronT)
	if err != nil {
		logrus.Fatal(err)
	}

	tgsToRun := filterTgsToRun(allTriggers, now)
	if len(tgsToRun) == 0 {
		return
	}

	lastBlock := fetchLastBlock(api)

	for _, tg := range tgsToRun {
		if err = idb.UpdateLastFired(tg.TriggerUUID, now.UTC()); err != nil {
			logrus.Fatal(err)
		}
		m, err := RunCronTgAgainstBlock(tg, lastBlock.Number, api)
		if err != nil {
			logrus.Warnf("cannot exec cron trig %s: %s", tg.TriggerUUID, err)
			continue
		}
		m.BlockTimestamp, m.BlockHash = lastBlock.Timestamp, lastBlock.Hash

		if err = idb.LogMatch(m); err != nil {
			logrus.Fatal(err)
		}
		matchesChan <- m
	}
	logrus.Infof("CronT: total tgs: %d; executed: %d in  %s", len(allTriggers), len(tgsToRun), time.Since(start))
}

func shouldFire(tg *trigger.Trigger, now time.Time) bool {
	expr, err := cronexpr.Parse(tg.CronJob.Rule)
	if err != nil {
		return false
	}

	const referenceLayout = "-0700"
	tz, _ := time.Parse(referenceLayout, tg.CronJob.Timezone)
	lastFired := tg.LastFired.In(tz.Location())
	nextTime := expr.Next(lastFired)

	return now.Equal(nextTime) || now.After(nextTime)
}

func filterTgsToRun(tgs []*trigger.Trigger, now time.Time) []*trigger.Trigger {
	var tgsToFire []*trigger.Trigger
	for _, tg := range tgs {
		if shouldFire(tg, now) {
			tgsToFire = append(tgsToFire, tg)
		}
	}
	return tgsToFire
}

func RunCronTgAgainstBlock(tg *trigger.Trigger, blockNo int, api tokenapi.ITokenAPI) (*trigger.CnMatch, error) {

	result, err := api.EthCall(tg.ContractAdd, tg.FunctionName, tg.ContractABI, blockNo, tg.CallArgs()...)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	m := &trigger.CnMatch{
		Trigger:     tg,
		BlockNumber: blockNo,
		AllValues:   utils.SprintfInterfaces(result),
	}
	return m, nil
}

func fetchLastBlock(api tokenapi.ITokenAPI) *ethrpc.Block {
	lastBlock, err := api.GetRPCCli().EthBlockNumber()
	if err != nil {
		logrus.Fatal(err)
	}
	block, err := api.GetRPCCli().EthGetBlockByNumber(lastBlock, false)
	if err != nil {
		logrus.Fatal(err)
	}
	return block
}
