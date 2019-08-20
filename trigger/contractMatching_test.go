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
