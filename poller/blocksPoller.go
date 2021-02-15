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

func BlocksPoller(
	txChan chan *ethrpc.Block,
	cnChan chan *ethrpc.Block,
	evChan chan *ethrpc.Block,
	client tokenapi.IEthRpc,
	idb db.IDB,
	blocksDelay int) {

	txLastBlockProcessed, err1 := idb.ReadLastBlockProcessed(trigger.WaT)
	cnLastBlockProcessed, err2 := idb.ReadLastBlockProcessed(trigger.WaC)
	evLastBlockProcessed, err3 := idb.ReadLastBlockProcessed(trigger.WaE)

	if err1 != nil || err2 != nil || err3 != nil {
		log.Fatal(err1, err2, err3)
	}

	ticker := time.NewTicker(time.Duration(config.Zconf.PollingInterval) * time.Second)
	for range ticker.C {
		lastBlockSeen, err := client.EthBlockNumber()
		if err != nil {
			log.Fatal("failed to poll ETH node -> ", err)
		}

		// Watch a Transaction
		fetchLastBlock(lastBlockSeen, &txLastBlockProcessed, txChan, client, true, blocksDelay)

		// Watch a Contract
		fetchLastBlock(lastBlockSeen, &cnLastBlockProcessed, cnChan, client, false, blocksDelay)

		// Watch an Event
		fetchLastBlock(lastBlockSeen, &evLastBlockProcessed, evChan, client, false, blocksDelay)
	}
}

func fetchLastBlock(
	lastBlockSeen int,
	lastBlockProcessed *int,
	ch chan *ethrpc.Block,
	client tokenapi.IEthRpc,
	withTxs bool,
	blocksDelay int) {

	// this is used to reset the last block processed
	if *lastBlockProcessed == 0 {
		*lastBlockProcessed = lastBlockSeen - blocksDelay
	}

	if lastBlockSeen-blocksDelay > *lastBlockProcessed {
		if withTxs {
			// Since templating client is shared between WaT/C/E, we reset the stats after every new
			// block discovered by WaT. This way stats will be overall consistent, although they might
			// be slightly off on a per-block basis.
			client.ResetCounterAndLogStats(*lastBlockProcessed)          // BlocksPoller eth client
			tokenapi.GetTokenAPI().ResetETHRPCstats(*lastBlockProcessed) // Templating eth client
		}

		block, err := client.EthGetBlockByNumber(*lastBlockProcessed+1, withTxs)
		if err != nil {
			log.Fatalf("failed to get block %d -> %s", *lastBlockProcessed+1, err)
		} else {
			*lastBlockProcessed += 1
			ch <- block
		}
	}
}
