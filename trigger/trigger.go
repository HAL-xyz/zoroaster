package trigger

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"math/big"
	"strings"
	"time"
)

type Trigger struct {
	TriggerUUID  string // UUID comes from Postgres
	TriggerName  string
	TriggerType  string
	ContractABI  string
	ContractAdd  string
	Filters      []Filter
	FunctionName string
	Inputs       []Input
	Outputs      []Output
	UserUUID     string
	CronJob      CronJob
	LastFired    time.Time
}

func (tg Trigger) hasBasicFilters() bool {
	for _, f := range tg.Filters {
		if f.FilterType == "BasicFilter" {
			return true
		}
	}
	return false
}

func (tg Trigger) eventName() string {
	// EventName must be the same for every Filter, so we just get the first one
	var eventName string
	for _, f := range tg.Filters {
		if f.FilterType == "CheckEventParameter" || f.FilterType == "CheckEventEmitted" {
			eventName = f.EventName
			break
		}
	}
	return eventName
}

func (tg Trigger) getABIObj() (abi.ABI, error) {
	abiObj, err := abi.JSON(strings.NewReader(tg.ContractABI))
	if err != nil {
		return abi.ABI{}, err
	}
	return abiObj, nil
}

type Filter struct {
	FilterType        string
	ParameterName     string
	ParameterType     string
	ParameterCurrency string
	FunctionName      string
	EventName         string
	Condition         Conditioner
	Index             *int
}

type Input struct {
	ParameterType  string
	ParameterValue string
}

type Output struct {
	Index          *int
	ReturnIndex    int
	ReturnType     string
	Condition      Conditioner
	Component      Component
	ReturnCurrency string
}

type Component struct {
	Type string
	Name string
}

type CronJob struct {
	Rule     string
	Timezone string
}

type Conditioner interface {
	I()
}

type Condition struct {
}

func (Condition) I() {} // Implements Conditioner interface

type ConditionFrom struct {
	Condition
	Predicate Predicate
	Attribute string
}

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

type ConditionGas struct {
	Condition
	Predicate Predicate
	Attribute int
}

type ConditionGasPrice struct {
	Condition
	Predicate Predicate
	Attribute *big.Int
}

type ConditionValue struct {
	Condition
	Predicate Predicate
	Attribute *big.Int
}

type ConditionFunctionParam struct {
	Condition
	Predicate Predicate
	Attribute string
}

type ConditionFunctionCalled struct {
	Condition
	Predicate Predicate
	Attribute string
}

type ConditionOutput struct {
	Condition
	Predicate         Predicate
	Attribute         string
	AttributeCurrency string
}

type ConditionEvent struct {
	Condition
	Predicate         Predicate
	Attribute         string
	AttributeCurrency string
}

type Predicate int

const (
	Eq Predicate = iota
	BiggerThan
	SmallerThan
	IsIn
)

func (p Predicate) String() string {
	return [...]string{"Eq", "BiggerThan", "SmallerThan"}[p]
}

func NewTriggerFromJson(json string) (*Trigger, error) {
	tjs, err := NewTriggerJson(json)
	if err != nil {
		return nil, &triggerCreationError{"cannot parse json trigger:", err}
	}
	tg, err := tjs.ToTrigger()
	if err != nil {
		return nil, &triggerCreationError{"cannot convert TriggerJson to Trigger:", err}
	}
	return tg, nil
}

type triggerCreationError struct {
	where string
	err   error
}

func (e *triggerCreationError) Error() string {
	return fmt.Sprintf("%s error: %s", e.where, e.err)
}
