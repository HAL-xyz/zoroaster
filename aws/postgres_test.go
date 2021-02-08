package aws

import (
	"fmt"
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/HAL-xyz/zoroaster/trigger"
	"github.com/HAL-xyz/zoroaster/utils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

var psqlClient = NewPostgresClient(config.Zconf)

func init() {
	if config.Zconf.Stage != config.TEST {
		log.Fatal("$STAGE must be TEST to run db tests")
	}
	if config.Zconf.Database.Network != "1_eth_mainnet" {
		log.Fatal("$NETWORK must be 1_eth_mainnet to run tests ")
	}
}

func TestPostgresClient_All(t *testing.T) {
	// TODO figure out how Go does teardown so I can split these tests;
	// for now I can't be bothered and I'll fit everything in one test,
	// closing the connection only once, at the end.

	defer psqlClient.Close()

	// clear up the database
	err := psqlClient.TruncateTables([]string{"triggers", "matches", "users", "state"})
	assert.NoError(t, err)

	// load Networks
	err = psqlClient.SetString("INSERT INTO networks(network_id_name, friendly_name, technology, network_name, endpoint) VALUES ('1_eth_mainnet', '', '', '', '') ON CONFLICT (network_id_name) DO NOTHING")
	err = psqlClient.SetString("INSERT INTO networks(network_id_name, friendly_name, technology, network_name, endpoint) VALUES ('2_eth_rinkeby', '', '', '', '') ON CONFLICT (network_id_name) DO NOTHING")
	assert.NoError(t, err)

	// load States
	err = psqlClient.SetString("INSERT INTO state(id, network_id) VALUES (1, '1_eth_mainnet')")
	err = psqlClient.SetString("INSERT INTO state(id, network_id, wat_last_block_processed) VALUES (2, '2_eth_rinkeby', 10)")

	// load a User
	userUUID, err := psqlClient.SaveUser(1000, 1)
	assert.NoError(t, err)

	// load two Triggers, one in 1_eth_mainnet (default network), one in 2_eth_rinkeby
	triggerSrc, err := ioutil.ReadFile("../resources/triggers/wac-uniswap.json")
	assert.NoError(t, err)
	triggerUUID, err := psqlClient.SaveTrigger(string(triggerSrc), true, false, userUUID, "1_eth_mainnet")
	assert.NoError(t, err)
	_, err = psqlClient.SaveTrigger(string(triggerSrc), true, false, userUUID, "2_eth_rinkeby")
	assert.NoError(t, err)

	// Load all the active triggers
	tgs, err := psqlClient.LoadTriggersFromDB(trigger.WaC)
	assert.NoError(t, err)
	assert.Len(t, tgs, 1) // only 1 bc the tests run with network = 1_eth_mainnet

	// load two Actions
	_, err = psqlClient.SaveAction(triggerUUID)
	_, err = psqlClient.SaveAction(triggerUUID)
	assert.NoError(t, err)

	// Log Tx Match
	tg, _ := trigger.GetTriggerFromFile("../resources/triggers/t1.json")
	tg.TriggerUUID = triggerUUID
	tg.UserUUID = userUUID

	tx, err := trigger.GetTransactionFromFile("../resources/transactions/tx1.json")
	assert.NoError(t, err)
	fnName := ""
	txMatch := trigger.TxMatch{
		MatchUUID:      "3b29b0c3-e403-4103-81ef-6685cd391cdm",
		BlockTimestamp: 1554828248,
		Tg:             tg,
		DecodedFnName:  &fnName,
		DecodedFnArgs:  map[string]interface{}{},
		Tx:             tx,
	}
	err = psqlClient.LogMatch(&txMatch)
	assert.NoError(t, err)
	assert.Len(t, txMatch.MatchUUID, 36)

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
	err = psqlClient.LogMatch(&cnMatch)
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
	err = psqlClient.LogMatch(&eventMatch)
	assert.NoError(t, err)

	// Update Matching Triggers: set triggered=true
	psqlClient.UpdateMatchingTriggers([]string{triggerUUID})
	triggered, err := psqlClient.ReadString(fmt.Sprintf("SELECT triggered FROM triggers WHERE uuid = '%s'", triggerUUID))
	assert.NoError(t, err)
	assert.Equal(t, triggered, "true")

	// Update Non-Matching Triggers: set triggered=false
	psqlClient.UpdateNonMatchingTriggers([]string{triggerUUID})
	triggered, err = psqlClient.ReadString(fmt.Sprintf("SELECT triggered FROM triggers WHERE uuid = '%s'", triggerUUID))
	assert.NoError(t, err)
	assert.Equal(t, triggered, "false")

	// Get all silent but matching triggers
	// if run after Update Non-Matching Triggers will find one trigger
	silent, err := psqlClient.GetSilentButMatchingTriggers([]string{triggerUUID})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(silent))

	// Log Outcomes
	payload := `{
  "BlockNumber":1,
  "ContractAdd":"0x",
  "FunctionName":"fn()",
  "ReturnedData":{
     "AllValues":"{}",
     "MatchedValues":"{}"
  },
  "BlockTimestamp":8888
}`
	outcome := `{"HttpCode":200}`
	o1 := trigger.Outcome{payload, outcome, true}
	err = psqlClient.LogOutcome(&o1, cnMatch.MatchUUID)
	assert.NoError(t, err)

	// Get all the active actions
	actions, err := psqlClient.GetActions(triggerUUID, userUUID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(actions))

	// Set and read app state
	err = psqlClient.SetLastBlockProcessed(99, trigger.WaT)
	assert.NoError(t, err)
	blockNo, err := psqlClient.ReadLastBlockProcessed(trigger.WaT)
	assert.NoError(t, err)
	assert.Equal(t, 99, blockNo)

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

	// make sure the state has NOT been changed for other networks
	v, err := psqlClient.ReadString("SELECT wat_last_block_processed FROM state WHERE network_id='2_eth_rinkeby'")
	assert.NoError(t, err)
	assert.Equal(t, "10", v) // this was set at the beginning

	// Write analytics
	err = psqlClient.LogAnalytics(trigger.WaT, 9999, 100, int(time.Now().Unix()), time.Now(), time.Now().Add(10*time.Second))
	assert.NoError(t, err)

	/* Test user limits */

	// load a User
	batmanUUID, err := psqlClient.SaveUser(1, 0)
	assert.NoError(t, err)

	// load a Trigger
	batmanTriggerSrc, err := ioutil.ReadFile("../resources/triggers/wac-uniswap.json")
	assert.NoError(t, err)
	batmanTriggerUUID, err := psqlClient.SaveTrigger(string(batmanTriggerSrc), true, false, batmanUUID, "1_eth_mainnet")
	assert.NoError(t, err)

	activeTriggers, err := psqlClient.LoadTriggersFromDB(trigger.WaC)
	assert.NoError(t, err)

	// check that batmanTriggerUUID is present within activeTriggers
	uuids := make([]string, len(activeTriggers))
	for _, tg := range activeTriggers {
		uuids = append(uuids, tg.TriggerUUID)
	}
	assert.True(t, utils.IsIn(batmanTriggerUUID, uuids))

	// fetch batman's trigger
	var batmanTrigger *trigger.Trigger
	for _, tg := range activeTriggers {
		if tg.UserUUID == batmanUUID {
			batmanTrigger = tg
		}
	}

	assert.NotNil(t, batmanTrigger.UserUUID)

	// when logging a match for batmanTriggerUUID, the batman's counter should ++
	batmanMatch := trigger.CnMatch{
		Trigger:        batmanTrigger,
		BlockNumber:    1,
		BlockTimestamp: 888888,
		BlockHash:      "0x",
		MatchUUID:      "3b29b0c3-e403-4103-81ef-6685cd391cdm",
		MatchedValues:  []string{},
		AllValues:      nil,
	}
	err = psqlClient.LogMatch(&batmanMatch)
	assert.NoError(t, err)

	newCounter, err := psqlClient.ReadString(fmt.Sprintf("SELECT counter_current_month FROM users WHERE uuid = '%s'", batmanTrigger.UserUUID))
	assert.NoError(t, err)
	assert.Equal(t, "1", newCounter)

	// now this trigger should not be loaded anymore
	activeTriggers, err = psqlClient.LoadTriggersFromDB(trigger.WaC)
	assert.NoError(t, err)

	// check that batmanTriggerUUID is *NOT* present within activeTriggers
	uuids = make([]string, len(activeTriggers))
	for _, tg := range activeTriggers {
		uuids = append(uuids, tg.TriggerUUID)
	}
	assert.False(t, utils.IsIn(batmanTriggerUUID, uuids))
}
