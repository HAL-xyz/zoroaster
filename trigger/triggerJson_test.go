package trigger

import (
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

// Test Json serialization of Triggers

func TestNewTriggerJson(t *testing.T) {
	json, _ := ioutil.ReadFile("../resources/triggers/t1.json")
	_, err := NewTriggerJson(string(json))
	assert.NoError(t, err)
}

func TestWaT(t *testing.T) {
	json, err := ioutil.ReadFile("../resources/triggers/t1.json")
	assert.NoError(t, err)

	tjs, err := NewTriggerJson(string(json))
	assert.NoError(t, err)
	trig, err := tjs.ToTrigger()
	assert.NoError(t, err)

	_, ok := trig.Filters[0].Condition.(ConditionTo)
	assert.True(t, ok)
}

func TestWaC(t *testing.T) {
	json, err := ioutil.ReadFile("../resources/triggers/wac1.json")
	assert.NoError(t, err)

	tjs, err := NewTriggerJson(string(json))
	assert.NoError(t, err)
	trig, err := tjs.ToTrigger()
	assert.NoError(t, err)

	_, ok := trig.Outputs[0].Condition.(ConditionOutput)
	assert.True(t, ok)
}

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
   "ContractABI":"",
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

func TestCronTrigger(t *testing.T) {
	js := `
{
  "TriggerName":"A time based trigger",
  "TriggerType":"CronTrigger",
  "ContractAdd":"0xbb9bc244d798123fde783fcc1c72d3bb8c189413",
  "ContractABI":"",
  "FunctionName": "balanceOf",
  "Inputs": [
    {
      "ParameterType":"address",
      "ParameterValue": "0xda4a4626d3e16e094de3225a751aab7128e96526"
    }
  ],
  "CronJob": {
	"Rule": "* * * * *",
	"Timezone": "-0800"
  }
}
`
	tg, err := NewTriggerFromJson(js)
	assert.NoError(t, err)

	assert.Equal(t, "* * * * *", tg.CronJob.Rule)
	assert.Equal(t, "-0800", tg.CronJob.Timezone)

	js = `
{
  "TriggerName":"A broken time based trigger",
  "TriggerType":"CronTrigger",
  "ContractAdd":"0xbb9bc244d798123fde783fcc1c72d3bb8c189413",
  "ContractABI":"",
  "FunctionName": "balanceOf",
  "Inputs": [
    {
      "ParameterType":"address",
      "ParameterValue": "0xda4a4626d3e16e094de3225a751aab7128e96526"
    }
  ],
  "CronJob": {
	"Rule": "* * * * *",
	"Timezone": "-08"
  }
}
`
	tg, err = NewTriggerFromJson(js)
	assert.Error(t, err)
}

func TestMacro(t *testing.T) {
	js := `
{
  "Inputs": [
    {
      "ParameterName": "users",
      "ParameterType": "address[]",
      "ParameterValue": "0x894cF868881a0781c20105b076adC644cB252203"
    },
    {
      "ParameterName": "tokens",
      "ParameterType": "address[]",
      "ParameterValue": "$test"
    }
  ],
  "CronJob": {
    "Rule": "*/5 * * * *",
    "Timezone": "-0000"
  },
  "ContractABI": "[{\"constant\":true,\"inputs\":[{\"name\":\"user\",\"type\":\"address\"},{\"name\":\"token\",\"type\":\"address\"}],\"name\":\"tokenBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"users\",\"type\":\"address[]\"},{\"name\":\"tokens\",\"type\":\"address[]\"}],\"name\":\"balances\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"}]", 
  "ContractAdd": "0xb1f8e55c7f64d203c1400b9d8555d050f94adf39",
  "TriggerName": "Balances test",
  "TriggerType": "CronTrigger",
  "FunctionName": "balances"
}
`
	tg, err := NewTriggerFromJson(js)
	assert.NoError(t, err)
	assert.Equal(t, "hello, HAL ;)", tg.Inputs[1].ParameterValue)

}

func TestMap2StringListSorted(t *testing.T) {

	m := map[string]tokenapi.ERC20Token{
		"0x345": {},
		"0x123": {},
		"0xxxx": {},
		"00000": {},
	}
	assert.Equal(t, "00000,0x123,0x345,0xxxx", mapToStringListSorted(m))
}

func TestWaE(t *testing.T) {
	json, err := ioutil.ReadFile("../resources/triggers/ev1.json")
	assert.NoError(t, err)

	tjs, err := NewTriggerJson(string(json))
	assert.NoError(t, err)
	trig, err := tjs.ToTrigger()
	assert.NoError(t, err)

	_, ok := trig.Filters[0].Condition.(ConditionEvent)
	assert.True(t, ok)
	assert.Equal(t, "Transfer", trig.Filters[0].EventName)
}

func TestWaEWithAttributeCurrency(t *testing.T) {
	js := `
{
  "Filters": [
	{
	  "FilterType":"CheckEventParameter",
	  "EventName": "ProtectionAdded",
	  "ParameterName":"_reserveAmount",
	  "ParameterType":"uint256",
	  "ParameterCurrency": "_reserveToken",
	  "Condition":{
		"Predicate":"Eq",
		"Attribute":"677420000",
		"AttributeCurrency": "usd"
	  }
	}
  ],
  "ContractABI": "",
  "ContractAdd": "0xf5fab5dbd2f3bf675de4cb76517d4767013cfb55",
  "TriggerName": "NewAdd1",
  "TriggerType": "WatchEvents"
}
`
	tg, err := NewTriggerFromJson(js)
	assert.NoError(t, err)
	assert.Equal(t, "_reserveToken", tg.Filters[0].ParameterCurrency)

	c, ok := tg.Filters[0].Condition.(ConditionEvent)
	assert.True(t, ok)
	assert.Equal(t, "usd", c.AttributeCurrency)
}

func TestWaEWithAttributeCurrencyMalformed(t *testing.T) {
	js := `
{
  "Filters": [
	{
	  "FilterType":"CheckEventParameter",
	  "EventName": "ProtectionAdded",
	  "ParameterName":"_reserveAmount",
	  "ParameterType":"uint256",
	  "Condition":{
		"Predicate":"Eq",
		"Attribute":"677420000"
		"AttributeCurrency":"usd"
	  }
	}
  ],
  "ContractABI": "",
  "ContractAdd": "0xf5fab5dbd2f3bf675de4cb76517d4767013cfb55",
  "TriggerName": "NewAdd1",
  "TriggerType": "WatchEvents"
}
`
	_, err := NewTriggerFromJson(js)
	assert.Error(t, err)
}

func TestWaCWithAttributeCurrency(t *testing.T) {
	js := `
{
   "Inputs":[
   ],
   "Outputs":[
      {
         "Condition":{
            "Attribute":"10000000000000",
            "Predicate":"BiggerThan",
			"AttributeCurrency": "usd"
         },
         "ReturnType":"tuple",
         "ReturnIndex":0,
		 "ReturnCurrency":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
         "Component":{
               "Name":"d",
               "Type":"uint256"
		 }
      }
   ],
   "ContractABI":"",
   "ContractAdd":"0x8d22F1a9dCe724D8c1B4c688D75f17A2fE2D32df",
   "TriggerName":"some trigger",
   "TriggerType":"WatchContracts",
   "FunctionName":"getSpotPrice"
}
`
	tg, err := NewTriggerFromJson(js)
	assert.NoError(t, err)

	c, ok := tg.Outputs[0].Condition.(ConditionOutput)
	assert.True(t, ok)

	assert.Equal(t, "usd", c.AttributeCurrency)
	assert.Equal(t, "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", tg.Outputs[0].ReturnCurrency)
}

func TestWaCWithAttributeCurrencyMalformed(t *testing.T) {
	js := `
{
   "Inputs":[
   ],
   "Outputs":[
      {
         "Condition":{
            "Attribute":"10000000000000",
            "Predicate":"BiggerThan",
			"AttributeCurrency": "usd"
         },
         "ReturnType":"tuple",
         "ReturnIndex":0,
		 "ReturnCurrency":"0xeeeeeee",
         "Component":{
               "Name":"d",
               "Type":"uint256"
		 }
      }
   ],
   "ContractABI":"",
   "ContractAdd":"0x8d22F1a9dCe724D8c1B4c688D75f17A2fE2D32df",
   "TriggerName":"some trigger",
   "TriggerType":"WatchContracts",
   "FunctionName":"getSpotPrice"
}
`
	_, err := NewTriggerFromJson(js)
	assert.Error(t, err)
}

func TestMalformedJsonTrigger(t *testing.T) {
	// handle broken TriggerJson creation
	_, err := NewTriggerFromJson("def not json")
	assert.Error(t, err)

	// handle broken Trigger creation
	_, err2 := GetTriggerFromFile("../resources/triggers/t11.json")
	assert.Error(t, err2)

	// handle some valid but random json
	_, err3 := NewTriggerFromJson(`{ "hello": 1 }`)
	assert.Error(t, err3)
}
