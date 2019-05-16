package trigger

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateFilter(t *testing.T) {
	block := getBlockFromFile("../resources/blocks/block1.json")
	trigger := getTriggerFromFile("../resources/triggers/t1.json")

	// BasicFilter / To
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[0], &trigger.ContractABI), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[1], &trigger.Filters[0], &trigger.ContractABI), false)

	// BasicFilter / Nonce
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[2], &trigger.ContractABI), true)
}

func TestValidateFilter2(t *testing.T) {
	block := getBlockFromFile("../resources/blocks/block1.json")
	trigger := getTriggerFromFile("../resources/triggers/t1.json")

	// Address
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[1], &trigger.ContractABI), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[1], &trigger.Filters[1], &trigger.ContractABI), false)
}

func TestValidateFilter3(t *testing.T) {
	block := getBlockFromFile("../resources/blocks/block1.json")
	trigger := getTriggerFromFile("../resources/triggers/t3.json")

	// From
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[0], &trigger.ContractABI), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[5], &trigger.Filters[0], &trigger.ContractABI), false)
}

func TestValidateFilter4(t *testing.T) {
	block := getBlockFromFile("../resources/blocks/block1.json")
	trigger := getTriggerFromFile("../resources/triggers/t4.json")

	// Value
	assert.Equal(t, ValidateFilter(&block.Transactions[2], &trigger.Filters[0], &trigger.ContractABI), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[0], &trigger.ContractABI), false)

	// Gas
	assert.Equal(t, ValidateFilter(&block.Transactions[0], &trigger.Filters[1], &trigger.ContractABI), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[5], &trigger.Filters[1], &trigger.ContractABI), false)

	// GasPrice
	assert.Equal(t, ValidateFilter(&block.Transactions[7], &trigger.Filters[2], &trigger.ContractABI), true)
	assert.Equal(t, ValidateFilter(&block.Transactions[4], &trigger.Filters[2], &trigger.ContractABI), false)
}

func TestValidateFilter5(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx1.json")
	trigger := getTriggerFromFile("../resources/triggers/t5.json")

	// uint256[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], &trigger.ContractABI), false)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], &trigger.ContractABI), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], &trigger.ContractABI), false)

	// bytes14[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[3], &trigger.ContractABI), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[4], &trigger.ContractABI), true)
}

func TestValidateFilter6(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx2.json")
	trigger := getTriggerFromFile("../resources/triggers/t6.json")

	// address[N]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], &trigger.ContractABI), true)

	// uint256[N]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], &trigger.ContractABI), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], &trigger.ContractABI), true)
}

func TestValidateFilter7(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx3.json")
	trigger := getTriggerFromFile("../resources/triggers/t7.json")

	// uint256
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], &trigger.ContractABI), true)

	// bool
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], &trigger.ContractABI), true)

	// int128
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], &trigger.ContractABI), true)
}

func TestValidateFilter8(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx4.json")
	trigger := getTriggerFromFile("../resources/triggers/t8.json")

	// int128[N]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], &trigger.ContractABI), true)

	// int[N]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], &trigger.ContractABI), true)

	// int40
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], &trigger.ContractABI), true)
}

func TestValidateFilter9(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx5.json")
	trigger := getTriggerFromFile("../resources/triggers/t9.json")

	// int32
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[0], &trigger.ContractABI), true)

	// int32[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[1], &trigger.ContractABI), true)

	// int32[6]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[2], &trigger.ContractABI), true)
}

func TestValidateFilter10(t *testing.T) {
	tx := getTransactionFromFile("../resources/transactions/tx6.json")
	trigger := getTriggerFromFile("../resources/triggers/t9.json")

	// address[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[3], &trigger.ContractABI), true)
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[4], &trigger.ContractABI), true)

	// bytes1[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[5], &trigger.ContractABI), true)

	// string[]
	assert.Equal(t, ValidateFilter(tx, &trigger.Filters[6], &trigger.ContractABI), true)
}

// Testing one Trigger vs one Transaction
func TestValidateTrigger(t *testing.T) {
	block := getBlockFromFile("../resources/blocks/block1.json")
	trigger := getTriggerFromFile("../resources/triggers/t1.json")

	assert.Equal(t, ValidateTrigger(trigger, &block.Transactions[0]), true)
	assert.Equal(t, ValidateTrigger(trigger, &block.Transactions[1]), false)
}

func TestValidateTrigger2(t *testing.T) {
	block := getBlockFromFile("../resources/blocks/block1.json")
	trigger := getTriggerFromFile("../resources/triggers/t2.json")

	assert.Equal(t, ValidateTrigger(trigger, &block.Transactions[6]), true)
	assert.Equal(t, ValidateTrigger(trigger, &block.Transactions[1]), false)
	assert.Equal(t, ValidateTrigger(trigger, &block.Transactions[8]), true)
}

// Testing one Trigger vs one Block
func TestMatchTrigger(t *testing.T) {
	block := getBlockFromFile("../resources/blocks/block1.json")
	trigger := getTriggerFromFile("../resources/triggers/t2.json")

	assert.Equal(t, MatchTrigger(trigger, block), 2)
}
