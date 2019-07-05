package matcher

import (
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"testing"
	"zoroaster/trigger"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

type mockDB struct{}

func (db mockDB) LoadTriggersFromDB(table string) ([]*trigger.Trigger, error) {
	tg, _ := trigger.NewTriggerFromFile("../resources/triggers/wac1.json")
	return []*trigger.Trigger{tg}, nil
}

func TestMatchContractsForBlock(t *testing.T) {

	// mocks
	mockGetModAccounts := func(a, b int) []string {
		return []string{"0xbb9bc244d798123fde783fcc1c72d3bb8c189413"}
	}

	var client = ethrpc.New("https://ethshared.bdnodes.net/?auth=_M92hYFzHxR4S1kNbYHfR6ResdtDRqvvLdnm3ZcdAXA")

	MatchContractsForBlock(8081000, mockGetModAccounts, "dev_triggers", mockDB{}, client)
}
