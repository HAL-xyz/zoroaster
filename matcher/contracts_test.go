package matcher

import (
	"fmt"
	"github.com/HAL-xyz/ethrpc"
	"github.com/HAL-xyz/zoroaster/aws"
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/HAL-xyz/zoroaster/rpc"
	"github.com/HAL-xyz/zoroaster/trigger"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

var psqlClient = aws.NewPostgresClient(config.Zconf)

func init() {
	if config.Zconf.Stage != config.TEST {
		log.Fatal("$STAGE must be TEST to run tests")
	}
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

	lastBlock, err := config.CliMain.EthBlockNumber()
	assert.NoError(t, err)

	cnMatches := matchContractsForBlock(lastBlock, 1554828248, "0x", mockDB{}, config.CliMain)

	assert.Equal(t, 1, len(cnMatches))
	assert.Equal(t, lastBlock, cnMatches[0].BlockNumber)
	fmt.Println(cnMatches[0].AllValues)
}

// ETHRPC Client mock, returns 189
type mockETHCli struct {
	rpc.IEthRpc
}

func (cli mockETHCli) EthCall(transaction ethrpc.T, tag string) (string, error) {
	return "0x00000000000000000000000000000000000000000000000000000000000000bd0000000100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", nil
}

// ETHRPC Client mock, returns 378
type mockETHCliNoMatch struct {
	rpc.IEthRpc
}

func (cli mockETHCliNoMatch) EthCall(transaction ethrpc.T, tag string) (string, error) {
	return "0x000000000000000000000000000000000000000000000000000000000000017a0000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", nil
}

// ETHRPC Client mock, returns an error
type mockETHCliWithError struct {
	rpc.IEthRpc
}

func (cli mockETHCliWithError) EthCall(transaction ethrpc.T, tag string) (string, error) {
	return "0x0", fmt.Errorf("some nasty error")
}

func TestMatchContractsWithRealDB(t *testing.T) {

	// clear up the database
	err := psqlClient.TruncateTables([]string{"triggers", "matches"})
	assert.NoError(t, err)

	// load a User
	userUUID, err := psqlClient.SaveUser(100, 0)
	assert.NoError(t, err)

	// load two Trigger, different networks
	triggerSrc, err := ioutil.ReadFile("../resources/triggers/wac-uniswap.json")
	assert.NoError(t, err)
	triggerUUID, err := psqlClient.SaveTrigger(string(triggerSrc), true, false, userUUID, "1_eth_mainnet")
	assert.NoError(t, err)
	rinkebyUUID, err := psqlClient.SaveTrigger(string(triggerSrc), true, false, userUUID, "2_eth_rinkeby")
	assert.NoError(t, err)

	// at creation, triggered=false
	status, err := psqlClient.ReadString(fmt.Sprintf("SELECT triggered FROM triggers WHERE uuid = '%s'", triggerUUID))
	assert.NoError(t, err)
	assert.Equal(t, "false", status)

	ethSuccessMock := mockETHCli{}

	// success
	cnMatches := matchContractsForBlock(0000, 1554828248, "0x", psqlClient, ethSuccessMock)
	assert.Equal(t, 1, len(cnMatches))

	// now trigger status will be triggered=true
	status, err = psqlClient.ReadString(fmt.Sprintf("SELECT triggered FROM triggers WHERE uuid = '%s'", triggerUUID))
	assert.NoError(t, err)
	assert.Equal(t, "true", status)

	// ... but only for the trigger on Mainnet; the trigger on Rinkeby should still be false
	status, err = psqlClient.ReadString(fmt.Sprintf("SELECT triggered FROM triggers WHERE uuid = '%s'", rinkebyUUID))
	assert.NoError(t, err)
	assert.Equal(t, "false", status)

	// just write the matches to the db
	for _, m := range cnMatches {
		uuid, err := psqlClient.LogMatch(m)
		assert.NoError(t, err)
		assert.Len(t, uuid, 36)
	}

	// subsequent calls won't match, because triggered is set to true
	cnMatches = matchContractsForBlock(0000, 1554828248, "0x", psqlClient, ethSuccessMock)
	assert.Equal(t, 0, len(cnMatches))

	// trigger is still set to true
	status, err = psqlClient.ReadString(fmt.Sprintf("SELECT triggered FROM triggers WHERE uuid = '%s'", triggerUUID))
	assert.NoError(t, err)
	assert.Equal(t, "true", status)

	// ... and the Rinkeby trigger is still set to false
	status, err = psqlClient.ReadString(fmt.Sprintf("SELECT triggered FROM triggers WHERE uuid = '%s'", rinkebyUUID))
	assert.NoError(t, err)
	assert.Equal(t, "false", status)

	// the rpc call fails and MatchContracts() returns an error; triggered remains true
	ethErrorMock := mockETHCliWithError{}
	cnMatches = matchContractsForBlock(0000, 1554828248, "0x", psqlClient, ethErrorMock)
	assert.Equal(t, 0, len(cnMatches))

	status, err = psqlClient.ReadString(fmt.Sprintf("SELECT triggered FROM triggers WHERE uuid = '%s'", triggerUUID))
	assert.NoError(t, err)
	assert.Equal(t, "true", status)

	// ... and the Rinkeby trigger is still set to false
	status, err = psqlClient.ReadString(fmt.Sprintf("SELECT triggered FROM triggers WHERE uuid = '%s'", rinkebyUUID))
	assert.NoError(t, err)
	assert.Equal(t, "false", status)

	// now the eth cli returns a non-matching value, matches should be zero, triggered=false
	ethNoMatchMock := mockETHCliNoMatch{}
	cnMatches = matchContractsForBlock(0000, 1554828248, "0x", psqlClient, ethNoMatchMock)
	assert.Equal(t, 0, len(cnMatches))

	status, err = psqlClient.ReadString(fmt.Sprintf("SELECT triggered FROM triggers WHERE uuid = '%s'", triggerUUID))
	assert.NoError(t, err)
	assert.Equal(t, "false", status)

	// ... and the Rinkeby trigger is still set to false
	status, err = psqlClient.ReadString(fmt.Sprintf("SELECT triggered FROM triggers WHERE uuid = '%s'", rinkebyUUID))
	assert.NoError(t, err)
	assert.Equal(t, "false", status)

	// back to success, matches=1, triggered=true
	cnMatches = matchContractsForBlock(0000, 1554828248, "0x", psqlClient, ethSuccessMock)
	assert.Equal(t, 1, len(cnMatches))

	status, err = psqlClient.ReadString(fmt.Sprintf("SELECT triggered FROM triggers WHERE uuid = '%s'", triggerUUID))
	assert.NoError(t, err)
	assert.Equal(t, "true", status)

	// ... and the Rinkeby trigger is still set to false
	status, err = psqlClient.ReadString(fmt.Sprintf("SELECT triggered FROM triggers WHERE uuid = '%s'", rinkebyUUID))
	assert.NoError(t, err)
	assert.Equal(t, "false", status)
}
