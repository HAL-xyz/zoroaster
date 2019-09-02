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

func TestMatchContract1(t *testing.T) {

	// () -> address
	tg, err := NewTriggerFromFile("../resources/triggers/wac1.json")
	if err != nil {
		t.Error(t)
	}
	value, allValues := MatchContract(client, tg, 8387102)
	assert.Equal(t, value, "0x4a574510c7014e4ae985403536074abe582adfc8")
	assert.Equal(t, fmt.Sprint(allValues), "[\"0x4a574510c7014e4ae985403536074abe582adfc8\"]")
}

func TestMatchContract2(t *testing.T) {

	// address -> uint256
	tg, err := NewTriggerFromFile("../resources/triggers/wac2.json")
	if err != nil {
		t.Error(t)
	}
	value, allValues := MatchContract(client, tg, 8387679)
	assert.Equal(t, value, "3876846319093283908984")
	assert.Equal(t, fmt.Sprint(allValues), "[3876846319093283908984]")
}

func TestMatchContract3(t *testing.T) {

	// () -> bool
	tg, err := NewTriggerFromFile("../resources/triggers/wac3.json")
	if err != nil {
		t.Error(t)
	}
	value, allValues := MatchContract(client, tg, 8387102)
	assert.Equal(t, value, "true")
	assert.Equal(t, fmt.Sprint(allValues), "[true]")
}

func TestMatchContract4(t *testing.T) {

	// uint256 -> address
	tg, err := NewTriggerFromFile("../resources/triggers/wac4.json")
	if err != nil {
		t.Error(t)
	}
	value, allValues := MatchContract(client, tg, 8387102)
	assert.Equal(t, value, "0xd4fe7bc31cedb7bfb8a345f31e668033056b2728")
	assert.Equal(t, fmt.Sprint(allValues), "[\"0xd4fe7bc31cedb7bfb8a345f31e668033056b2728\"]")
}

func TestMatchContract5(t *testing.T) {

	// uint16 -> address
	tg, err := NewTriggerFromFile("../resources/triggers/wac5.json")
	if err != nil {
		t.Error(t)
	}
	value, allValues := MatchContract(client, tg, 8387102)
	assert.Equal(t, value, "0x02ca0dfabf5285b0b9d09dfaa241167013355c35")
	assert.Equal(t, fmt.Sprint(allValues), "[\"0x02ca0dfabf5285b0b9d09dfaa241167013355c35\"]")
}

func TestMatchContract6(t *testing.T) {

	// () -> uint256[3]
	tg, err := NewTriggerFromFile("../resources/triggers/wac6.json")
	if err != nil {
		t.Error(t)
	}

	cli := ethrpc.New("https://rinkebyshared.bdnodes.net?auth=dKvc9d7tXrOdmnKK9nsfl119I19PH4GZPbACnbH-QW0")

	value, allValues := MatchContract(cli, tg, 4974958)
	assert.Equal(t, value, "12")
	assert.Equal(t, fmt.Sprint(allValues), "[[4,8,12]]")
}

func TestMatchContract7(t *testing.T) {

	cli := ethrpc.New("https://rinkebyshared.bdnodes.net?auth=dKvc9d7tXrOdmnKK9nsfl119I19PH4GZPbACnbH-QW0")

	// () -> (int128, int128, int128)
	tg, err := NewTriggerFromFile("../resources/triggers/wac7.json")
	if err != nil {
		t.Error(t)
	}
	value, allValues := MatchContract(cli, tg, 4974958)
	assert.Equal(t, value, "4")
	assert.Equal(t, fmt.Sprint(allValues), "[4#END# 8#END# 12]")
}

func TestMatchContract8(t *testing.T) {

	cli := ethrpc.New("https://rinkebyshared.bdnodes.net?auth=dKvc9d7tXrOdmnKK9nsfl119I19PH4GZPbACnbH-QW0")

	// () -> (int128, string, string)
	tg, err := NewTriggerFromFile("../resources/triggers/wac8.json")
	if err != nil {
		t.Error(t)
	}
	value, allValues := MatchContract(cli, tg, 4974958)
	assert.Equal(t, value, "moon")
	assert.Equal(t, fmt.Sprint(allValues), "[4#END# \"sailor\"#END# \"moon\"]")
}

func TestMatchContract9(t *testing.T) {

	cli := ethrpc.New("https://rinkebyshared.bdnodes.net?auth=dKvc9d7tXrOdmnKK9nsfl119I19PH4GZPbACnbH-QW0")

	// () -> string[3]
	tg, err := NewTriggerFromFile("../resources/triggers/wac9.json")
	if err != nil {
		t.Error(t)
	}
	value, allValues := MatchContract(cli, tg, 4974958)
	assert.Equal(t, value, "ciao")
	assert.Equal(t, fmt.Sprint(allValues), "[[\"ciao\",\"come\",\"stai\"]]")
}
