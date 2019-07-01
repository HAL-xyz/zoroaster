package trigger

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

// Test Json serialization of Triggers

func TestNewTriggerJson(t *testing.T) {
	json, _ := ioutil.ReadFile("../resources/triggers/t1.json")
	_, err := NewTriggerJson(string(json))
	assert.Nil(t, err)
}

func TestTriggerJson_ToTrigger(t *testing.T) {
	json, _ := ioutil.ReadFile("../resources/triggers/t1.json")

	tjs, _ := NewTriggerJson(string(json))
	trig, err := tjs.ToTrigger()
	assert.Nil(t, err)

	_, ok := trig.Filters[0].Condition.(ConditionTo)
	assert.True(t, ok)
}

func TestMalformedJsonTrigger(t *testing.T) {
	// handle broken TriggerJson creation
	_, ok := NewTriggerFromJson("def not json")
	assert.NotNil(t, ok)

	// handle broken Trigger creation
	_, ok2 := NewTriggerFromFile("../resources/triggers/t11.json")
	assert.NotNil(t, ok2)

	// handle some valid but random json
	_, ok3 := NewTriggerFromJson(`{ "hello": 1 }`)
	assert.NotNil(t, ok3)
}
