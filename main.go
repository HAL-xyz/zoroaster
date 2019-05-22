package main

import (
	"github.com/onrik/ethrpc"
	"log"
	"os"
	"zoroaster/aws"
	"zoroaster/rpc"
	"zoroaster/triggers"
)

func main() {

	// Load table config
	table := os.Getenv("DB_TABLE")
	if table == "" {
		table = "trigger1"
	}

	// Load triggers from DB
	aws.InitDB()
	triggers, err := aws.LoadTriggersFromDB(table)
	if err != nil {
		log.Fatal(err)
	}

	// Poll ETH node
	c := make(chan *ethrpc.Block)
	go rpc.PollForLastBlock(c)

	// Main routine
	for {
		block := <-c
		log.Println("=> Discovered new block: #", block.Number)

		for _, tg := range triggers {
			txs := trigger.MatchTrigger(tg, block)
			for _, tx := range txs {
				log.Printf("\tTrigger %d matched transaction %s", tg.TriggerId, tx.Hash)
			}
		}
	}

}
