package tests

import (
	"encoding/json"
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"testing"
	trig "zoroaster/triggers"
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

func TestIntegration(t *testing.T) {

	client := ethrpc.New("https://ethshared.bdnodes.net/?auth=_M92hYFzHxR4S1kNbYHfR6ResdtDRqvvLdnm3ZcdAXA")

	data, err := ioutil.ReadFile("rules.json")
	if err != nil {
		log.Fatal(err)
	}

	allRules := AllRules{}
	err = json.Unmarshal(data, &allRules)
	if err != nil {
		log.Fatal(err)
	}

	doneChan := make(chan bool)
	for _, r := range allRules.Rules {
		go Run(doneChan, client, r, t)
	}
	<-doneChan
}

func Run(done chan bool, client *ethrpc.EthRPC, r Rule, t *testing.T) {
	trigger, err := trig.NewTriggerFromFile("triggers/" + r.TriggerFile)
	if err != nil {
		log.Fatal(err)
	}
	block, err := client.EthGetBlockByNumber(r.BlockNo, true)
	if err != nil {
		log.Fatal(err)
	}
	txs := trig.MatchTrigger(trigger, block)

	if len(txs) != r.Occurrences {
		t.Error(r.Assert)
	}
	done <- true
}
