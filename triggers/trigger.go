package trigger

import (
	"fmt"
	"math/big"
)

type Trigger struct {
	TriggerId   int
	TriggerName string
	TriggerType string // TODO use enum
	ContractABI string
	ContractAdd string
	Filters     []Filter
}

type Filter struct {
	FilterType    string // TODO use enum
	ParameterName string
	ParameterType string // TODO use enum
	Condition     Conditioner
	Index         *int
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
