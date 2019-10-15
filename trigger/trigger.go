package trigger

import (
	"fmt"
	"math/big"
)

type Trigger struct {
	TriggerUUID string // UUID comes from Postgres
	TriggerName string
	TriggerType string
	ContractABI string
	ContractAdd string
	Filters     []Filter
	MethodName  string
	Inputs      []Input
	Outputs     []Output
	UserUUID    string
}

type Filter struct {
	FilterType    string
	ParameterName string
	ParameterType string
	FunctionName  string
	Condition     Conditioner
	Index         *int
}

type Input struct {
	ParameterType  string
	ParameterValue string
}

type Output struct {
	Index      *int
	ReturnType string
	Condition  Conditioner
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
