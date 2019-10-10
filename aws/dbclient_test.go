package aws

import (
	"log"
	"testing"
	"zoroaster/config"
	"zoroaster/trigger"
)

var psqlClient = PostgresClient{}
var zconf = config.Load("../config")

func init() {
	if zconf.Stage != "DEV" {
		log.Fatal("$STAGE must be DEV to run db tests")
	}
	psqlClient.InitDB(zconf)
}

func TestPostgresClient_All(t *testing.T) {
	// TODO figure out how Go does teardown so I can split these tests;
	// for now I can't be bothered and I'll fit everything in one test,
	// closing the connection only once, at the end.

	// Also note that these tests they are, at best, asserting for non-errors.
	// The way I'm using them is to run them as a stand-alone module and see
	// what they return.
	// In the future it would be nice to have some real assertions;
	// we would need to populate a database and have asserts on the returned values.

	defer psqlClient.Close()

	// Log Tx Match
	tg, _ := trigger.NewTriggerFromFile("../resources/triggers/t1.json")
	tx := trigger.GetTransactionFromFile("../resources/transactions/tx1.json")
	fnArgs := "{}"
	ztx := trigger.ZTransaction{
		BlockTimestamp: 1554828248,
		DecodedFnName:  &fnArgs,
		DecodedFnArgs:  &fnArgs,
		Tx:             tx,
	}
	txMatch := trigger.TxMatch{
		MatchId: 1,
		Tg:      tg,
		ZTx:     &ztx,
	}
	psqlClient.LogMatch(txMatch)

	// Log Contract Match

	cnMatch := trigger.CnMatch{
		Trigger:        tg,
		BlockNo:        1,
		BlockTimestamp: 888888,
		BlockHash:      "0x",
		MatchId:        1,
		MatchedValues:  "{}",
		AllValues:      "{}",
	}
	psqlClient.LogMatch(cnMatch)

	// Update Matching Triggers
	psqlClient.UpdateMatchingTriggers([]int{21, 31})

	// Update Non-Matching Triggers
	psqlClient.UpdateNonMatchingTriggers([]int{21, 31})

	// Log Outcomes
	payload := `{"BlockNo":8888,"BlockTimestamp":1554828248,"ReturnedValue":"matched values","AllValues":"all values"}`
	outcome := `{"StatusCode":200}`
	o1 := trigger.Outcome{payload, outcome}
	psqlClient.LogOutcome(&o1, 1)

	// Load all the active triggers
	_, err := psqlClient.LoadTriggersFromDB("WatchTransactions")
	if err != nil {
		t.Error(err)
	}

	// Get all the active actions
	_, err = psqlClient.GetActions(34, 1)
	if err != nil {
		t.Error(err)
	}
}
