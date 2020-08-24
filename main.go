package main

import (
	"github.com/HAL-xyz/ethrpc"
	"github.com/HAL-xyz/zoroaster/aws"
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/HAL-xyz/zoroaster/db"
	"github.com/HAL-xyz/zoroaster/matcher"
	"github.com/HAL-xyz/zoroaster/poller"
	"github.com/HAL-xyz/zoroaster/rpc"
	"github.com/HAL-xyz/zoroaster/trigger"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func main() {

	// Load AWS SES session
	sesSession := aws.GetSESSession()

	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)

	log.Info("Starting up Zoroaster, stage = ", config.Zconf.Stage)

	// Init Postgres DB client
	psqlClient := aws.PostgresClient{}
	psqlClient.InitDB(config.Zconf)

	// HTTP client
	httpClient := http.Client{}

	// ETH client
	ethClient := ethrpc.New(config.Zconf.EthNode)

	// Run monthly matches update
	go db.MatchesMonthlyUpdate(&psqlClient)

	// Channels are buffered so the poller doesn't stop queueing blocks
	// if one of the Matcher isn't up (during tests) of if WaC is very slow (which it is)
	// Another solution would be to have three different pollers, but for now this should do.
	txBlocksChan := make(chan *ethrpc.Block, 10000)
	cnBlocksChan := make(chan *ethrpc.Block, 10000)
	evBlocksChan := make(chan *ethrpc.Block, 10000)
	matchesChan := make(chan trigger.IMatch)

	// Poll ETH node
	pollerCli := rpc.New(ethClient, "BlocksPoller")
	go poller.BlocksPoller(txBlocksChan, cnBlocksChan, evBlocksChan, pollerCli, &psqlClient, config.Zconf.BlocksDelay)

	// Watch a Transaction
	go matcher.TxMatcher(txBlocksChan, matchesChan, &psqlClient)

	// Watch a Contract
	wacCli := rpc.New(ethClient, "Watch a Contract")
	go matcher.ContractMatcher(cnBlocksChan, matchesChan, &psqlClient, wacCli)

	// Watch an Event
	waeCli := rpc.New(ethClient, "Watch an Event")
	go matcher.EventMatcher(evBlocksChan, matchesChan, &psqlClient, waeCli)

	// Main routine - process matches
	for {
		match := <-matchesChan
		go matcher.ProcessMatch(match, &psqlClient, sesSession, &httpClient)
	}
}
