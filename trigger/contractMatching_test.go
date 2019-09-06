package trigger

import (
	"fmt"
	"github.com/onrik/ethrpc"
	"github.com/stretchr/testify/assert"
	"testing"
	"zoroaster/config"
)

var zconf = config.Load("../config")
var client = ethrpc.New(zconf.EthNode)
var cliRinkeby = ethrpc.New("https://rinkebyshared.bdnodes.net?auth=dKvc9d7tXrOdmnKK9nsfl119I19PH4GZPbACnbH-QW0")

func TestMatchContract1(t *testing.T) {

	// () -> address
	tg, err := NewTriggerFromFile("../resources/triggers/wac1.json")
	if err != nil {
		t.Error(t)
	}
	isMatch, matchingVals, returnedVals := MatchContract(client, tg, 8387102)
	assert.True(t, isMatch)
	assert.Equal(t, matchingVals[0], "0x4a574510c7014e4ae985403536074abe582adfc8")
	assert.Equal(t, fmt.Sprint(returnedVals), "[\"0x4a574510c7014e4ae985403536074abe582adfc8\"]")
}

func TestMatchContract2(t *testing.T) {

	// address -> uint256
	tg, err := NewTriggerFromFile("../resources/triggers/wac2.json")
	if err != nil {
		t.Error(t)
	}
	isMatch, matchingVals, returnedVals := MatchContract(client, tg, 8387679)
	assert.True(t, isMatch)
	assert.Equal(t, matchingVals[0], "3876846319093283908984")
	assert.Equal(t, fmt.Sprint(returnedVals), "[3876846319093283908984]")
}

func TestMatchContract3(t *testing.T) {

	// () -> bool
	tg, err := NewTriggerFromFile("../resources/triggers/wac3.json")
	if err != nil {
		t.Error(t)
	}
	isMatch, matchingVals, returnedVals := MatchContract(client, tg, 8387102)
	assert.True(t, isMatch)
	assert.Equal(t, matchingVals[0], "true")
	assert.Equal(t, fmt.Sprint(returnedVals), "[true]")
}

func TestMatchContract4(t *testing.T) {

	// uint256 -> address
	tg, err := NewTriggerFromFile("../resources/triggers/wac4.json")
	if err != nil {
		t.Error(t)
	}
	isMatch, matchingVals, returnedVals := MatchContract(client, tg, 8387102)
	assert.True(t, isMatch)
	assert.Equal(t, matchingVals[0], "0xd4fe7bc31cedb7bfb8a345f31e668033056b2728")
	assert.Equal(t, fmt.Sprint(returnedVals), "[\"0xd4fe7bc31cedb7bfb8a345f31e668033056b2728\"]")
}

func TestMatchContract5(t *testing.T) {

	// uint16 -> address
	tg, err := NewTriggerFromFile("../resources/triggers/wac5.json")
	if err != nil {
		t.Error(t)
	}
	isMatch, matchingVals, returnedVals := MatchContract(client, tg, 8387102)
	assert.True(t, isMatch)
	assert.Equal(t, matchingVals[0], "0x02ca0dfabf5285b0b9d09dfaa241167013355c35")
	assert.Equal(t, fmt.Sprint(returnedVals), "[\"0x02ca0dfabf5285b0b9d09dfaa241167013355c35\"]")
}

func TestMatchContract6(t *testing.T) {

	// () -> uint256[3]
	tg, err := NewTriggerFromFile("../resources/triggers/wac6.json")
	if err != nil {
		t.Error(t)
	}

	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)
	assert.True(t, isMatch)
	assert.Equal(t, matchingVals[0], "12")
	assert.Equal(t, fmt.Sprint(returnedVals), "[[4,8,12]]")
}

func TestMatchContract7(t *testing.T) {

	// () -> (int128, int128, int128)
	tg, err := NewTriggerFromFile("../resources/triggers/wac7.json")
	if err != nil {
		t.Error(t)
	}
	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)
	assert.True(t, isMatch)
	assert.Equal(t, matchingVals[0], "4")
	assert.Equal(t, fmt.Sprint(returnedVals), "[4#END# 8#END# 12]")
}

func TestMatchContract8(t *testing.T) {

	// () -> (int128, string, string)
	tg, err := NewTriggerFromFile("../resources/triggers/wac8.json")
	if err != nil {
		t.Error(t)
	}
	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)
	assert.True(t, isMatch)
	assert.Equal(t, matchingVals[0], "moon")
	assert.Equal(t, fmt.Sprint(returnedVals), "[4#END# \"sailor\"#END# \"moon\"]")
}

func TestMatchContract9(t *testing.T) {

	// () -> string[3]
	tg, err := NewTriggerFromFile("../resources/triggers/wac9.json")
	if err != nil {
		t.Error(t)
	}
	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)
	assert.True(t, isMatch)
	assert.Equal(t, matchingVals[0], "ciao")
	assert.Equal(t, fmt.Sprint(returnedVals), "[[\"ciao\",\"come\",\"stai\"]]")
}

func TestMatchContract10(t *testing.T) {

	// () -> string[3]
	tg, err := NewTriggerFromFile("../resources/triggers/wac10.json")
	if err != nil {
		t.Error(t)
	}
	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)

	assert.True(t, isMatch)
	assert.Equal(t, len(matchingVals), 2)
	assert.Equal(t, matchingVals[0], "10")
	assert.Equal(t, matchingVals[1], "0x4fed1fc4144c223ae3c1553be203cdfcbd38c581")
	assert.Equal(t, returnedVals, `[["0x4fed1fc4144c223ae3c1553be203cdfcbd38c581","0x65d21616594825a738bcd08a5227358593a9aaf2","0xd76f7d7d2ede0631ad23e4a01176c0e59878abda"]]`)
}

func TestMatchContract11(t *testing.T) {

	// () -> string[3]
	tg, err := NewTriggerFromFile("../resources/triggers/wac11.json")
	if err != nil {
		t.Error(t)
	}
	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)

	assert.False(t, isMatch)
	assert.Equal(t, matchingVals[0], "10")
	assert.Equal(t, matchingVals[1], "")
	assert.Equal(t, returnedVals, `[["0x4fed1fc4144c223ae3c1553be203cdfcbd38c581","0x65d21616594825a738bcd08a5227358593a9aaf2","0xd76f7d7d2ede0631ad23e4a01176c0e59878abda"]]`)
}

func TestMatchContractUniswap(t *testing.T) {

	tg, err := NewTriggerFromFile("../resources/triggers/wac-uniswap.json")
	if err != nil {
		t.Error(t)
	}

	// TODO: why is this returning 0x?
	_, _, _ = MatchContract(client, tg, 8496486)
}
