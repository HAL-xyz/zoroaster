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

// WaT
func TestTriggerJson_ToTrigger(t *testing.T) {
	json, _ := ioutil.ReadFile("../resources/triggers/t1.json")

	tjs, _ := NewTriggerJson(string(json))
	trig, err := tjs.ToTrigger()
	assert.Nil(t, err)

	_, ok := trig.Filters[0].Condition.(ConditionTo)
	assert.True(t, ok)
}

// WaC
func TestTriggerJson_ToTrigger2(t *testing.T) {
	json, _ := ioutil.ReadFile("../resources/triggers/wac1.json")

	tjs, _ := NewTriggerJson(string(json))
	trig, err := tjs.ToTrigger()
	assert.Nil(t, err)

	_, ok := trig.Outputs[0].Condition.(ConditionOutput)
	assert.True(t, ok)
}

// WaC with Components
func TestWaCWithComponents(t *testing.T) {
	js := `
{
   "Inputs":[
   ],
   "Outputs":[
      {
         "Condition":{
            "Attribute":"10000000000000",
            "Predicate":"BiggerThan"
         },
         "ReturnType":"tuple",
         "ReturnIndex":0,
         "Component":{
               "Name":"d",
               "Type":"uint256"
		 }
      }
   ],
   "ContractABI":"[{\"inputs\":[],\"name\":\"getSpotPrice\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"d\",\"type\":\"uint256\"}],\"internalType\":\"struct Decimal.decimal\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
   "ContractAdd":"0x8d22F1a9dCe724D8c1B4c688D75f17A2fE2D32df",
   "TriggerName":"some trigger",
   "TriggerType":"WatchContracts",
   "FunctionName":"getSpotPrice"
}
`
	tg, err := NewTriggerFromJson(js)
	assert.NoError(t, err)
	assert.Equal(t, "d", tg.Outputs[0].Component.Name)
	assert.Equal(t, "uint256", tg.Outputs[0].Component.Type)
}

// WaE
func TestTriggerJson_ToTrigger3(t *testing.T) {
	json, _ := ioutil.ReadFile("../resources/triggers/ev1.json")

	tjs, _ := NewTriggerJson(string(json))
	trig, err := tjs.ToTrigger()
	assert.Nil(t, err)

	_, ok := trig.Filters[0].Condition.(ConditionEvent)
	assert.True(t, ok)
	assert.Equal(t, "Transfer", trig.Filters[0].EventName)
}

func TestMalformedJsonTrigger(t *testing.T) {
	// handle broken TriggerJson creation
	_, ok := NewTriggerFromJson("def not json")
	assert.NotNil(t, ok)

	// handle broken Trigger creation
	_, ok2 := GetTriggerFromFile("../resources/triggers/t11.json")
	assert.NotNil(t, ok2)

	// handle some valid but random json
	_, ok3 := NewTriggerFromJson(`{ "hello": 1 }`)
	assert.NotNil(t, ok3)
}
