package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/onrik/ethrpc"
	"os"
	"time"
	"zoroaster/aws"
	"zoroaster/config"
	"zoroaster/rpc"
	"zoroaster/triggers"
)

func main() {

	// Load Config
	zconf := config.Load()

	// Persist logs
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC822,
	})
	log.SetLevel(log.DebugLevel)
	f, err := os.OpenFile(zconf.LogsFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(f)

	// Connect to triggers' DB
	aws.InitDB(zconf)

	// ETH client
	client := ethrpc.New(zconf.EthNode)

	// Poll ETH node
	c := make(chan *ethrpc.Block)
	go rpc.PollForLastBlock(c, client, zconf)

	lastBlockProcessed := 0
	// Main routine
	for {
		block := <-c
		start := time.Now()
		log.Info("New block: #", block.Number)
		logLostBlocks(lastBlockProcessed, block.Number)

		triggers, err := aws.LoadTriggersFromDB(zconf.TriggersDB.TableData)
		if err != nil {
			log.Fatal(err)
		}
		for _, tg := range triggers {
			txs := trigger.MatchTrigger(tg, block)
			for _, tx := range txs {
				log.Infof("\tTrigger %d matched transaction https://etherscan.io/tx/%s", tg.TriggerId, tx.Hash)
				aws.LogMatch(zconf.TriggersDB.TableLogs, tg, tx, block.Timestamp)
			}
		}
		log.Infof("\tProcessed %d triggers in %s", len(triggers), time.Since(start))
		lastBlockProcessed = block.Number
		aws.SetLastBlockProcessed(zconf.TriggersDB.TableStats, lastBlockProcessed)
	}
}

func logLostBlocks(lastBlockProcessed int, lastBlockPolled int) {
	delta := lastBlockPolled - lastBlockProcessed
	if delta != 1 && lastBlockProcessed != 0 {
		log.Warnf("we lost %d block(s)", delta-1)
	}
}
