package main

import (
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
	"zoroaster/actions"
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
	matchesChan := make(chan *trigger.Match)

	// Poll ETH node
	go eth.BlocksPoller(blocksChan, client, zconf)

	// Matching blocks and triggers
	go Matcher(blocksChan, matchesChan, zconf)

	// Main routine - process actions
	for {
		match := <-matchesChan
		go func() {
			acts := getActions(zconf.TriggersDB.TableActions, match.Tg.TriggerId, match.Tg.UserId)
			eventJson := actions.ActionEventJson{ZTx: match.ZTx, Actions: acts}
			actions.HandleEvent(eventJson)
		}()
	}
}

func getActions(table string, tgid int, usrid int) []string {
	var actions []string
	actions, err := aws.GetActions(table, tgid, usrid)
	if err != nil {
		log.Warnf("cannot get actions from db: %v", err)
	}
	log.Debugf("\tMatched %d actions", len(actions))
	return actions
}

func Matcher(blocksChan chan *ethrpc.Block, matchesChan chan *trigger.Match, zconf *config.ZConfiguration) {
	for {
		block := <-blocksChan
		start := time.Now()
		log.Info("New block: #", block.Number)

		triggers, err := aws.LoadTriggersFromDB(zconf.TriggersDB.TableData)
		if err != nil {
			log.Fatal(err)
		}
		for _, tg := range triggers {
			matchingZTxs := trigger.MatchTrigger(tg, block)
			for _, ztx := range matchingZTxs {
				log.Debugf("\tTrigger %d matched transaction https://etherscan.io/tx/%s", tg.TriggerId, ztx.Tx.Hash)
				m := trigger.Match{tg, ztx}
				aws.LogMatch(zconf.TriggersDB.TableLogs, m)
				matchesChan <- &m
			}
		}
		aws.SetLastBlockProcessed(zconf.TriggersDB.TableStats, block.Number)
		log.Infof("\tProcessed %d triggers in %s", len(triggers), time.Since(start))
	}
}
