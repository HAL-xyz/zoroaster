package trigger

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type TriggerJson struct {
	TriggerID    int          `json:"TriggerId"`
	TriggerName  string       `json:"TriggerName"`
	TriggerType  string       `json:"TriggerType"`
	CreatorID    int          `json:"CreatorId"`
	CreationDate string       `json:"CreationDate"`
	ContractABI  string       `json:"ContractABI"`
	Filters      []FilterJson `json:"Filters"`
}

type FilterJson struct {
	FilterType    string `json:"FilterType"`
	ToContract    string `json:"ToContract"`
	ParameterName string `json:"ParameterName"`
	ParameterType string `json:"ParameterType"`
	Condition     struct {
		Predicate string `json:"Predicate"`
		Attribute string `json:"Attribute"`
	} `json:"Condition"`
	FunctionName string `json:"FunctionName,omitempty"`
}

// creates a new TriggerJson from JSON
func NewTriggerJson(input string) (*TriggerJson, error) {
	tj := TriggerJson{}
	err := json.Unmarshal([]byte(input), &tj)
	if err != nil {
		return nil, err
	}
	return &tj, nil
}

// converts a TriggerJson to a Trigger
func (tjs *TriggerJson) ToTrigger() (*Trigger, error) {

	trigger := Trigger{
		TriggerId:   tjs.TriggerID,
		TriggerName: tjs.TriggerName,
		TriggerType: tjs.TriggerType,
		ContractABI: tjs.ContractABI,
	}

	// populate the filters in the trigger
	for _, fjs := range tjs.Filters {
		f, err := fjs.ToFilter()
		if err != nil {
			return nil, err
		}
		trigger.Filters = append(trigger.Filters, *f)
	}
	return &trigger, nil
}

// converts a FilterJson to a Filter
func (fjs FilterJson) ToFilter() (*Filter, error) {

	condition, err := makeCondition(fjs)
	if err != nil {
		return nil, err
	}

	f := Filter{
		FilterType:    fjs.FilterType,
		ToContract:    fjs.ToContract,
		ParameterName: fjs.ParameterName,
		ParameterType: fjs.ParameterType,
		Condition:     condition,
	}

	return &f, nil
}

func makeCondition(fjs FilterJson) (Conditioner, error) {

	pred := unpackPredicate(fjs.Condition.Predicate)
	if pred < 0 {
		return nil, fmt.Errorf("unsupported predicate type %s", fjs.Condition.Predicate)
	}

	if fjs.FilterType == "BasicFilter" {
		switch fjs.ParameterName {
		case "To":
			c := ConditionTo{Condition{}, pred, fjs.Condition.Attribute}
			return c, nil
		case "Nonce":
			nonce, err := strconv.Atoi(fjs.Condition.Attribute)
			if err != nil {
				return nil, err
			}
			c := ConditionNonce{Condition{}, pred, nonce}
			return c, nil
		default:
			return nil, fmt.Errorf("parameter name not supported: %s", fjs.ParameterName)
		}
	}
	if fjs.FilterType == "CheckFunctionParameter" {
		c := FunctionParamCondition{Condition{}, pred, fjs.Condition.Attribute}
		return c, nil
	}
	return nil, fmt.Errorf("unsupported filter type %s", fjs.FilterType)
}

func unpackPredicate(p string) Predicate {
	switch p {
	case "Eq":
		return Eq
	case "BiggerThan":
		return BiggerThan
	case "SmallerThan":
		return SmallerThan
	default:
		return -1
	}
}
