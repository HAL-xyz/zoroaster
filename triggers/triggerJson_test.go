package trigger

import (
	"testing"
)

const trigger = `
{
   "TriggerId":101,
   "TriggerName":"Basic + Function filters",
   "TriggerType":"WatchTransactions",
   "CreatorId":223,
   "CreationDate":"2019-03-24 17:45:12",
   "ContractABI":"",
   "Filters":[
      {
         "FilterType":"BasicFilter",
         "ParameterName":"To",
         "ParameterType":"Address",
         "Condition":{
            "Predicate":"Eq",
            "Attribute":"0xe8663a64a96169ff4d95b4299e7ae9a76b905b31"
         }
      },
      {
         "FilterType":"CheckFunctionParameter",
   		 "ToContract":"0xe8663a64a96169ff4d95b4299e7ae9a76b905b31",
         "FunctionName":"depositToken",
         "ParameterName":"_to",
         "ParameterType":"Address",
         "Condition":{
            "Predicate":"Eq",
            "Attribute":"0000000000000000000000007abe49749989a53b8d9e584b0ee93bb773ca0b9e"
         }
      },
      {
         "FilterType":"CheckFunctionParameter",
   		 "ToContract":"0xe8663a64a96169ff4d95b4299e7ae9a76b905b31",
         "FunctionName":"depositToken",
         "ParameterName":"_not_there",
         "ParameterType":"Address",
         "Condition":{
            "Predicate":"Eq",
            "Attribute":"0000000000000000000000007abe49749989a53b8d9e584b0ee93bb773ca0b9e"
         }
      },
      {
         "FilterType":"BasicFilter",
         "ParameterName":"Nonce",
         "ParameterType":"Int",
         "Condition":{
            "Predicate":"BiggerThan",
            "Attribute": "1000"
         }
      }
   ]
}`

func TestNewTriggerJson(t *testing.T) {
	_, err := NewTriggerJson(trigger)
	if err != nil {
		t.Error(err)
	}
}

func TestTriggerJson_ToTrigger(t *testing.T) {
	tjs, _ := NewTriggerJson(trigger)
	trig, err := tjs.ToTrigger()
	if err != nil {
		t.Error(err)
	}
	_, ok := trig.Filters[0].Condition.(ConditionTo)
	if ok != true {
		t.Error("Expected type ConditionTo")
	}
}
