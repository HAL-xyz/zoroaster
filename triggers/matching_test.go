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
	tn, abi, cnt := trigger.TriggerName, &trigger.ContractABI, trigger.ContractAdd

	// BasicFilter / To
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[0], cnt, abi, tn), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[1], &trigger.Filters[0], cnt, abi, tn), false)

	// BasicFilter / Nonce
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[2], cnt, abi, tn), true)
}

func TestValidateFilter2(t *testing.T) {
	block := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t1.json")
	tn, abi, cnt := trigger.TriggerName, &trigger.ContractABI, trigger.ContractAdd

	// Address
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[1], cnt, abi, tn), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[1], &trigger.Filters[1], cnt, abi, tn), false)
}

func TestValidateFilter3(t *testing.T) {
	block := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t3.json")
	tn, abi, cnt := trigger.TriggerName, &trigger.ContractABI, trigger.ContractAdd

	// From
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[0], cnt, abi, tn), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[5], &trigger.Filters[0], cnt, abi, tn), false)
}

func TestValidateFilter4(t *testing.T) {
	block := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t4.json")
	tn, abi, cnt := trigger.TriggerName, &trigger.ContractABI, trigger.ContractAdd

	// Value
	assert.Equal(t, ValidateFilter(&block.Transactions[2], &trigger.Filters[0], cnt, abi, tn), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[0], cnt, abi, tn), false)

	// Gas
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[1], cnt, abi, tn), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[5], &trigger.Filters[1], cnt, abi, tn), false)

	// GasPrice
	assert.Equal(t, ValidateFilter(&block.Transactions[7], &trigger.Filters[2], cnt, abi, tn), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[4], &trigger.Filters[2], cnt, abi, tn), false)

	// Nonce
	assert.Equal(t, ValidateFilter(&block.Transactions[5], &trigger.Filters[3], cnt, abi, tn), true)
}

func TestValidateFilter5(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx1.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t5.json")
	tn, abi, cnt := trigger.TriggerName, &trigger.ContractABI, trigger.ContractAdd

	// uint256[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], cnt, abi, tn), false)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], cnt, abi, tn), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], cnt, abi, tn), false)

	// bytes14[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[3], cnt, abi, tn), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[4], cnt, abi, tn), true)

	// Gas
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[5], cnt, abi, tn), true)

	// Nonce
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[6], cnt, abi, tn), true)
}

func TestValidateFilter6(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx2.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t6.json")
	tn, abi, cnt := trigger.TriggerName, &trigger.ContractABI, trigger.ContractAdd

	// address[N]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], cnt, abi, tn), true)

	// uint256[N]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], cnt, abi, tn), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], cnt, abi, tn), true)
}

func TestValidateFilter7(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx3.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t7.json")
	tn, abi, cnt := trigger.TriggerName, &trigger.ContractABI, trigger.ContractAdd

	// uint256
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], cnt, abi, tn), true)

	// bool
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], cnt, abi, tn), true)

	// int128
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], cnt, abi, tn), true)
}

func TestValidateFilter8(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx4.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t8.json")
	tn, abi, cnt := trigger.TriggerName, &trigger.ContractABI, trigger.ContractAdd

	// int128[N]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], cnt, abi, tn), true)

	// int[N]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], cnt, abi, tn), true)

	// int40
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], cnt, abi, tn), true)
}

func TestValidateFilter9(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx5.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t9.json")
	tn, abi, cnt := trigger.TriggerName, &trigger.ContractABI, trigger.ContractAdd

	// int32
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], cnt, abi, tn), true)

	// int32[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], cnt, abi, tn), true)

	// int32[6]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], cnt, abi, tn), true)

	// Index int32[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[7], cnt, abi, tn), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[8], cnt, abi, tn), false)
}

func TestValidateFilter10(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx6.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t9.json")
	tn, abi, cnt := trigger.TriggerName, &trigger.ContractABI, trigger.ContractAdd

	// address[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[3], cnt, abi, tn), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[4], cnt, abi, tn), true)

	// bytes1[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[5], cnt, abi, tn), true)

	// string[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[6], cnt, abi, tn), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[9], cnt, abi, tn), false)
}

func TestValidateFilter11(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx5.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t10.json")
	tn, abi, cnt := trigger.TriggerName, &trigger.ContractABI, trigger.ContractAdd

	// wrong func param type - for now we're just happy to log and assume the filter didn't match
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], cnt, abi, tn), false)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], cnt, abi, tn), false)

	// checkFunctionCalled
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], cnt, abi, tn), true)
}

func TestValidateFilter12(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx2.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t12.json")
	tn, abi, cnt := trigger.TriggerName, &trigger.ContractABI, trigger.ContractAdd

	// Index on bigInt[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], cnt, abi, tn), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], cnt, abi, tn), false)

	// Index on address[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[3], cnt, abi, tn), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[4], cnt, abi, tn), false)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[5], cnt, abi, tn), false)

	// ConditionFunctionCalled
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[6], cnt, abi, tn), true)
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
