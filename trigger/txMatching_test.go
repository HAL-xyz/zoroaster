package trigger

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"zoroaster/config"
	"zoroaster/utils"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestValidateFilter1(t *testing.T) {
	block, _ := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := GetTriggerFromFile("../resources/triggers/t1.json")
	tid, abi, cnt := trigger.TriggerUUID, &trigger.ContractABI, trigger.ContractAdd

	// BasicFilter / To
	assert.Equal(t, validateFilter(&block.Transactions[0], &trigger.Filters[0], cnt, abi, tid), true)
	assert.Equal(t, validateFilter(&block.Transactions[1], &trigger.Filters[0], cnt, abi, tid), false)

	// BasicFilter / Nonce
	assert.Equal(t, validateFilter(&block.Transactions[0], &trigger.Filters[2], cnt, abi, tid), true)
}

func TestValidateFilter2(t *testing.T) {
	block, _ := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := GetTriggerFromFile("../resources/triggers/t1.json")
	tid, abi, cnt := trigger.TriggerUUID, &trigger.ContractABI, trigger.ContractAdd

	// Address
	assert.Equal(t, validateFilter(&block.Transactions[0], &trigger.Filters[1], cnt, abi, tid), true)
	assert.Equal(t, validateFilter(&block.Transactions[1], &trigger.Filters[1], cnt, abi, tid), false)
}

func TestValidateFilter3(t *testing.T) {
	block, _ := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := GetTriggerFromFile("../resources/triggers/t3.json")
	tid, abi, cnt := trigger.TriggerUUID, &trigger.ContractABI, trigger.ContractAdd

	// From
	assert.Equal(t, validateFilter(&block.Transactions[0], &trigger.Filters[0], cnt, abi, tid), true)
	assert.Equal(t, validateFilter(&block.Transactions[5], &trigger.Filters[0], cnt, abi, tid), false)
}

func TestValidateFilter4(t *testing.T) {
	block, _ := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := GetTriggerFromFile("../resources/triggers/t4.json")
	tid, abi, cnt := trigger.TriggerUUID, &trigger.ContractABI, trigger.ContractAdd

	// Value
	assert.Equal(t, validateFilter(&block.Transactions[2], &trigger.Filters[0], cnt, abi, tid), true)
	assert.Equal(t, validateFilter(&block.Transactions[0], &trigger.Filters[0], cnt, abi, tid), false)

	// Gas
	assert.Equal(t, validateFilter(&block.Transactions[0], &trigger.Filters[1], cnt, abi, tid), true)
	assert.Equal(t, validateFilter(&block.Transactions[5], &trigger.Filters[1], cnt, abi, tid), false)

	// GasPrice
	assert.Equal(t, validateFilter(&block.Transactions[7], &trigger.Filters[2], cnt, abi, tid), true)
	assert.Equal(t, validateFilter(&block.Transactions[4], &trigger.Filters[2], cnt, abi, tid), false)

	// Nonce
	assert.Equal(t, validateFilter(&block.Transactions[5], &trigger.Filters[3], cnt, abi, tid), true)
}

func TestValidateFilter5(t *testing.T) {
	tx, _ := GetTransactionFromFile("../resources/transactions/tx1.json")
	trigger, _ := GetTriggerFromFile("../resources/triggers/t5.json")
	tid, abi, cnt := trigger.TriggerUUID, &trigger.ContractABI, trigger.ContractAdd

	// uint256[]
	assert.Equal(t, validateFilter(tx, &trigger.Filters[0], cnt, abi, tid), false)
	assert.Equal(t, validateFilter(tx, &trigger.Filters[1], cnt, abi, tid), true)
	assert.Equal(t, validateFilter(tx, &trigger.Filters[2], cnt, abi, tid), false)

	// bytes14[]
	assert.Equal(t, validateFilter(tx, &trigger.Filters[3], cnt, abi, tid), true)
	assert.Equal(t, validateFilter(tx, &trigger.Filters[4], cnt, abi, tid), true)

	// Gas
	assert.Equal(t, validateFilter(tx, &trigger.Filters[5], cnt, abi, tid), true)

	// Nonce
	assert.Equal(t, validateFilter(tx, &trigger.Filters[6], cnt, abi, tid), true)
}

func TestValidateFilter6(t *testing.T) {
	tx, _ := GetTransactionFromFile("../resources/transactions/tx2.json")
	trigger, _ := GetTriggerFromFile("../resources/triggers/t6.json")
	tid, abi, cnt := trigger.TriggerUUID, &trigger.ContractABI, trigger.ContractAdd

	// address[N]
	assert.Equal(t, validateFilter(tx, &trigger.Filters[0], cnt, abi, tid), true)

	// uint256[N]
	assert.Equal(t, validateFilter(tx, &trigger.Filters[1], cnt, abi, tid), true)
	assert.Equal(t, validateFilter(tx, &trigger.Filters[2], cnt, abi, tid), true)
	assert.Equal(t, validateFilter(tx, &trigger.Filters[3], cnt, abi, tid), true)
}

func TestValidateFilter7(t *testing.T) {
	tx, _ := GetTransactionFromFile("../resources/transactions/tx3.json")
	trigger, _ := GetTriggerFromFile("../resources/triggers/t7.json")
	tid, abi, cnt := trigger.TriggerUUID, &trigger.ContractABI, trigger.ContractAdd

	// uint256
	assert.Equal(t, validateFilter(tx, &trigger.Filters[0], cnt, abi, tid), true)

	// bool
	assert.Equal(t, validateFilter(tx, &trigger.Filters[1], cnt, abi, tid), true)

	// int128
	assert.Equal(t, validateFilter(tx, &trigger.Filters[2], cnt, abi, tid), true)
}

func TestValidateFilter8(t *testing.T) {
	tx, _ := GetTransactionFromFile("../resources/transactions/tx4.json")
	trigger, _ := GetTriggerFromFile("../resources/triggers/t8.json")
	tid, abi, cnt := trigger.TriggerUUID, &trigger.ContractABI, trigger.ContractAdd

	// int128[N]
	assert.Equal(t, validateFilter(tx, &trigger.Filters[0], cnt, abi, tid), true)

	// int[N]
	assert.Equal(t, validateFilter(tx, &trigger.Filters[1], cnt, abi, tid), true)

	// int40
	assert.Equal(t, validateFilter(tx, &trigger.Filters[2], cnt, abi, tid), true)
}

func TestValidateFilter9(t *testing.T) {
	tx, _ := GetTransactionFromFile("../resources/transactions/tx5.json")
	trigger, _ := GetTriggerFromFile("../resources/triggers/t9.json")
	tid, abi, cnt := trigger.TriggerUUID, &trigger.ContractABI, trigger.ContractAdd

	// int32
	assert.Equal(t, validateFilter(tx, &trigger.Filters[0], cnt, abi, tid), true)

	// int32[]
	assert.Equal(t, validateFilter(tx, &trigger.Filters[1], cnt, abi, tid), true)

	// int32[6]
	assert.Equal(t, validateFilter(tx, &trigger.Filters[2], cnt, abi, tid), true)
	assert.Equal(t, validateFilter(tx, &trigger.Filters[10], cnt, abi, tid), true)

	// Index int32[]
	assert.Equal(t, validateFilter(tx, &trigger.Filters[7], cnt, abi, tid), true)
	assert.Equal(t, validateFilter(tx, &trigger.Filters[8], cnt, abi, tid), false)
}

func TestValidateFilter10(t *testing.T) {
	tx, _ := GetTransactionFromFile("../resources/transactions/tx6.json")
	trigger, _ := GetTriggerFromFile("../resources/triggers/t9.json")
	tid, abi, cnt := trigger.TriggerUUID, &trigger.ContractABI, trigger.ContractAdd

	// address[]
	assert.Equal(t, validateFilter(tx, &trigger.Filters[3], cnt, abi, tid), true)
	assert.Equal(t, validateFilter(tx, &trigger.Filters[4], cnt, abi, tid), true)

	// bytes1[]
	assert.Equal(t, validateFilter(tx, &trigger.Filters[5], cnt, abi, tid), true)

	// string[]
	assert.Equal(t, validateFilter(tx, &trigger.Filters[6], cnt, abi, tid), true)
	assert.Equal(t, validateFilter(tx, &trigger.Filters[9], cnt, abi, tid), false)
}

func TestValidateFilter11(t *testing.T) {
	// mute logging just for these tests to reduce noise
	log.SetLevel(log.WarnLevel)

	tx, _ := GetTransactionFromFile("../resources/transactions/tx5.json")
	trigger, _ := GetTriggerFromFile("../resources/triggers/t10.json")
	tid, abi, cnt := trigger.TriggerUUID, &trigger.ContractABI, trigger.ContractAdd

	// wrong func param type - for now we're just happy to log and assume the filter didn't match
	assert.Equal(t, validateFilter(tx, &trigger.Filters[0], cnt, abi, tid), false)
	assert.Equal(t, validateFilter(tx, &trigger.Filters[1], cnt, abi, tid), false)

	// checkFunctionCalled
	assert.Equal(t, validateFilter(tx, &trigger.Filters[2], cnt, abi, tid), true)

	log.SetLevel(log.DebugLevel)
}

func TestValidateFilter12(t *testing.T) {
	tx, _ := GetTransactionFromFile("../resources/transactions/tx2.json")
	trigger, _ := GetTriggerFromFile("../resources/triggers/t12.json")
	tid, abi, cnt := trigger.TriggerUUID, &trigger.ContractABI, trigger.ContractAdd

	// Index on bigInt[]
	assert.Equal(t, validateFilter(tx, &trigger.Filters[1], cnt, abi, tid), true)
	assert.Equal(t, validateFilter(tx, &trigger.Filters[2], cnt, abi, tid), false)

	// Index on address[]
	assert.Equal(t, validateFilter(tx, &trigger.Filters[3], cnt, abi, tid), true)
	assert.Equal(t, validateFilter(tx, &trigger.Filters[4], cnt, abi, tid), false)
	assert.Equal(t, validateFilter(tx, &trigger.Filters[5], cnt, abi, tid), false)

	// ConditionFunctionCalled
	assert.Equal(t, validateFilter(tx, &trigger.Filters[6], cnt, abi, tid), true)

	// address[]
	assert.Equal(t, validateFilter(tx, &trigger.Filters[7], cnt, abi, tid), true)
}

func TestValidateFilter13(t *testing.T) {
	// mute logging just for these tests to reduce noise
	log.SetLevel(log.WarnLevel)

	tx, _ := GetTransactionFromFile("../resources/transactions/tx2.json")
	trigger, _ := GetTriggerFromFile("../resources/triggers/t12.json")
	tid, abi, cnt := trigger.TriggerUUID, &trigger.ContractABI, trigger.ContractAdd

	// CheckFunctionParameter - different method name
	trigger.Filters[7].FunctionName = "xxx"
	assert.Equal(t, validateFilter(tx, &trigger.Filters[7], cnt, abi, tid), false)

	// ConditionFunctionCalled - wrong ABI
	trigger.ContractABI = "xxx"
	assert.Equal(t, validateFilter(tx, &trigger.Filters[6], cnt, abi, tid), false)

	log.SetLevel(log.DebugLevel)
}

func TestValidateFilter14(t *testing.T) {
	tx, _ := GetTransactionFromFile("../resources/transactions/tx7.json")
	trigger, _ := GetTriggerFromFile("../resources/triggers/t13.json")
	tid, abi, cnt := trigger.TriggerUUID, &trigger.ContractABI, trigger.ContractAdd

	// uint32
	assert.Equal(t, validateFilter(tx, &trigger.Filters[1], cnt, abi, tid), true)
}

func TestValidateFilter15(t *testing.T) {
	tx, _ := GetTransactionFromFile("../resources/transactions/tx8.json")
	trigger, _ := GetTriggerFromFile("../resources/triggers/t14.json")
	tid, abi, cnt := trigger.TriggerUUID, &trigger.ContractABI, trigger.ContractAdd

	// uint16
	assert.Equal(t, validateFilter(tx, &trigger.Filters[1], cnt, abi, tid), true)
}

// Testing one Trigger vs one Transaction
func TestValidateTrigger(t *testing.T) {
	block, _ := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := GetTriggerFromFile("../resources/triggers/t1.json")

	assert.Equal(t, validateTrigger(trigger, &block.Transactions[0]), true)
	assert.Equal(t, validateTrigger(trigger, &block.Transactions[1]), false)
}

func TestValidateTrigger2(t *testing.T) {
	block, _ := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := GetTriggerFromFile("../resources/triggers/t2.json")

	assert.Equal(t, validateTrigger(trigger, &block.Transactions[6]), true)
	assert.Equal(t, validateTrigger(trigger, &block.Transactions[1]), false)
	assert.Equal(t, validateTrigger(trigger, &block.Transactions[8]), true)
}

func TestValidateTriggerWithNoInputData(t *testing.T) {

	trigger, err := GetTriggerFromFile("../resources/triggers/t15.json")
	assert.NoError(t, err)

	block, err := config.CliMain.EthGetBlockByNumber(9466264, true)
	res := MatchTransaction(trigger, block)

	assert.Equal(t, 1, len(res))
	assert.Equal(t, "0xd11c6e75052c838c944608c478da95c09c2d239e417afbcb869c87469643ce57", res[0].Tx.Hash)
	assert.Nil(t, res[0].DecodedFnName)
	assert.Nil(t, res[0].DecodedFnArgs)
}

// Testing one Trigger vs one Block
func TestMatchTrigger(t *testing.T) {
	block, _ := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := GetTriggerFromFile("../resources/triggers/t2.json")

	matches := MatchTransaction(trigger, block)

	assert.Equal(t, len(matches), 2)
	assert.Equal(t, *(matches[0].Tx.TransactionIndex), 6)
	assert.Equal(t, *(matches[1].Tx.TransactionIndex), 8)

	assert.Equal(t, *(matches[0].DecodedFnName), "transfer")
	assert.Equal(t, *(matches[0].DecodedFnArgs), `{"_to":"0xfea2f9433058cd555fd67cdde8efd7e6031e56c0","_value":4000000000000000000}`)

	// testing ToPersistent()
	persistentJson, err := utils.GimmePrettyJson(matches[0].ToPersistent())
	expectedJsn := `{
  "DecodedData": {
    "FunctionArguments": "{\"_to\":\"0xfea2f9433058cd555fd67cdde8efd7e6031e56c0\",\"_value\":4000000000000000000}",
    "FunctionName": "transfer"
  },
  "Transaction": {
    "BlockHash": "0xb972fb8fe7a2aca471fa649e790ac51f59f920a2b71ec522aee606f1ccc99f6e",
    "BlockNumber": 7535077,
    "BlockTimestamp": 1554828248,
    "From": "0x3d2339bf362a9b0f8ef3ca0867bd73f350ed66ac",
    "Gas": 115960,
    "GasPrice": 7000000000,
    "Nonce": 414,
    "To": "0x174bfa6600bf90c885c7c01c7031389ed1461ab9",
    "Hash": "0x42c8de77ef5d76f36aea6e051b9059ece6e34619d9fb4a1d97f3224d5c990a67",
    "Value": 0,
    "InputData": "0xa9059cbb000000000000000000000000fea2f9433058cd555fd67cdde8efd7e6031e56c00000000000000000000000000000000000000000000000003782dace9d900000"
  }
}`
	ok, err := utils.AreEqualJSON(persistentJson, expectedJsn)
	assert.NoError(t, err)
	assert.True(t, ok)
}

// test that hex values are correctly decoded
func TestJsonToTransaction(t *testing.T) {
	tx, err := GetTransactionFromFile("../resources/transactions/tx1.json")

	assert.NoError(t, err)
	assert.Equal(t, *tx.BlockNumber, 7669714)
	assert.Equal(t, *tx.TransactionIndex, 4)
	assert.Equal(t, tx.Gas, 79068)
	assert.Equal(t, tx.Nonce, 233172)
}

func TestJsonToBlock(t *testing.T) {
	block, err := GetBlockFromFile("../resources/blocks/block1.json")

	assert.NoError(t, err)
	assert.Equal(t, block.Number, 7535077)
	assert.Equal(t, block.Size, 5392)
}
