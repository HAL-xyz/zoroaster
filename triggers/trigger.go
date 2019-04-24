package trigger

import (
	"encoding/json"
	"fmt"
	"github.com/INFURA/go-libs/jsonrpc_client"
	"github.com/ethereum/go-ethereum/common"
	"log"
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

type Trigger struct {
	TriggerId   int
	TriggerName string
	TriggerType string
	ContractABI string
	Filters     []Filter
}

type Filter struct {
	FilterType    string
	ToContract    string
	ParameterName string
	ParameterType string
	Condition     Conditioner
}

type Conditioner interface {
	Yes()
}

type Condition struct {
}

type Predicate int

const (
	Eq Predicate = iota
	BiggerThan
	SmallerThan
)

// Implements Conditioner interface
func (Condition) Yes() {}

type ConditionTo struct {
	Condition
	Predicate Predicate
	Attribute string
}

type ConditionNonce struct {
	Condition
	Predicate Predicate
	Attribute int
}

type FunctionParamCondition struct {
	Condition
	Predicate Predicate
	Attribute string
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

//////////////////// TODO BREAK UP TO DIFFERENT FILE (Json stuff, types, executions)

// TODO ABI
// TODO AND logic
// TODO return type
// TODO tests
func process(triggers []Trigger, block jsonrpc_client.Block) {
	for _, ts := range block.Transactions {
		for _, tg := range triggers {
			for _, f := range tg.Filters {
				ValidateFilter(ts, f, tg.ContractABI)
			}
		}
	}
}

// TODO implement all Conditions and FunctionParamConditions
func ValidateFilter(ts jsonrpc_client.Transaction, f Filter, abi string) bool {

	switch v := f.Condition.(type) {
	case ConditionTo:
		return v.Attribute == *ts.To
	case ConditionNonce:
		switch v.Predicate {
		case Eq:
			return v.Attribute == ts.Nonce
		case BiggerThan:
			return ts.Nonce > v.Attribute
		case SmallerThan:
			return ts.Nonce < v.Attribute
		}
	// TODO extract to func?
	// TODO use typed errors?
	case FunctionParamCondition:
		// check smart contract TO
		if f.ToContract == *ts.To {

			// decode function arguments
			funcArgs := DecodeInputData(ts.Input, abi)

			// extract params
			contractArg := funcArgs[f.ParameterName]
			if contractArg == nil {
				log.Println("Cannot find params in the function")
				return false
			}

			// cast
			if f.ParameterType == "Address" {
				triggerAddress := common.HexToAddress(v.Attribute)
				if triggerAddress == contractArg {
					return true
				}
			} else {
				log.Print("Parameter type not supported", f.ParameterType)
			}
		}
	}
	return false
}
