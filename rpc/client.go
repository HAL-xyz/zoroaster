package rpc

import (
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"time"
	"zoroaster/aws"
	"zoroaster/config"
)

const K = 8 // next block to process is (last block mined - K)

func PollForLastBlock(c chan *ethrpc.Block, client *ethrpc.EthRPC, zconf *config.ZConfiguration) {

	lastBlockProcessed := aws.ReadLastBlockProcessed(zconf.TriggersDB.TableStats)

	ticker := time.NewTicker(2500 * time.Millisecond)
	for range ticker.C {
		n, err := client.EthBlockNumber()
		if err != nil {
			log.Warn("failed to poll ETH node -> ", err)
			time.Sleep(5 * time.Second)
			continue
		}
		// this should only happen during dev
		if lastBlockProcessed == 0 {
			lastBlockProcessed = n - K
		}
		if n-K > lastBlockProcessed {
			block, err := client.EthGetBlockByNumber(lastBlockProcessed+1, true)
			if err != nil {
				log.Warnf("failed to get block %d -> %s", n, err)
				time.Sleep(5 * time.Second)
				continue
			}
			lastBlockProcessed += 1
			log.Infof("\t(%d blocks behind)", n-lastBlockProcessed)
			c <- block
		}
	}
}
