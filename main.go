package main

import (
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
	"zoroaster/aws"
	"zoroaster/config"
	"zoroaster/eth"
	"zoroaster/matcher"
	"zoroaster/trigger"
)

func main() {

	// Load Config
	zconf := config.Load("config")

	// Load AWS SES session
	sesSession := aws.GetSESSession()

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

	log.Info("Starting up Zoroaster, stage = ", zconf.Stage)

	// Init Postgres DB client
	psqlClient := aws.PostgresClient{}
	psqlClient.InitDB(zconf)

	// HTTP client
	httpClient := http.Client{}

	// ETH client
	ethClient := ethrpc.New(zconf.EthNode)

	// Channels are buffered so the poller doesn't stop queueing blocks
	// if one of the Matcher isn't up (during tests) of if WaC is very slow (which it is)
	// Another solution would be to have three different pollers, but for now this should do.
	txBlocksChan := make(chan *ethrpc.Block, 10000)
	cnBlocksChan := make(chan *ethrpc.Block, 10000)
	evBlocksChan := make(chan *ethrpc.Block, 10000)
	matchesChan := make(chan trigger.IMatch)

	// Poll ETH node
	go eth.BlocksPoller(txBlocksChan, cnBlocksChan, evBlocksChan, ethClient, &psqlClient)

	// Watch a Transaction
	go matcher.TxMatcher(txBlocksChan, matchesChan, &psqlClient)

	// Watch a Contract
	go matcher.ContractMatcher(cnBlocksChan, matchesChan, eth.GetModifiedAccounts, &psqlClient, ethClient)

	// Watch an Event
	go matcher.EventMatcher(evBlocksChan, matchesChan, &psqlClient, ethClient)

	// Main routine - process matches
	for {
		match := <-matchesChan
		go matcher.ProcessMatch(match, &psqlClient, sesSession, &httpClient)
	}
}
