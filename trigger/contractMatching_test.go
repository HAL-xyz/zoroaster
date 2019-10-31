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
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(client, tg, 8387102)
	assert.True(t, isMatch)
	assert.Equal(t, "0x4a574510c7014e4ae985403536074abe582adfc8", matchingVals[0])
	assert.Equal(t, "[\"0x4a574510c7014e4ae985403536074abe582adfc8\"]", fmt.Sprint(returnedVals))
}

func TestMatchContract2(t *testing.T) {

	// address -> uint256
	tg, err := NewTriggerFromFile("../resources/triggers/wac2.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(client, tg, 8387679)
	assert.True(t, isMatch)
	assert.Equal(t, "3876846319093283908984", matchingVals[0])
	assert.Equal(t, "[3876846319093283908984]", fmt.Sprint(returnedVals))
}

func TestMatchContract3(t *testing.T) {

	// () -> bool
	tg, err := NewTriggerFromFile("../resources/triggers/wac3.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(client, tg, 8387102)
	assert.True(t, isMatch)
	assert.Equal(t, "true", matchingVals[0])
	assert.Equal(t, "[true]", fmt.Sprint(returnedVals))
}

func TestMatchContract4(t *testing.T) {

	// uint256 -> address
	tg, err := NewTriggerFromFile("../resources/triggers/wac4.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(client, tg, 8387102)
	assert.True(t, isMatch)
	assert.Equal(t, "0xd4fe7bc31cedb7bfb8a345f31e668033056b2728", matchingVals[0])
	assert.Equal(t, "[\"0xd4fe7bc31cedb7bfb8a345f31e668033056b2728\"]", fmt.Sprint(returnedVals))
}

func TestMatchContract5(t *testing.T) {

	// uint16 -> address
	tg, err := NewTriggerFromFile("../resources/triggers/wac5.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(client, tg, 8387102)
	assert.True(t, isMatch)
	assert.Equal(t, "0x02ca0dfabf5285b0b9d09dfaa241167013355c35", matchingVals[0])
	assert.Equal(t, "[\"0x02ca0dfabf5285b0b9d09dfaa241167013355c35\"]", fmt.Sprint(returnedVals))
}

func TestMatchContract6(t *testing.T) {

	// () -> uint256[3]
	tg, err := NewTriggerFromFile("../resources/triggers/wac6.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)
	assert.True(t, isMatch)
	assert.Equal(t, "12", matchingVals[0])
	assert.Equal(t, "[[4,8,12]]", fmt.Sprint(returnedVals))
}

func TestMatchContract7(t *testing.T) {

	// () -> (int128, int128, int128)
	tg, err := NewTriggerFromFile("../resources/triggers/wac7.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)
	assert.True(t, isMatch)
	assert.Equal(t, "4", matchingVals[0])
	assert.Equal(t, "[4#END# 8#END# 12]", fmt.Sprint(returnedVals))
}

func TestMatchContract8(t *testing.T) {

	// () -> (int128, string, string)
	tg, err := NewTriggerFromFile("../resources/triggers/wac8.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)
	assert.True(t, isMatch)
	assert.Equal(t, "moon", matchingVals[0])
	assert.Equal(t, "[4#END# \"sailor\"#END# \"moon\"]", fmt.Sprint(returnedVals))
}

func TestMatchContract9(t *testing.T) {

	// () -> string[3]
	tg, err := NewTriggerFromFile("../resources/triggers/wac9.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)
	assert.True(t, isMatch)
	assert.Equal(t, "ciao", matchingVals[0])
	assert.Equal(t, "[[\"ciao\",\"come\",\"stai\"]]", fmt.Sprint(returnedVals))
}

func TestMatchContract10(t *testing.T) {

	// () -> string[3]
	tg, err := NewTriggerFromFile("../resources/triggers/wac10.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)
	assert.True(t, isMatch)
	assert.Equal(t, len(matchingVals), 2)
	assert.Equal(t, "10", matchingVals[0])
	assert.Equal(t, "0x4fed1fc4144c223ae3c1553be203cdfcbd38c581", matchingVals[1])
	assert.Equal(t, `[["0x4fed1fc4144c223ae3c1553be203cdfcbd38c581","0x65d21616594825a738bcd08a5227358593a9aaf2","0xd76f7d7d2ede0631ad23e4a01176c0e59878abda"]]`, returnedVals)
}

func TestMatchContract11(t *testing.T) {

	// () -> string[3]
	tg, err := NewTriggerFromFile("../resources/triggers/wac11.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)
	assert.False(t, isMatch)
	assert.Equal(t, "10", matchingVals[0])
	assert.Equal(t, "", matchingVals[1])
	assert.Equal(t, `[["0x4fed1fc4144c223ae3c1553be203cdfcbd38c581","0x65d21616594825a738bcd08a5227358593a9aaf2","0xd76f7d7d2ede0631ad23e4a01176c0e59878abda"]]`, returnedVals)
}

func TestMatchContractUniswap(t *testing.T) {

	tg, err := NewTriggerFromFile("../resources/triggers/wac-uniswap.json")
	assert.NoError(t, err)

	isMatch, _, _ := MatchContract(client, tg, 8496486)
	assert.True(t, isMatch)
}
