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

func (db mockDB) LogOutcome(table string, outcome *trigger.Outcome, matchId int) {
	panic("implement me")
}

func (db mockDB) GetActions(table string, tgId int, userId int) ([]string, error) {
	panic("implement me")
}

func (db mockDB) ReadLastBlockProcessed(table string, watOrWac string) int {
	panic("implement me")
}

func (db mockDB) SetLastBlockProcessed(table string, blockNo int, watOrWac string) {

}

func (db mockDB) LogTxMatch(table string, match trigger.TxMatch) int {
	panic("implement me")
}

func (db mockDB) LogCnMatch(table string, match trigger.CnMatch) int {
	panic("implement me")
}

func (db mockDB) LoadTriggersFromDB(table string, watOrWac string) ([]*trigger.Trigger, error) {
	tg, _ := trigger.NewTriggerFromFile("../resources/triggers/wac1.json")
	return []*trigger.Trigger{tg}, nil
}

func TestMatchContractsForBlock(t *testing.T) {

	// mocks
	mockGetModAccounts := func(a, b int) []string {
		return []string{"0xbb9bc244d798123fde783fcc1c72d3bb8c189413"}
	}

	var client = ethrpc.New("https://ethshared.bdnodes.net/?auth=_M92hYFzHxR4S1kNbYHfR6ResdtDRqvvLdnm3ZcdAXA")

	zconf := config.Load("../config")

	cnMatches := MatchContractsForBlock(8081000, mockGetModAccounts, zconf, mockDB{}, client)

	assert.Equal(t, len(cnMatches), 1)
	assert.Equal(t, cnMatches[0].BlockNo, 8081000)
}
