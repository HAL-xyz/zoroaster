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

func TestValidateFilter(t *testing.T) {
	block := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t1.json")
	tid, abi := trigger.TriggerId, &trigger.ContractABI

	// BasicFilter / To
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[0], abi, tid), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[1], &trigger.Filters[0], abi, tid), false)

	// BasicFilter / Nonce
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[2], abi, tid), true)
}

func TestValidateFilter2(t *testing.T) {
	block := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t1.json")
	tid, abi := trigger.TriggerId, &trigger.ContractABI

	// Address
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[1], abi, tid), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[1], &trigger.Filters[1], abi, tid), false)
}

func TestValidateFilter3(t *testing.T) {
	block := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t3.json")
	tid, abi := trigger.TriggerId, &trigger.ContractABI

	// From
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[0], abi, tid), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[5], &trigger.Filters[0], abi, tid), false)
}

func TestValidateFilter4(t *testing.T) {
	block := GetBlockFromFile("../resources/blocks/block1.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t4.json")
	tid, abi := trigger.TriggerId, &trigger.ContractABI

	// Value
	assert.Equal(t, ValidateFilter(&block.Transactions[2], &trigger.Filters[0], abi, tid), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[0], abi, tid), false)

	// Gas
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[1], abi, tid), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[5], &trigger.Filters[1], abi, tid), false)

	// GasPrice
	assert.Equal(t, ValidateFilter(&block.Transactions[7], &trigger.Filters[2], abi, tid), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[4], &trigger.Filters[2], abi, tid), false)

	// Nonce
	assert.Equal(t, ValidateFilter(&block.Transactions[5], &trigger.Filters[3], abi, tid), true)
}

func TestValidateFilter5(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx1.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t5.json")
	tid, abi := trigger.TriggerId, &trigger.ContractABI

	// uint256[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], abi, tid), false)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], abi, tid), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], abi, tid), false)

	// bytes14[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[3], abi, tid), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[4], abi, tid), true)

	// Gas
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[5], abi, tid), true)

	// Nonce
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[6], abi, tid), true)
}

func TestValidateFilter6(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx2.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t6.json")
	tid, abi := trigger.TriggerId, &trigger.ContractABI

	// address[N]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], abi, tid), true)

	// uint256[N]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], abi, tid), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], abi, tid), true)
}

func TestValidateFilter7(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx3.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t7.json")
	tid, abi := trigger.TriggerId, &trigger.ContractABI

	// uint256
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], abi, tid), true)

	// bool
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], abi, tid), true)

	// int128
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], abi, tid), true)
}

func TestValidateFilter8(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx4.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t8.json")
	tid, abi := trigger.TriggerId, &trigger.ContractABI

	// int128[N]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], abi, tid), true)

	// int[N]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], abi, tid), true)

	// int40
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], abi, tid), true)
}

func TestValidateFilter9(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx5.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t9.json")
	tid, abi := trigger.TriggerId, &trigger.ContractABI

	// int32
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], abi, tid), true)

	// int32[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], abi, tid), true)

	// int32[6]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], abi, tid), true)

	// Index int32[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[7], abi, tid), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[8], abi, tid), false)
}

func TestValidateFilter10(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx6.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t9.json")
	tid, abi := trigger.TriggerId, &trigger.ContractABI

	// address[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[3], abi, tid), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[4], abi, tid), true)

	// bytes1[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[5], abi, tid), true)

	// string[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[6], abi, tid), true)
}

func TestValidateFilter11(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx5.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t10.json")
	tid, abi := trigger.TriggerId, &trigger.ContractABI

	// wrong func param type - for now we're just happy to log and assume the filter didn't match
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], abi, tid), false)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], abi, tid), false)
}

func TestValidateFilter12(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx2.json")
	trigger, _ := NewTriggerFromFile("../resources/triggers/t12.json")
	tid, abi := trigger.TriggerId, &trigger.ContractABI

	// Index on bigInt[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], abi, tid), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], abi, tid), false)

	// Index on address[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[3], abi, tid), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[4], abi, tid), false)
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
