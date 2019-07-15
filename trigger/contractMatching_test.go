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
	value := MatchContract(client, tg, 8081000)
	assert.Equal(t, value, "0x4a574510C7014E4AE985403536074ABE582AdfC8")
}

func TestMatchContract2(t *testing.T) {

	// Address -> uint256
	tg, err := NewTriggerFromFile("../resources/triggers/wac2.json")
	if err != nil {
		t.Error(t)
	}
	value := MatchContract(client, tg, 8081000)
	assert.Equal(t, value, "3876846319093283908984")
}

func TestMatchContract3(t *testing.T) {

	// () -> bool
	tg, err := NewTriggerFromFile("../resources/triggers/wac3.json")
	if err != nil {
		t.Error(t)
	}
	value := MatchContract(client, tg, 8081000)
	assert.Equal(t, value, "true")
}
