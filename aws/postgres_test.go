package aws

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
	"zoroaster/config"
	"zoroaster/trigger"
)

var psqlClient = PostgresClient{}

func init() {
	if config.Zconf.Stage != config.TEST {
		log.Fatal("$STAGE must be TEST to run db tests")
	}
	psqlClient.InitDB(config.Zconf)
}

func TestPostgresClient_All(t *testing.T) {
	// TODO figure out how Go does teardown so I can split these tests;
	// for now I can't be bothered and I'll fit everything in one test,
	// closing the connection only once, at the end.

	// Also note that these tests are, at best, asserting for non-errors.
	// In the future it would be nice to have some real assertions;
	// we would need to populate a database and have asserts on the returned values.

	defer psqlClient.Close()

	// Log Tx Match
	tg, _ := trigger.GetTriggerFromFile("../resources/triggers/t1.json")
	tg.TriggerUUID = "3b29b0c3-e403-4103-81ef-6685cd391cda"
	tx, err := trigger.GetTransactionFromFile("../resources/transactions/tx1.json")
	assert.NoError(t, err)
	fnArgs := "{}"
	txMatch := trigger.TxMatch{
		MatchUUID:      "3b29b0c3-e403-4103-81ef-6685cd391cdm",
		BlockTimestamp: 1554828248,
		Tg:             tg,
		DecodedFnName:  &fnArgs,
		DecodedFnArgs:  &fnArgs,
		Tx:             tx,
	}
	_, err = psqlClient.LogMatch(txMatch)
	assert.NoError(t, err)

	// Log Contract Match
	cnMatch := trigger.CnMatch{
		Trigger:        tg,
		BlockNumber:    1,
		BlockTimestamp: 888888,
		BlockHash:      "0x",
		MatchUUID:      "3b29b0c3-e403-4103-81ef-6685cd391cdm",
		MatchedValues:  []string{},
		AllValues:      nil,
	}
	_, err = psqlClient.LogMatch(cnMatch)
	assert.NoError(t, err)

	// Log Event Match
	logs, _ := trigger.GetLogsFromFile("../resources/events/logs1.json")
	eventMatch := trigger.EventMatch{
		MatchUUID:      "3b29b0c3-e403-4103-81ef-6685cd391cdm",
		Tg:             tg,
		Log:            &logs[0],
		EventParams:    map[string]interface{}{},
		BlockTimestamp: 888888,
	}
	_, err = psqlClient.LogMatch(eventMatch)
	assert.NoError(t, err)

	// Update Matching Triggers
	psqlClient.UpdateMatchingTriggers([]string{"3b29b0c3-e403-4103-81ef-6685cd391cda"})

	// Update Non-Matching Triggers
	psqlClient.UpdateNonMatchingTriggers([]string{"3b29b0c3-e403-4103-81ef-6685cd391cda"})

	// Log Outcomes
	payload := `{ "BlockNumber": 1, "ContractAdd": "0x", "FunctionName": "fn()", "ReturnedData": { "AllValues": "{}", "MatchedValues": "{}" }, "BlockTimestamp": 8888 }`
	outcome := `{"HttpCode":200}`
	o1 := trigger.Outcome{payload, outcome}
	psqlClient.LogOutcome(&o1, "3b29b0c3-e403-4103-81ef-6685cd391cda")

	// Load all the active triggers
	_, err = psqlClient.LoadTriggersFromDB(trigger.WaT)
	assert.NoError(t, err)

	// Get all the active actions
	_, err = psqlClient.GetActions("3b29b0c3-e403-4103-81ef-6685cd391cda", "3b29b0c3-e403-4103-81ef-6685cd391cde")
	assert.NoError(t, err)

	// Get all silent but matching triggers
	_ = psqlClient.GetSilentButMatchingTriggers([]string{"96f195bd-2fc7-491e-b1a0-b19dca514cb0"})

	// Set and read app state
	err = psqlClient.SetLastBlockProcessed(0, trigger.WaT)
	assert.NoError(t, err)
	blockNo, err := psqlClient.ReadLastBlockProcessed(trigger.WaT)
	assert.NoError(t, err)
	assert.Equal(t, 0, blockNo)

	err = psqlClient.SetLastBlockProcessed(0, trigger.WaC)
	assert.NoError(t, err)
	blockNo, err = psqlClient.ReadLastBlockProcessed(trigger.WaC)
	assert.NoError(t, err)
	assert.Equal(t, 0, blockNo)

	err = psqlClient.SetLastBlockProcessed(0, trigger.WaE)
	assert.NoError(t, err)
	blockNo, err = psqlClient.ReadLastBlockProcessed(trigger.WaE)
	assert.NoError(t, err)
	assert.Equal(t, 0, blockNo)

	// Write analytics
	err = psqlClient.LogAnalytics(trigger.WaT, 9999, 100, int(time.Now().Unix()), time.Now(), time.Now().Add(10*time.Second))
	assert.NoError(t, err)
}
