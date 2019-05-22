package main

import (
	"github.com/onrik/ethrpc"
	"log"
	"zoroaster/aws"
	"zoroaster/rpc"
	"zoroaster/triggers"
)

func main() {

	// Load triggers from DB
	aws.InitDB()
	triggers, err := aws.LoadTriggersFromDB()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("=> Loaded %d triggers from DB\n", len(triggers))

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
