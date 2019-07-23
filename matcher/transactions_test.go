package matcher

import (
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"zoroaster/config"
	"zoroaster/trigger"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

type mockDB struct{}

func (db mockDB) InitDB(c *config.ZConfiguration) {
	panic("implement me")
}

func (db mockDB) Close() {
	panic("implement me")
}

func (db mockDB) LogOutcome(outcome *trigger.Outcome, matchId int) {
	panic("implement me")
}

func (db mockDB) GetActions(tgId int, userId int) ([]string, error) {
	panic("implement me")
}

func (db mockDB) ReadLastBlockProcessed(watOrWac string) int {
	panic("implement me")
}

func (db mockDB) SetLastBlockProcessed(blockNo int, watOrWac string) {

}

func (db mockDB) LogTxMatch(match trigger.TxMatch) int {
	panic("implement me")
}

func (db mockDB) LogCnMatch(match trigger.CnMatch) int {
	panic("implement me")
}

func (db mockDB) LoadTriggersFromDB(watOrWac string) ([]*trigger.Trigger, error) {
	tg, _ := trigger.NewTriggerFromFile("../resources/triggers/wac1.json")
	return []*trigger.Trigger{tg}, nil
}

func (db mockDB) UpdateMatchingTriggers(triggerIds []int) {
}

func (db mockDB) UpdateNonMatchingTriggers(triggerIds []int) {
}

func TestMatchContractsForBlock(t *testing.T) {

	// mocks
	mockGetModAccounts := func(a, b int) []string {
		return []string{"0xbb9bc244d798123fde783fcc1c72d3bb8c189413"}
	}

	zconf := config.Load("../config")
	var client = ethrpc.New(zconf.EthNode)

	cnMatches := MatchContractsForBlock(8081000, mockGetModAccounts, mockDB{}, client)

	assert.Equal(t, len(cnMatches), 1)
	assert.Equal(t, cnMatches[0].BlockNo, 8081000)
}
