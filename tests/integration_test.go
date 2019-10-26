package tests

import (
	"encoding/json"
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"sync"
	"testing"
	"zoroaster/config"
	trig "zoroaster/trigger"
)

type AllRules struct {
	Rules []Rule
}
type Rule struct {
	Assert      string `json:"assert"`
	Occurrences int    `json:"occurrences"`
	BlockNo     int    `json:"block_no"`
	TriggerFile string `json:"trigger_file"`
}

var zconf = config.Load("../config")

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestIntegration(t *testing.T) {

	var client = ethrpc.New(zconf.EthNode)

	data, err := ioutil.ReadFile("rules.json")
	if err != nil {
		log.Fatal(err)
	}

	allRules := AllRules{}
	err = json.Unmarshal(data, &allRules)
	if err != nil {
		log.Fatal(err)
	}

	const MAX = 5
	sem := make(chan int, MAX)
	var wg sync.WaitGroup

	for _, r := range allRules.Rules {
		sem <- 1
		wg.Add(1)
		go func(rule Rule) {
			defer wg.Done()
			trigger, err := trig.NewTriggerFromFile("triggers/" + rule.TriggerFile)
			if err != nil {
				log.Fatal(err)
			}
			block, err := client.EthGetBlockByNumber(rule.BlockNo, true)
			if err != nil {
				log.Fatal(err)
			}
			txs := trig.MatchTransaction(trigger, block)

			if len(txs) != rule.Occurrences {
				log.Errorf("%s failed (expected %d, got %d instead)\n", rule.Assert, rule.Occurrences, len(txs))
				t.Error()
			}
			<-sem
		}(r)
	}
	wg.Wait()
}
