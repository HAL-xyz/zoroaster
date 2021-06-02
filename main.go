package main

import (
	"github.com/HAL-xyz/ethrpc"
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/HAL-xyz/zoroaster/db"
	"github.com/HAL-xyz/zoroaster/matcher"
	"github.com/HAL-xyz/zoroaster/poller"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/HAL-xyz/zoroaster/trigger"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func main() {

	// Load AWS SES session
	sesSession := config.GetSESSession()

	log.SetLevel(config.Zconf.LogLevel)
	log.SetOutput(os.Stdout)

	log.Infof("Starting up Zoroaster, stage = %s, network = %s\n", config.Zconf.Stage, config.Zconf.Network)

	// Postgres DB client
	psqlClient := db.NewPostgresClient(config.Zconf)

	// HTTP client
	httpClient := http.Client{}

	// Run monthly matches update
	go db.MatchesMonthlyUpdate(psqlClient)

	// Channels are buffered so the poller doesn't stop queueing blocks
	// if one of the Matcher isn't up (during tests) of if WaC is very slow (which it is)
	// Another solution would be to have three different pollers, but for now this should do.
	txBlocksChan := make(chan *ethrpc.Block, 10000)
	cnBlocksChan := make(chan *ethrpc.Block, 10000)
	evBlocksChan := make(chan *ethrpc.Block, 10000)
	matchesChan := make(chan trigger.IMatch)

	// Poll ETH node
	pollerCli := tokenapi.NewZRPC(config.Zconf.EthNode, "BlocksPoller")
	go poller.BlocksPoller(txBlocksChan, cnBlocksChan, evBlocksChan, pollerCli, psqlClient, config.Zconf.BlocksDelay)

	// Watch a Transaction
	watApi := tokenapi.New(tokenapi.NewZRPC(config.Zconf.EthNode, "Watch a Transaction"))
	go matcher.TxMatcher(txBlocksChan, matchesChan, psqlClient, watApi)

	// Watch a Contract
	wacApi := tokenapi.New(tokenapi.NewZRPC(config.Zconf.EthNode, "Watch a Contract"))
	go matcher.ContractMatcher(cnBlocksChan, matchesChan, psqlClient, wacApi)

	// Watch an Event
	waeApi := tokenapi.New(tokenapi.NewZRPC(config.Zconf.EthNode, "Watch an Event"))
	go matcher.EventMatcher(evBlocksChan, matchesChan, psqlClient, waeApi)

	// Cron Triggers
	cronApi := tokenapi.New(tokenapi.NewZRPC(config.Zconf.BackupNode, "Cron Trig"))
	go matcher.CronScheduler(psqlClient, cronApi, matchesChan)

	// Main routine - process matches
	for {
		match := <-matchesChan
		go matcher.ProcessMatch(match, psqlClient, sesSession, &httpClient)
	}
}
