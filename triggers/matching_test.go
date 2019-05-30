package trigger

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// no logs when running tests
//func TestMain(m *testing.M) {
//	log.SetOutput(ioutil.Discard)
//	m.Run()
//}

func TestValidateFilter1(t *testing.T) {
	block := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t1.json")
	tid, abi, cnt := trigger.TriggerId, &trigger.ContractABI, trigger.ContractAdd

	// BasicFilter / To
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[0], cnt, abi, tid), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[1], &trigger.Filters[0], cnt, abi, tid), false)

	// BasicFilter / Nonce
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[2], cnt, abi, tid), true)
}

func TestValidateFilter2(t *testing.T) {
	block := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t1.json")
	tid, abi, cnt := trigger.TriggerId, &trigger.ContractABI, trigger.ContractAdd

	// Address
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[1], cnt, abi, tid), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[1], &trigger.Filters[1], cnt, abi, tid), false)
}

func TestValidateFilter3(t *testing.T) {
	block := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t3.json")
	tid, abi, cnt := trigger.TriggerId, &trigger.ContractABI, trigger.ContractAdd

	// From
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[0], cnt, abi, tid), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[5], &trigger.Filters[0], cnt, abi, tid), false)
}

func TestValidateFilter4(t *testing.T) {
	block := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t4.json")
	tid, abi, cnt := trigger.TriggerId, &trigger.ContractABI, trigger.ContractAdd

	// Value
	assert.Equal(t, ValidateFilter(&block.Transactions[2], &trigger.Filters[0], cnt, abi, tid), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[0], cnt, abi, tid), false)

	// Gas
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[1], cnt, abi, tid), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[5], &trigger.Filters[1], cnt, abi, tid), false)

	// GasPrice
	assert.Equal(t, ValidateFilter(&block.Transactions[7], &trigger.Filters[2], cnt, abi, tid), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[4], &trigger.Filters[2], cnt, abi, tid), false)

	// Nonce
	assert.Equal(t, ValidateFilter(&block.Transactions[5], &trigger.Filters[3], cnt, abi, tid), true)
}

func TestValidateFilter5(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx1.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t5.json")
	tid, abi, cnt := trigger.TriggerId, &trigger.ContractABI, trigger.ContractAdd

	// uint256[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], cnt, abi, tid), false)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], cnt, abi, tid), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], cnt, abi, tid), false)

	// bytes14[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[3], cnt, abi, tid), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[4], cnt, abi, tid), true)

	// Gas
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[5], cnt, abi, tid), true)

	// Nonce
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[6], cnt, abi, tid), true)
}

func TestValidateFilter6(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx2.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t6.json")
	tid, abi, cnt := trigger.TriggerId, &trigger.ContractABI, trigger.ContractAdd

	// address[N]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], cnt, abi, tid), true)

	// uint256[N]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], cnt, abi, tid), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], cnt, abi, tid), true)
}

func TestValidateFilter7(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx3.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t7.json")
	tid, abi, cnt := trigger.TriggerId, &trigger.ContractABI, trigger.ContractAdd

	// uint256
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], cnt, abi, tid), true)

	// bool
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], cnt, abi, tid), true)

	// int128
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], cnt, abi, tid), true)
}

func TestValidateFilter8(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx4.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t8.json")
	tid, abi, cnt := trigger.TriggerId, &trigger.ContractABI, trigger.ContractAdd

	// int128[N]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], cnt, abi, tid), true)

	// int[N]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], cnt, abi, tid), true)

	// int40
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], cnt, abi, tid), true)
}

func TestValidateFilter9(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx5.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t9.json")
	tid, abi, cnt := trigger.TriggerId, &trigger.ContractABI, trigger.ContractAdd

	// int32
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], cnt, abi, tid), true)

	// int32[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], cnt, abi, tid), true)

	// int32[6]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], cnt, abi, tid), true)

	// Index int32[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[7], cnt, abi, tid), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[8], cnt, abi, tid), false)
}

func TestValidateFilter10(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx6.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t9.json")
	tid, abi, cnt := trigger.TriggerId, &trigger.ContractABI, trigger.ContractAdd

	// address[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[3], cnt, abi, tid), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[4], cnt, abi, tid), true)

	// bytes1[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[5], cnt, abi, tid), true)

	// string[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[6], cnt, abi, tid), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[9], cnt, abi, tid), false)
}

func TestValidateFilter11(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx5.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t10.json")
	tid, abi, cnt := trigger.TriggerId, &trigger.ContractABI, trigger.ContractAdd

	// wrong func param type - for now we're just happy to log and assume the filter didn't match
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], cnt, abi, tid), false)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], cnt, abi, tid), false)

	// checkFunctionCalled
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], cnt, abi, tid), true)
}

func TestValidateFilter12(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx2.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t12.json")
	tid, abi, cnt := trigger.TriggerId, &trigger.ContractABI, trigger.ContractAdd

	// Index on bigInt[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], cnt, abi, tid), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], cnt, abi, tid), false)

	// Index on address[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[3], cnt, abi, tid), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[4], cnt, abi, tid), false)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[5], cnt, abi, tid), false)

	// ConditionFunctionCalled
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[6], cnt, abi, tid), true)
}

// Testing one Trigger vs one Transaction
func TestValidateTrigger(t *testing.T) {
	block := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t1.json")

	assert.Equal(t, ValidateTrigger(trigger, &block.Transactions[0]), true)
	assert.Equal(t, ValidateTrigger(trigger, &block.Transactions[1]), false)
}

func TestValidateTrigger2(t *testing.T) {
	block := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t2.json")

	assert.Equal(t, ValidateTrigger(trigger, &block.Transactions[6]), true)
	assert.Equal(t, ValidateTrigger(trigger, &block.Transactions[1]), false)
	assert.Equal(t, ValidateTrigger(trigger, &block.Transactions[8]), true)
}

// Testing one Trigger vs one Block
func TestMatchTrigger(t *testing.T) {
	block := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t2.json")

	txs := MatchTrigger(trigger, block)

	assert.Equal(t, len(txs), 2)
	assert.Equal(t, *txs[0].TransactionIndex, 6)
	assert.Equal(t, *txs[1].TransactionIndex, 8)
}
