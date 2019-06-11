package main

import (
	"bytes"
	"encoding/json"
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
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
			acts, err := aws.GetActions(zconf.TriggersDB.TableActions, match.Tg.TriggerId, match.Tg.UserId)
			if err != nil {
				log.Warnf("cannot get actions from db: %v", err)
			} else {
				if len(acts) == 0 {
					return
				}
				log.Debugf("\tMatched %d actions", len(acts))
				event := ActionEvent{match.BlockNo, match.Tx, acts}
				eventData, err := json.Marshal(event)
				if err != nil {
					log.Debug(err)
				} else {
					sendToHercules(eventData, zconf.HerculesEndpoint)
				}
			}
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

type ActionEvent struct {
	BlockNo int
	Tx      *ethrpc.Transaction
	Actions []string
}

func sendToHercules(data []byte, endpoint string) {

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Warn(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Warn(resp.Status)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Warn(err)
		} else {
			log.Warn(string(body))
		}
	} else {
		log.Debug("\tActionEvents successfully received :)")
	}
}