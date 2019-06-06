package main

import (
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
	"zoroaster/aws"
	"zoroaster/config"
	"zoroaster/eth"
	"zoroaster/triggers"
)

func main() {

	// Load Config
	zconf := config.Load()

	// Persist logs
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.Stamp,
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

	// Channels
	blocksChan := make(chan *ethrpc.Block)
	matchesChan := make(chan *Match)

	// Poll ETH node
	go eth.BlocksPoller(blocksChan, client, zconf)

	// Matching blocks and triggers
	go Matcher(blocksChan, matchesChan, zconf)

	// Main routine - send actions
	for {
		match := <-matchesChan
		go func() {
			log.Debugf("\tgot a match from %d", match.BlockNo)
			time.Sleep(1 * time.Second)
			// TODO: create an Action json to be sent to the Lambda
		}()
	}
}

func Matcher(blocksChan chan *ethrpc.Block, matchesChan chan *Match, zconf *config.ZConfiguration) {
	for {
		block := <-blocksChan
		start := time.Now()
		log.Info("New block: #", block.Number)

		triggers, err := aws.LoadTriggersFromDB(zconf.TriggersDB.TableData)
		if err != nil {
			log.Fatal(err)
		}

		for _, tg := range triggers {
			matchingTxs := trigger.MatchTrigger(tg, block)
			for _, tx := range matchingTxs {
				log.Debugf("\tTrigger %d matched transaction https://etherscan.io/tx/%s", tg.TriggerId, tx.Hash)
				aws.LogMatch(zconf.TriggersDB.TableLogs, tg, tx, block.Timestamp)

				m := Match{block.Number, tg, tx}
				matchesChan <- &m
			}
		}
		aws.SetLastBlockProcessed(zconf.TriggersDB.TableStats, block.Number)
		log.Infof("\tProcessed %d triggers in %s", len(triggers), time.Since(start))
	}
}

type Match struct {
	BlockNo int
	Tg      *trigger.Trigger
	Tx      *ethrpc.Transaction
}
