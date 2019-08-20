package trigger

import (
	"github.com/onrik/ethrpc"
	"github.com/stretchr/testify/assert"
	"testing"
	"zoroaster/config"
)

var zconf = config.Load("../config")
var client = ethrpc.New(zconf.EthNode)

func TestMatchContract(t *testing.T) {

	// () -> Address
	tg, err := NewTriggerFromFile("../resources/triggers/wac1.json")
	if err != nil {
		t.Error(t)
	}
	value := MatchContract(client, tg, 8387102)
	assert.Equal(t, value, "0x4a574510c7014e4ae985403536074abe582adfc8")
}

func TestMatchContract2(t *testing.T) {

	// Address -> uint256
	tg, err := NewTriggerFromFile("../resources/triggers/wac2.json")
	if err != nil {
		t.Error(t)
	}
	value := MatchContract(client, tg, 8387679)
	assert.Equal(t, value, "3876846319093283908984")
}

func TestMatchContract3(t *testing.T) {

	// () -> bool
	tg, err := NewTriggerFromFile("../resources/triggers/wac3.json")
	if err != nil {
		t.Error(t)
	}
	value := MatchContract(client, tg, 8387102)
	assert.Equal(t, value, "true")
}

func TestMatchContract4(t *testing.T) {

	// uint256 -> Address
	tg, err := NewTriggerFromFile("../resources/triggers/wac4.json")
	if err != nil {
		t.Error(t)
	}
	value := MatchContract(client, tg, 8387102)
	assert.Equal(t, value, "0xd4fe7bc31cedb7bfb8a345f31e668033056b2728")
}

func TestMatchContract5(t *testing.T) {

	// uint16 -> address
	tg, err := NewTriggerFromFile("../resources/triggers/wac5.json")
	if err != nil {
		t.Error(t)
	}
	value := MatchContract(client, tg, 8387102)
	assert.Equal(t, value, "0x02ca0dfabf5285b0b9d09dfaa241167013355c35")
}

func TestValidateContractReturnValue(t *testing.T) {

	// test the decoding of different types returned when invoking a contract

	// Address
	res := validateContractReturnValue(
		"Address",
		"0x000000000000000000000000f06e8ac2d2d449f5cf3605d8b33f736a28d512c4",
		ConditionOutput{Condition{}, Eq, "0x000000000000000000000000f06e8ac2d2d449f5cf3605d8b33f736a28d512c4"})
	assert.Equal(t, res, "0xf06e8ac2d2d449f5cf3605d8b33f736a28d512c4")

	// string
	res2 := validateContractReturnValue(
		"string",
		"0x0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000642617a6161720000000000000000000000000000000000000000000000000000",
		ConditionOutput{Condition{}, Eq, "Bazaar"})
	assert.Equal(t, res2, "Bazaar")

	res3 := validateContractReturnValue(
		"string",
		"0x0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000c5265736561726368204c61620000000000000000000000000000000000000000",
		ConditionOutput{Condition{}, Eq, "Research Lab"})
	assert.Equal(t, res3, "Research Lab")

	// uint32
	res4 := validateContractReturnValue(
		"uint32",
		"0x0000000000000000000000000000000000000000000000000000000000007530",
		ConditionOutput{Condition{}, Eq, "30000"})
	assert.Equal(t, res4, "30000")

}
