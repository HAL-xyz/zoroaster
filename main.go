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
	matchesChan := make(chan *trigger.Match)

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
				event := trigger.ActionEvent{ZTx: match.ZTx, Actions: acts}
				sendToHercules(event, zconf.HerculesEndpoint)
			}
		}()
	}
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

func sendToHercules(event trigger.ActionEvent, endpoint string) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Debug(err)
		return
	}
	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Warn(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Debug("\tActionEvents successfully received :)")
	} else {
		log.Warn(resp.Status)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Warn(err)
		} else {
			log.Warn(string(body))
		}
	}
}
