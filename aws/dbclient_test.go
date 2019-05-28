package aws

import (
	"testing"
	"zoroaster/config"
	"zoroaster/triggers"
)

func TestLogMatch(t *testing.T) {
	zconf := config.Load()
	zconf.TriggersDB.TableLogs = "test_trigger_log"

	InitDB(zconf)

	block := trigger.GetBlockFromFile("../resources/blocks/block1.json")
	tg, _ := trigger.NewTriggerFromFile("../resources/triggers/t2.json")

	LogMatch(tg, &block.Transactions[0], zconf.TriggersDB.TableLogs)

	_, err := db.Exec("DELETE FROM TEST_trigger_log")
	if err != nil {
		t.Error(err)
	}
}
