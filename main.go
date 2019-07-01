package main

import (
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
	"zoroaster/action"
	"zoroaster/aws"
	"zoroaster/config"
	"zoroaster/eth"
	"zoroaster/matcher"
	"zoroaster/trigger"
)

func main() {

	// Load Config
	zconf := config.Load()

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
	go matcher.TxMatcher(blocksChan, matchesChan, zconf)

	// Main routine - process actions
	for {
		match := <-matchesChan
		go func() {
			acts, err := aws.GetActions(zconf.TriggersDB.TableActions, match.Tg.TriggerId, match.Tg.UserId)
			if err != nil {
				log.Warnf("cannot get actions from db: %v", err)
			}
			log.Debugf("\tMatched %d actions", len(acts))
			eventJson := action.ActionEventJson{ZTx: match.ZTx, Actions: acts}
			outcomes := action.HandleEvent(eventJson, sesSession)
			for _, out := range outcomes {
				aws.LogOutcome(zconf.TriggersDB.TableOutcomes, out, match.MatchId)
				log.Debug("\tLogged outcome for match id ", match.MatchId)
			}
		}()
	}
}
