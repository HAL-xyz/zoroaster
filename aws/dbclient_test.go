package aws

import (
	"testing"
	"zoroaster/triggers"
)

func TestLogMatch(t *testing.T) {
	InitDB()
	block := trigger.GetBlockFromFile("../resources/blocks/block1.json")
	tg, _ := trigger.NewTriggerFromFile("../resources/triggers/t2.json")

	LogMatch(tg, &block.Transactions[0], "test_trigger_log")

	_, err := db.Exec("DELETE FROM TEST_trigger_log")
	if err != nil {
		t.Error(err)
	}
}
