package poller

import (
	"github.com/HAL-xyz/ethrpc"
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/HAL-xyz/zoroaster/db"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/HAL-xyz/zoroaster/trigger"
	log "github.com/sirupsen/logrus"
	"time"
)

func BlocksPoller(txChan, cnChan, evChan chan *ethrpc.Block, idb db.IDB, client tokenapi.IEthRpc) {

	txLastBlockProcessed, err1 := idb.ReadLastBlockProcessed(trigger.WaT)
	cnLastBlockProcessed, err2 := idb.ReadLastBlockProcessed(trigger.WaC)
	evLastBlockProcessed, err3 := idb.ReadLastBlockProcessed(trigger.WaE)

	if err1 != nil || err2 != nil || err3 != nil {
		log.Fatal(err1, err2, err3)
	}

	ticker := time.NewTicker(time.Duration(config.Zconf.PollingInterval) * time.Second)
	for range ticker.C {

		// Watch a Transaction
		blockTx := fetchLastBlock(time.Now(), txLastBlockProcessed, client, true)
		if blockTx != nil {
			txChan <- blockTx
			txLastBlockProcessed = blockTx.Number
		}

		// Watch a Contract
		blockCn := fetchLastBlock(time.Now(), cnLastBlockProcessed, client, false)
		if blockCn != nil {
			// Since templating client is shared between WaT/C/E, we reset the stats after every new
			// block discovered by WaC. This way stats will be overall consistent, although they might
			// be slightly off on a per-block basis.
			client.ResetCounterAndLogStats(cnLastBlockProcessed)              // BlocksPoller eth client
			tokenapi.GetTokenAPI().ResetETHRPCstats(cnLastBlockProcessed)     // Templating eth client
			tokenapi.GetTokenAPI().LogFiatStatsAndReset(cnLastBlockProcessed) // Templating eth client

			cnChan <- blockCn
			cnLastBlockProcessed = blockCn.Number
		}

		// Watch an Event
		blockEv := fetchLastBlock(time.Now(), evLastBlockProcessed, client, true)
		if blockEv != nil {
			evChan <- blockEv
			evLastBlockProcessed = blockEv.Number
		}
	}
}

func fetchLastBlock(timeNow time.Time, lastBlockProcessed int, client tokenapi.IEthRpc, withTxs bool) *ethrpc.Block {

	log.Info("last block processed is: ", lastBlockProcessed)
	// this is used to reset the last block processed
	if lastBlockProcessed == 0 {
		log.Infof("read block 0; fetching latest block...")
		lastBlockSeen, err := client.EthBlockNumber()
		if err != nil {
			log.Fatal("failed to poll ETH node -> ", err)
		}
		lastBlockProcessed = lastBlockSeen - 5 // leave some room for soft forks
		log.Infof("latest block set to: %d", lastBlockProcessed)
	}

	block, err := client.EthGetBlockByNumber(lastBlockProcessed+1, withTxs)
	if err != nil {
		log.Warnf("failed to get block by number: %d - %s", lastBlockProcessed+1, err)
		return nil
	}
	if block == nil {
		log.Debugf("block %d doesn't exist yet", lastBlockProcessed+1)
		return nil // the block doesn't exist yet
	}

	log.Debug("Delay is: ", time.Since(time.Unix(int64(block.Timestamp), 0)))
	if timeNow.Sub(time.Unix(int64(block.Timestamp), 0)) > 30*time.Second {
		return block
	}
	return nil // respect delay
}
