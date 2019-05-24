package main

import (
	"github.com/onrik/ethrpc"
	"log"
	"os"
	"time"
	"zoroaster/aws"
	"zoroaster/rpc"
	"zoroaster/triggers"
)

func main() {

	// Load table config
	table := os.Getenv("DB_TABLE")
	if table == "" {
		table = "trigger_default"
	}

	// Connect to triggers' DB
	aws.InitDB()

	// Poll ETH node
	c := make(chan *ethrpc.Block)
	go rpc.PollForLastBlock(c)

	lastBlockProcessed := 0
	// Main routine
	for {
		block := <-c
		start := time.Now()
		log.Println("New block: #", block.Number)
		logLostBlocks(lastBlockProcessed, block.Number)

		triggers, err := aws.LoadTriggersFromDB(table)
		if err != nil {
			log.Fatal(err)
		}

		for _, tg := range triggers {
			txs := trigger.MatchTrigger(tg, block)
			for _, tx := range txs {
				log.Printf("\tTrigger %d matched transaction https://etherscan.io/tx/%s", tg.TriggerId, tx.Hash)
				aws.LogMatch(tg, tx, table+"_log")

			}
		}
		log.Printf("\tProcessed %d triggers in %s", len(triggers), time.Since(start))
		lastBlockProcessed = block.Number
	}

}

func logLostBlocks(lastBlockProcessed int, lastBlockPolled int) {
	delta := lastBlockPolled - lastBlockProcessed
	if delta != 1 && lastBlockProcessed != 0 {
		log.Printf("WARN: we lost %d block(s)", delta-1)
	}
}
