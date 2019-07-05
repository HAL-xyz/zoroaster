package trigger

import (
	"github.com/onrik/ethrpc"
	"github.com/stretchr/testify/assert"
	"testing"
)

var client = ethrpc.New("https://ethshared.bdnodes.net/?auth=_M92hYFzHxR4S1kNbYHfR6ResdtDRqvvLdnm3ZcdAXA")

func TestMatchContract(t *testing.T) {

	// () -> Address
	tg, err := NewTriggerFromFile("../resources/triggers/wac1.json")
	if err != nil {
		t.Error(t)
	}
	assert.True(t, MatchContract(client, tg, 8081000))
}

func TestMatchContract2(t *testing.T) {

	// Address -> uint256
	tg, err := NewTriggerFromFile("../resources/triggers/wac2.json")
	if err != nil {
		t.Error(t)
	}
	assert.True(t, MatchContract(client, tg, 8081000))
}

func TestMatchContract3(t *testing.T) {

	// () -> bool
	tg, err := NewTriggerFromFile("../resources/triggers/wac3.json")
	if err != nil {
		t.Error(t)
	}
	assert.True(t, MatchContract(client, tg, 8081000))
}
