package matcher

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"zoroaster/aws"
	"zoroaster/config"
	"zoroaster/trigger"
)

var psqlClient = aws.PostgresClient{}

func init() {
	if config.Zconf.Stage != config.TEST {
		log.Fatal("$STAGE must be TEST to run tests")
	}
	psqlClient.InitDB(config.Zconf)
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
	tg.ContractAdd = "0xBB9bc244d798123fde783fCC1c72d3bb8c189413" // uppercase some letters
	return []*trigger.Trigger{tg}, nil
}

func (db mockDB) UpdateMatchingTriggers(triggerIds []string) {
	// void
}

func (db mockDB) UpdateNonMatchingTriggers(triggerIds []string) {
	// void
}

func (db mockDB) GetSilentButMatchingTriggers(triggerUUIDs []string) ([]string, error) {
	return []string{"some-complicated-uuid"}, nil
}

func TestMatchContractsForBlock(t *testing.T) {

	// mocks
	mockGetModAccounts := func(a, b int, node string) []string {
		return []string{"0xbb9bc244d798123fde783fcc1c72d3bb8c189413"}
	}
	lastBlock, err := config.CliMain.EthBlockNumber()
	assert.NoError(t, err)

	cnMatches := matchContractsForBlock(
		lastBlock,
		1554828248,
		"0x",
		mockGetModAccounts,
		mockDB{},
		config.CliMain)

	assert.Equal(t, 1, len(cnMatches))
	assert.Equal(t, lastBlock, cnMatches[0].BlockNumber)
}

func TestMatchContractsWithRealDB(t *testing.T) {

	// clear up the database
	err := psqlClient.TruncateTables([]string{"triggers", "matches"})
	assert.NoError(t, err)
	if err != nil {
		log.Fatal(err)
	}

	// load a User
	userUUID, err := psqlClient.SaveUser(100, 0)
	assert.NoError(t, err)

	// load a Trigger
	triggerSrc, err := ioutil.ReadFile("../resources/triggers/wac-uniswap.json")
	assert.NoError(t, err)
	triggerUUID, err := psqlClient.SaveTrigger(string(triggerSrc), true, false, userUUID)
	assert.NoError(t, err)

	// at creation, triggered=false
	status, err := psqlClient.ReadString(fmt.Sprintf("SELECT triggered FROM triggers WHERE uuid = '%s'", triggerUUID))
	assert.NoError(t, err)
	assert.Equal(t, status, "false")

	mockGetModAccounts := func(a, b int, node string) []string {
		return []string{"0x09cabec1ead1c0ba254b09efb3ee13841712be14"}
	}

	lastBlock, err := config.CliMain.EthBlockNumber()
	assert.NoError(t, err)

	// here we call getTokenToEthOutputPrice(1) which returns the
	// current ETH price in USD; since the trigger condition is "biggerThan 3"
	// we expect this trigger to always match
	cnMatches := matchContractsForBlock(
		lastBlock,
		1554828248,
		"0x",
		mockGetModAccounts,
		&psqlClient,
		config.CliMain)

	assert.Equal(t, 1, len(cnMatches))

	// now trigger status will be triggered=true
	status, err = psqlClient.ReadString(fmt.Sprintf("SELECT triggered FROM triggers WHERE uuid = '%s'", triggerUUID))
	assert.NoError(t, err)
	assert.Equal(t, status, "true")

	for _, m := range cnMatches {
		uuid, err := psqlClient.LogMatch(m)
		assert.NoError(t, err)
		assert.Len(t, uuid, 36)
	}

	// subsequent calls won't match, because triggered is set to true
	cnMatches = matchContractsForBlock(
		lastBlock,
		1554828248,
		"0x",
		mockGetModAccounts,
		&psqlClient,
		config.CliMain)

	assert.Equal(t, 0, len(cnMatches))

	// TODO mock the eth client to return a value that changes the trigger status again
}
