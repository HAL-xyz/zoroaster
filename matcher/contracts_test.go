package matcher

import (
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"zoroaster/aws"
	"zoroaster/config"
	"zoroaster/trigger"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

type mockDB struct {
	aws.IDB
}

func (db mockDB) SetLastBlockProcessed(blockNo int, watOrWac string) {
	// void
}

func (db mockDB) LoadTriggersFromDB(watOrWac string) ([]*trigger.Trigger, error) {
	tg, _ := trigger.NewTriggerFromFile("../resources/triggers/wac1.json")
	return []*trigger.Trigger{tg}, nil
}

func (db mockDB) UpdateMatchingTriggers(triggerIds []int) {
	// void
}

func (db mockDB) UpdateNonMatchingTriggers(triggerIds []int) {
	// void
}

func TestMatchContractsForBlock(t *testing.T) {

	// mocks
	mockGetModAccounts := func(a, b int, node string) []string {
		return []string{"0xbb9bc244d798123fde783fcc1c72d3bb8c189413"}
	}

	zconf := config.Load("../config")
	var client = ethrpc.New(zconf.EthNode)

	cnMatches := matchContractsForBlock(8081000, 1554828248, mockGetModAccounts, mockDB{}, client)

	assert.Equal(t, len(cnMatches), 1)
	assert.Equal(t, cnMatches[0].BlockNo, 8081000)
}
