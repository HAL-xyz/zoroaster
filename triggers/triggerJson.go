package trigger

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
)

type TriggerJson struct {
	TriggerID    int          `json:"TriggerId"`
	TriggerName  string       `json:"TriggerName"`
	TriggerType  string       `json:"TriggerType"`
	CreatorID    int          `json:"CreatorId"`
	CreationDate string       `json:"CreationDate"`
	ContractABI  string       `json:"ContractABI"`
	ContractAdd  string       `json:"ContractAdd"`
	Filters      []FilterJson `json:"Filters"`
}

type FilterJson struct {
	FilterType    string `json:"FilterType"`
	ParameterName string `json:"ParameterName"`
	ParameterType string `json:"ParameterType"`
	Condition     struct {
		Predicate string `json:"Predicate"`
		Attribute string `json:"Attribute"`
	} `json:"Condition"`
	FunctionName string `json:"FunctionName,omitempty"`
	Index        *int   `json:"Index"`
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
	if tjs.TriggerID == 0 {
		return nil, fmt.Errorf("missing TriggerID")
	}

	trigger := Trigger{
		TriggerId:   tjs.TriggerID,
		TriggerName: tjs.TriggerName,
		TriggerType: tjs.TriggerType,
		ContractABI: tjs.ContractABI,
		ContractAdd: tjs.ContractAdd,
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
		ParameterName: fjs.ParameterName,
		ParameterType: fjs.ParameterType,
		FunctionName:  fjs.FunctionName,
		Index:         fjs.Index,
		Condition:     condition,
	}
	return &f, nil
}

func makeCondition(fjs FilterJson) (Conditioner, error) {

	predicate := unpackPredicate(fjs.Condition.Predicate)
	if predicate < 0 && fjs.FilterType != "CheckFunctionCalled" {
		return nil, fmt.Errorf("unsupported predicate type %s", fjs.Condition.Predicate)
	}
	attribute := fjs.Condition.Attribute
	if len(attribute) < 1 && fjs.FilterType != "CheckFunctionCalled" {
		return nil, fmt.Errorf("unsupported attribute type %s", attribute)
	}

	if fjs.FilterType == "BasicFilter" {
		switch fjs.ParameterName {
		case "From":
			return ConditionFrom{Condition{}, predicate, attribute}, nil
		case "To":
			return ConditionTo{Condition{}, predicate, attribute}, nil
		case "Nonce":
			nonce, err := strconv.Atoi(attribute)
			if err != nil {
				return nil, err
			}
			return ConditionNonce{Condition{}, predicate, nonce}, nil
		case "Gas":
			gas, err := strconv.Atoi(attribute)
			if err != nil {
				return nil, err
			}
			return ConditionGas{Condition{}, predicate, gas}, nil
		case "GasPrice":
			gasPrice := new(big.Int)
			_, ok := gasPrice.SetString(attribute, 0)
			if !ok {
				return nil, fmt.Errorf("invalid gasPrice %v", attribute)
			}
			return ConditionGasPrice{Condition{}, predicate, gasPrice}, nil
		case "Value":
			value := new(big.Int)
			_, ok := value.SetString(attribute, 0)
			if !ok {
				return nil, fmt.Errorf("invalid value %v", attribute)
			}
			return ConditionValue{Condition{}, predicate, value}, nil
		default:
			return nil, fmt.Errorf("parameter name not supported: %s", fjs.ParameterName)
		}
	}
	if fjs.FilterType == "CheckFunctionParameter" {
		c := ConditionFunctionParam{Condition{}, predicate, fjs.Condition.Attribute}
		return c, nil
	}
	if fjs.FilterType == "CheckFunctionCalled" {
		c := ConditionFunctionCalled{Condition{}, predicate, fjs.Condition.Attribute}
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
	case "IsIn":
		return IsIn
	default:
		return -1
	}
}
