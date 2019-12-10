package matcher

import (
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"zoroaster/aws"
	"zoroaster/config"
	"zoroaster/trigger"
)

var zconf = config.Load("../config")
var psqlClient = aws.PostgresClient{}

func init() {
	if zconf.Stage != config.TEST {
		log.Fatal("$STAGE must be TEST to run tests")
	}
	psqlClient.InitDB(zconf)
	log.SetLevel(log.DebugLevel)
}

type mockDB struct {
	aws.IDB
}

func (db mockDB) SetLastBlockProcessed(blockNo int, tgType trigger.TgType) error {
	return nil
}

func (db mockDB) LoadTriggersFromDB(tgType trigger.TgType) ([]*trigger.Trigger, error) {
	tg, _ := trigger.GetTriggerFromFile("../resources/triggers/wac1.json")
	tg.TriggerUUID = "some-complicated-uuid"
	return []*trigger.Trigger{tg}, nil
}

func (db mockDB) UpdateMatchingTriggers(triggerIds []string) {
	// void
}

func (db mockDB) UpdateNonMatchingTriggers(triggerIds []string) {
	// void
}

func (db mockDB) GetSilentButMatchingTriggers(triggerUUIDs []string) []string {
	return []string{"some-complicated-uuid"}
}

func TestMatchContractsForBlock(t *testing.T) {

	// mocks
	mockGetModAccounts := func(a, b int, node string) []string {
		return []string{"0xbb9bc244d798123fde783fcc1c72d3bb8c189413"}
	}

	var client = ethrpc.New(zconf.EthNode)

	cnMatches := matchContractsForBlock(
		8081000,
		1554828248,
		"0x",
		mockGetModAccounts,
		mockDB{},
		client)

	assert.Equal(t, 1, len(cnMatches))
	assert.Equal(t, 8081000, cnMatches[0].BlockNumber)
}

func TestMatchContractsWithRealDB(t *testing.T) {

	// clear up the database
	err := psqlClient.TruncateTables([]string{"triggers", "matches"})
	assert.NoError(t, err)

	// load one trigger
	triggerSrc, err := ioutil.ReadFile("../resources/triggers/wac-uniswap.json")
	assert.NoError(t, err)
	err = psqlClient.SaveTrigger(string(triggerSrc), true, false)
	assert.NoError(t, err)

	mockGetModAccounts := func(a, b int, node string) []string {
		return []string{"0x09cabec1ead1c0ba254b09efb3ee13841712be14"}
	}

	// TODO mock this
	var client = ethrpc.New(zconf.EthNode)

	// this should match
	cnMatches := matchContractsForBlock(
		8081000,
		1554828248,
		"0x",
		mockGetModAccounts,
		&psqlClient,
		client)

	assert.Equal(t, 1, len(cnMatches))

	for _, m := range cnMatches {
		uuid, err := psqlClient.LogMatch(m)
		assert.NoError(t, err)
		assert.Len(t, uuid, 36)
	}

	// subsequent calls won't match, because triggered is set to true
	cnMatches = matchContractsForBlock(
		8081000,
		1554828248,
		"0x",
		mockGetModAccounts,
		&psqlClient,
		client)

	assert.Equal(t, 0, len(cnMatches))
}
