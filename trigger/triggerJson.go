package trigger

import (
	"encoding/json"
	"fmt"
	"github.com/HAL-xyz/zoroaster/utils"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

type TriggerJson struct {
	TriggerName  string       `json:"TriggerName"`
	TriggerType  string       `json:"TriggerType"`
	CreationDate string       `json:"CreationDate"`
	ContractABI  string       `json:"ContractABI"`
	ContractAdd  string       `json:"ContractAdd"`
	FunctionName string       `json:"FunctionName,omitempty"`
	Filters      []FilterJson `json:"Filters"`
	Inputs       []InputJson  `json:"Inputs"`
	Outputs      []OutputJson `json:"Outputs"`
	CronJob      CronJobJson  `json:"CronJob"`
}

type FilterJson struct {
	FilterType        string        `json:"FilterType"`
	ParameterName     string        `json:"ParameterName"`
	ParameterType     string        `json:"ParameterType"`
	ParameterCurrency string        `json:"ParameterCurrency,omitempty"`
	FunctionName      string        `json:"FunctionName,omitempty"`
	EventName         string        `json:"EventName,omitempty"`
	Index             *int          `json:"Index"`
	Condition         ConditionJson `json:"Condition"`
}

type ConditionJson struct {
	Predicate         string `json:"Predicate"`
	Attribute         string `json:"Attribute"`
	AttributeCurrency string `json:"AttributeCurrency,omitempty"`
}

type InputJson struct {
	FunctionName   string `json:"FunctionName"`
	ParameterType  string `json:"ParameterType"`
	ParameterValue string `json:"ParameterValue"`
}

type OutputJson struct {
	Index          *int          `json:"Index"`
	ReturnIndex    int           `json:"ReturnIndex"`
	ReturnType     string        `json:"ReturnType"`
	Condition      ConditionJson `json:"Condition"`
	Component      ComponentJson `json:"Component"`
	ReturnCurrency string        `json:"ReturnCurrency,omitempty"`
}

type ComponentJson struct {
	Name string
	Type string
}

type CronJobJson struct {
	Rule     string `json:"Rule"`
	Timezone string `json:"Timezone"`
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

	if tjs.TriggerName == "" {
		return nil, fmt.Errorf("cannot read trigger: missing TriggerName")
	}
	validTriggerTypes := []string{"WatchTransactions", "WatchContracts", "WatchEvents", "CronTrigger"}
	if !utils.IsIn(tjs.TriggerType, validTriggerTypes) {
		return nil, fmt.Errorf("invalid trigger type: %s", tjs.TriggerType)
	}

	if tjs.TriggerType == "WatchContracts" && tjs.FunctionName == "" {
		return nil, fmt.Errorf("cannot read WaC trigger: missing FunctionName")
	}

	if tjs.TriggerType == "CronTrigger" && len(strings.Split(tjs.CronJob.Rule, " ")) != 5 {
		return nil, fmt.Errorf("invalid CronJob expression: %s", tjs.CronJob.Rule)
	}

	timeZoneRgx := regexp.MustCompile(`[-+]\d{4}$`)
	if tjs.TriggerType == "CronTrigger" && !timeZoneRgx.MatchString(tjs.CronJob.Timezone) {
		return nil, fmt.Errorf("invalid CronJob timezone: %s", tjs.CronJob.Timezone)
	}

	trigger := Trigger{
		TriggerName:  tjs.TriggerName,
		TriggerType:  tjs.TriggerType,
		ContractABI:  tjs.ContractABI,
		ContractAdd:  tjs.ContractAdd,
		FunctionName: tjs.FunctionName,
		CronJob: CronJob{
			Rule:     tjs.CronJob.Rule,
			Timezone: tjs.CronJob.Timezone,
		},
	}

	// populate Input/Output for Watch a Contract
	for _, inputJs := range tjs.Inputs {
		in := Input{inputJs.ParameterType, inputJs.ParameterValue}
		trigger.Inputs = append(trigger.Inputs, in)
	}
	for _, outputJs := range tjs.Outputs {
		cond := ConditionOutput{Condition{}, unpackPredicate(outputJs.Condition.Predicate), outputJs.Condition.Attribute, outputJs.Condition.AttributeCurrency}
		if outputJs.Condition.AttributeCurrency != "" {
			if outputJs.ReturnCurrency == "" {
				return nil, fmt.Errorf("missing ReturnCurrency")
			}
			if !common.IsHexAddress(utils.NormalizeAddress(outputJs.ReturnCurrency)) {
				return nil, fmt.Errorf("invalid ReturnCurrency %s", outputJs.ReturnCurrency)
			}
		}
		out := Output{
			Index:       outputJs.Index,
			ReturnIndex: outputJs.ReturnIndex,
			ReturnType:  outputJs.ReturnType,
			Condition:   cond,
			Component: Component{
				Type: outputJs.Component.Type,
				Name: outputJs.Component.Name,
			},
			ReturnCurrency: outputJs.ReturnCurrency,
		}
		trigger.Outputs = append(trigger.Outputs, out)
	}

	// populate Filters for WaT and WaE
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
	if fjs.ParameterCurrency != "" && fjs.Condition.AttributeCurrency == "" {
		return nil, fmt.Errorf("missing AttributeCurrency")
	}
	if fjs.ParameterCurrency == "" && fjs.Condition.AttributeCurrency != "" {
		return nil, fmt.Errorf("missing ParameterCurrency")
	}
	condition, err := makeCondition(fjs)
	if err != nil {
		return nil, err
	}
	f := Filter{
		FilterType:        fjs.FilterType,
		ParameterName:     fjs.ParameterName,
		ParameterType:     fjs.ParameterType,
		ParameterCurrency: fjs.ParameterCurrency,
		FunctionName:      fjs.FunctionName,
		EventName:         fjs.EventName,
		Index:             fjs.Index,
		Condition:         condition,
	}
	return &f, nil
}

func makeCondition(fjs FilterJson) (Conditioner, error) {

	predicate := unpackPredicate(fjs.Condition.Predicate)
	if predicate < 0 && !utils.IsIn(fjs.FilterType, []string{"CheckFunctionCalled", "CheckEventEmitted"}) {
		return nil, fmt.Errorf("unsupported predicate type %s", fjs.Condition.Predicate)
	}
	attribute := fjs.Condition.Attribute
	if len(attribute) < 1 && !utils.IsIn(fjs.FilterType, []string{"CheckFunctionCalled", "CheckEventEmitted"}) {
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
	if fjs.FilterType == "CheckEventParameter" {
		c := ConditionEvent{Condition{}, predicate, fjs.Condition.Attribute, fjs.Condition.AttributeCurrency}
		return c, nil
	}
	if fjs.FilterType == "CheckEventEmitted" {
		c := ConditionEvent{Condition{}, predicate, fjs.Condition.Attribute, fjs.Condition.AttributeCurrency}
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
