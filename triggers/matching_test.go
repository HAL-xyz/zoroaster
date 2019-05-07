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

}

// Testing one Trigger vs one Transaction
func TestValidateTrigger(t *testing.T) {

	block := getBlockFromFile("../resources/blocks/block1.json")
	trigger := getTriggerFromFile("../resources/triggers/t1.json")

	trig, ok := ValidateTrigger(trigger, &block.Transactions[0])
	if trig.TriggerId != 101 || ok != true {
		t.Error()
	}

	trig2, ok2 := ValidateTrigger(trigger, &block.Transactions[1])
	if trig2 != nil || ok2 != false {
		t.Error()
	}
}

func TestValidateTrigger2(t *testing.T) {

	block := getBlockFromFile("../resources/blocks/block1.json")
	trigger := getTriggerFromFile("../resources/triggers/t2.json")

	trig, ok := ValidateTrigger(trigger, &block.Transactions[6])
	if trig.TriggerId != 102 || ok != true {
		t.Error()
	}

	trig2, ok2 := ValidateTrigger(trigger, &block.Transactions[1])
	if trig2 != nil || ok2 != false {
		t.Error()
	}

	trig3, ok3 := ValidateTrigger(trigger, &block.Transactions[8])
	if trig3.TriggerId != 102 || ok3 != true {
		t.Error()
	}
}

// Testing one Trigger vs one Block
func TestMatchTrigger(t *testing.T) {
	block := getBlockFromFile("../resources/blocks/block1.json")
	trigger := getTriggerFromFile("../resources/triggers/t2.json")

	matches := MatchTrigger(trigger, block)
	if len(matches) != 2 || matches[0].TriggerId != matches[1].TriggerId {
		t.Error()
	}

}
