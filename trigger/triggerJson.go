package trigger

import (
	"encoding/json"
	"fmt"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/HAL-xyz/zoroaster/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorhill/cronexpr"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type TriggerJson struct {
	TriggerUUID  string       `json:"TriggerUUID"`
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

	if tjs.TriggerType == "CronTrigger" {
		_, err := cronexpr.Parse(tjs.CronJob.Rule)
		if err != nil {
			return nil, fmt.Errorf("invalid CronJob expression: %s", tjs.CronJob.Rule)
		}

		timeZoneRgx := regexp.MustCompile(`[-+]\d{4}$`)
		if !timeZoneRgx.MatchString(tjs.CronJob.Timezone) {
			return nil, fmt.Errorf("invalid CronJob timezone: %s", tjs.CronJob.Timezone)
		}
	}

	trigger := Trigger{
		TriggerUUID:  tjs.TriggerUUID,
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

	// populate Input/Output for Watch a Contract & Cron Trigger
	for _, inputJs := range tjs.Inputs {
		trigger.Inputs = append(trigger.Inputs, *(inputJs.ToInput()))
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

// converts an InputJson to an Input
func (inputJs InputJson) ToInput() *Input {
	return &Input{inputJs.ParameterType, expandMacro(inputJs.ParameterValue)}
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

type Expander func(string) string

func expandMacro(s string) string {

	var macros = map[string]Expander{
		"$test": func(string) string {
			return "hello, HAL ;)"
		},
		"$all_erc20_tokens": func(string) string {
			return mapToStringListSorted(tokenapi.GetTokenAPI().GetAllERC20TokensMap())
		},
	}
	f, ok := macros[s]
	if ok {
		return f(s)
	}
	return s
}

func mapToStringListSorted(m map[string]tokenapi.ERC20Token) string {
	// only use tokens on eth main net
	ethTokensNo := 0
	for _, t := range m {
		if t.ChainId == 1 {
			ethTokensNo++
		}
	}

	var i = 0
	ls := make([]string, ethTokensNo)
	for _, v := range m {
		if v.ChainId == 1 {
			ls[i] = v.Address
			i++
		}
	}
	sort.Strings(ls)
	s := strings.ReplaceAll(fmt.Sprintf("%s", ls), " ", ",")
	return s[1 : len(s)-1]
}
