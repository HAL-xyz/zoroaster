package eth

import (
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"time"
	"zoroaster/aws"
	"zoroaster/trigger"
)

func BlocksPoller(
	txChan chan *ethrpc.Block,
	cnChan chan *ethrpc.Block,
	evChan chan *ethrpc.Block,
	client *ethrpc.EthRPC,
	idb aws.IDB,
	blocksDelay int) {

	txLastBlockProcessed, err1 := idb.ReadLastBlockProcessed(trigger.WaT)
	cnLastBlockProcessed, err2 := idb.ReadLastBlockProcessed(trigger.WaC)
	evLastBlockProcessed, err3 := idb.ReadLastBlockProcessed(trigger.WaE)

	if err1 != nil || err2 != nil || err3 != nil {
		log.Fatal(err1, err2, err3)
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	for range ticker.C {
		lastBlockSeen, err := client.EthBlockNumber()
		if err != nil {
			log.Warn("failed to poll ETH node -> ", err)
			time.Sleep(5 * time.Second)
			continue
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
	client *ethrpc.EthRPC,
	withTxs bool,
	blocksDelay int) {

	// this is used to reset the last block processed
	if *lastBlockProcessed == 0 {
		*lastBlockProcessed = lastBlockSeen - blocksDelay
	}

	if lastBlockSeen-blocksDelay > *lastBlockProcessed {
		block, err := client.EthGetBlockByNumber(*lastBlockProcessed+1, withTxs)
		if err != nil {
			log.Warnf("failed to get block %d -> %s", *lastBlockProcessed+1, err)
			time.Sleep(5 * time.Second)
		} else {
			*lastBlockProcessed += 1
			ch <- block
		}
	}
}
