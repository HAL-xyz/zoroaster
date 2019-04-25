package trigger

import (
	"io/ioutil"
	"testing"
)

// Test Json serialization of Triggers

func TestNewTriggerJson(t *testing.T) {

	json, _ := ioutil.ReadFile("../resources/triggers/t1.json")

	_, err := NewTriggerJson(string(json))
	if err != nil {
		t.Error(err)
	}
}

// TODO more tests
func TestTriggerJson_ToTrigger(t *testing.T) {

	json, _ := ioutil.ReadFile("../resources/triggers/t1.json")

	tjs, _ := NewTriggerJson(string(json))
	trig, err := tjs.ToTrigger()
	if err != nil {
		t.Error(err)
	}
	_, ok := trig.Filters[0].Condition.(ConditionTo)
	if ok != true {
		t.Error("Expected type ConditionTo")
	}
}
