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

func TestMatchContract6(t *testing.T) {

	// () -> uint256[3]
	tg, err := NewTriggerFromFile("../resources/triggers/wac6.json")
	if err != nil {
		t.Error(t)
	}

	cli := ethrpc.New("https://rinkebyshared.bdnodes.net?auth=dKvc9d7tXrOdmnKK9nsfl119I19PH4GZPbACnbH-QW0")

	value := MatchContract(cli, tg, 4974958)
	assert.Equal(t, value, "12")
}

func TestMatchContract7(t *testing.T) {

	// () -> (int128, int128, int128)
	tg, err := NewTriggerFromFile("../resources/triggers/wac7.json")
	if err != nil {
		t.Error(t)
	}

	cli := ethrpc.New("https://rinkebyshared.bdnodes.net?auth=dKvc9d7tXrOdmnKK9nsfl119I19PH4GZPbACnbH-QW0")

	value := MatchContract(cli, tg, 4974958)
	assert.Equal(t, value, "4")
}

func TestValidateContractReturnValue(t *testing.T) {

	// test the decoding of different types returned when invoking a contract

	// Address
	res := validateContractReturnValue(
		"Address",
		"0x000000000000000000000000f06e8ac2d2d449f5cf3605d8b33f736a28d512c4",
		ConditionOutput{Condition{}, Eq, "0x000000000000000000000000f06e8ac2d2d449f5cf3605d8b33f736a28d512c4"}, nil)
	assert.Equal(t, res, "0xf06e8ac2d2d449f5cf3605d8b33f736a28d512c4")

	// string
	res2 := validateContractReturnValue(
		"string",
		"0x0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000642617a6161720000000000000000000000000000000000000000000000000000",
		ConditionOutput{Condition{}, Eq, "Bazaar"}, nil)
	assert.Equal(t, res2, "Bazaar")

	res3 := validateContractReturnValue(
		"string",
		"0x0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000c5265736561726368204c61620000000000000000000000000000000000000000",
		ConditionOutput{Condition{}, Eq, "Research Lab"}, nil)
	assert.Equal(t, res3, "Research Lab")

	// uint32
	res4 := validateContractReturnValue(
		"uint32",
		"0x0000000000000000000000000000000000000000000000000000000000007530",
		ConditionOutput{Condition{}, Eq, "30000"}, nil)
	assert.Equal(t, res4, "30000")

	// uint32[3]
	index := 1
	res5 := validateContractReturnValue(
		"uint32[3]",
		"0x000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000780000000000000000000000000000000000000000000000000000000000000000",
		ConditionOutput{Condition{}, Eq, "120"}, &index)

	assert.Equal(t, res5, "120")

	// address[3]
	index = 0
	res6 := validateContractReturnValue(
		"address[3]",
		"0x0000000000000000000000004fed1fc4144c223ae3c1553be203cdfcbd38c58100000000000000000000000065d21616594825a738bcd08a5227358593a9aaf2000000000000000000000000d76f7d7d2ede0631ad23e4a01176c0e59878abda",
		ConditionOutput{Condition{}, Eq, "0x4FED1fC4144c223aE3C1553be203cDFcbD38C581"}, &index)

	assert.Equal(t, res6, "0x4fed1fc4144c223ae3c1553be203cdfcbd38c581")

	// (int128, int128, int128)
	index = 0
	res7 := validateContractReturnValue(
		"(int128, int128, int128)",
		"0x00000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c",
		ConditionOutput{Condition{}, Eq, "4"}, &index)

	assert.Equal(t, res7, "4")
}
