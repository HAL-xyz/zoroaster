package trigger

import (
	"testing"
)

// Testing one Filter VS one Transaction, Basic Filters
func TestValidateFilter(t *testing.T) {

	block := getBlockFromFile("../resources/blocks/block1.json")
	trigger := getTriggerFromFile("../resources/triggers/t1.json")

	// BasicFilter / To
	if ValidateFilter(&block.Transactions[0], &trigger.Filters[0], &trigger.ContractABI) != true {
		t.Error("Basic Filter / To should match")
	}
	if ValidateFilter(&block.Transactions[1], &trigger.Filters[0], &trigger.ContractABI) != false {
		t.Error("Basic Filter / To should NOT match")
	}

	// BasicFilter / Nonce
	if ValidateFilter(&block.Transactions[0], &trigger.Filters[2], &trigger.ContractABI) != true {
		t.Error("Basic Filter / Nonce should match")
	}
	if ValidateFilter(&block.Transactions[4], &trigger.Filters[2], &trigger.ContractABI) != false {
		t.Error("Basic Filter / Nonce should match")
	}
}

// Testing one Filter VS one Transaction, Function Params
func TestValidateFilter2(t *testing.T) {

	block := getBlockFromFile("../resources/blocks/block1.json")
	trigger := getTriggerFromFile("../resources/triggers/t1.json")

	// FunctionParameter / Address

	if ValidateFilter(&block.Transactions[0], &trigger.Filters[1], &trigger.ContractABI) != true {
		t.Error("FuncParam should match")
	}

	if ValidateFilter(&block.Transactions[1], &trigger.Filters[1], &trigger.ContractABI) != false {
		t.Error("FuncParam should NOT match")
	}
}

// BasicFilter / From
func TestValidateFilter3(t *testing.T) {

	block := getBlockFromFile("../resources/blocks/block1.json")
	trigger := getTriggerFromFile("../resources/triggers/t3.json")

	if ValidateFilter(&block.Transactions[0], &trigger.Filters[0], &trigger.ContractABI) != true {
		t.Error()
	}

	if ValidateFilter(&block.Transactions[5], &trigger.Filters[0], &trigger.ContractABI) != false {
		t.Error()
	}
}

// BasicFilter / Value, Gas, GasPrice
func TestValidateFilter4(t *testing.T) {

	block := getBlockFromFile("../resources/blocks/block1.json")
	trigger := getTriggerFromFile("../resources/triggers/t4.json")

	// Value
	if ValidateFilter(&block.Transactions[2], &trigger.Filters[0], &trigger.ContractABI) != true {
		t.Error()
	}

	if ValidateFilter(&block.Transactions[0], &trigger.Filters[0], &trigger.ContractABI) != false {
		t.Error()
	}

	// Gas
	if ValidateFilter(&block.Transactions[0], &trigger.Filters[1], &trigger.ContractABI) != true {
		t.Error()
	}

	if ValidateFilter(&block.Transactions[5], &trigger.Filters[1], &trigger.ContractABI) != false {
		t.Error()
	}

	// GasPrice
	if ValidateFilter(&block.Transactions[7], &trigger.Filters[2], &trigger.ContractABI) != true {
		t.Error()
	}

	if ValidateFilter(&block.Transactions[4], &trigger.Filters[2], &trigger.ContractABI) != false {
		t.Error()
	}

}

// Function Params, uint256[], bytes1...32[]
func TestValidateFilter5(t *testing.T) {

	tx := getTransactionFromFile("../resources/transactions/tx1.json")
	trigger := getTriggerFromFile("../resources/triggers/t5.json")

	// uint256[]
	if ValidateFilter(tx, &trigger.Filters[0], &trigger.ContractABI) != false {
		t.Error()
	}

	if ValidateFilter(tx, &trigger.Filters[1], &trigger.ContractABI) != true {
		t.Error()
	}

	if ValidateFilter(tx, &trigger.Filters[2], &trigger.ContractABI) != false {
		t.Error()
	}

	// bytes14[]
	if ValidateFilter(tx, &trigger.Filters[3], &trigger.ContractABI) != true {
		t.Error()
	}

	if ValidateFilter(tx, &trigger.Filters[4], &trigger.ContractABI) != true {
		t.Error()
	}
}

func TestValidateFilter6(t *testing.T) {

	tx := getTransactionFromFile("../resources/transactions/tx2.json")
	trigger := getTriggerFromFile("../resources/triggers/t6.json")

	// address[N]
	if ValidateFilter(tx, &trigger.Filters[0], &trigger.ContractABI) != true {
		t.Error()
	}

	// uint256[N]
	if ValidateFilter(tx, &trigger.Filters[1], &trigger.ContractABI) != true {
		t.Error()
	}

	if ValidateFilter(tx, &trigger.Filters[2], &trigger.ContractABI) != true {
		t.Error()
	}
}

func TestValidateFilter7(t *testing.T) {

	tx := getTransactionFromFile("../resources/transactions/tx3.json")
	trigger := getTriggerFromFile("../resources/triggers/t7.json")

	// uint256
	if ValidateFilter(tx, &trigger.Filters[0], &trigger.ContractABI) != true {
		t.Error()
	}

	// bool
	if ValidateFilter(tx, &trigger.Filters[1], &trigger.ContractABI) != true {
		t.Error()
	}

	// int128
	if ValidateFilter(tx, &trigger.Filters[2], &trigger.ContractABI) != true {
		t.Error()
	}
}

func TestValidateFilter8(t *testing.T) {

	tx := getTransactionFromFile("../resources/transactions/tx4.json")
	trigger := getTriggerFromFile("../resources/triggers/t8.json")

	// int128[N]
	if ValidateFilter(tx, &trigger.Filters[0], &trigger.ContractABI) != true {
		t.Error()
	}

	// int[N]
	if ValidateFilter(tx, &trigger.Filters[1], &trigger.ContractABI) != true {
		t.Error()
	}

	// int40
	if ValidateFilter(tx, &trigger.Filters[2], &trigger.ContractABI) != true {
		t.Error()
	}
}

func TestValidateFilter9(t *testing.T) {

	tx := getTransactionFromFile("../resources/transactions/tx5.json")
	trigger := getTriggerFromFile("../resources/triggers/t9.json")

	// int32
	if ValidateFilter(tx, &trigger.Filters[0], &trigger.ContractABI) != true {
		t.Error()
	}

	// int32[]
	if ValidateFilter(tx, &trigger.Filters[1], &trigger.ContractABI) != true {
		t.Error()
	}

	// int32[6]
	if ValidateFilter(tx, &trigger.Filters[2], &trigger.ContractABI) != true {
		t.Error()
	}
}

func TestValidateFilter10(t *testing.T) {

	tx := getTransactionFromFile("../resources/transactions/tx6.json")
	trigger := getTriggerFromFile("../resources/triggers/t9.json")

	// address[]
	if ValidateFilter(tx, &trigger.Filters[3], &trigger.ContractABI) != true {
		t.Error()
	}

	if ValidateFilter(tx, &trigger.Filters[4], &trigger.ContractABI) != true {
		t.Error()
	}

	// bytes1[]
	if ValidateFilter(tx, &trigger.Filters[5], &trigger.ContractABI) != true {
		t.Error()
	}

	// string[]
	if ValidateFilter(tx, &trigger.Filters[6], &trigger.ContractABI) != true {
		t.Error()
	}
}

// Testing one Trigger vs one Transaction
func TestValidateTrigger(t *testing.T) {

	block := getBlockFromFile("../resources/blocks/block1.json")
	trigger := getTriggerFromFile("../resources/triggers/t1.json")

	ok := ValidateTrigger(trigger, &block.Transactions[0])
	if ok != true {
		t.Error()
	}

	ok2 := ValidateTrigger(trigger, &block.Transactions[1])
	if ok2 != false {
		t.Error()
	}
}

func TestValidateTrigger2(t *testing.T) {

	block := getBlockFromFile("../resources/blocks/block1.json")
	trigger := getTriggerFromFile("../resources/triggers/t2.json")

	ok := ValidateTrigger(trigger, &block.Transactions[6])
	if ok != true {
		t.Error()
	}

	ok2 := ValidateTrigger(trigger, &block.Transactions[1])
	if ok2 != false {
		t.Error()
	}

	ok3 := ValidateTrigger(trigger, &block.Transactions[8])
	if ok3 != true {
		t.Error()
	}
}

// Testing one Trigger vs one Block
func TestMatchTrigger(t *testing.T) {
	block := getBlockFromFile("../resources/blocks/block1.json")
	trigger := getTriggerFromFile("../resources/triggers/t2.json")

	matches := MatchTrigger(trigger, block)
	if matches != 2 {
		t.Error()
	}
}
