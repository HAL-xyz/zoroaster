package trigger

import "math/big"

type Trigger struct {
	TriggerId   int
	TriggerName string
	TriggerType string // TODO use enum
	ContractABI string
	Filters     []Filter
}

type Filter struct {
	FilterType    string // TODO use enum
	ToContract    string
	ParameterName string
	ParameterType string // TODO use enum
	Condition     Conditioner
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

type FunctionParamCondition struct {
	Condition
	Predicate Predicate
	Attribute string
}

type Predicate int

const (
	Eq Predicate = iota
	BiggerThan
	SmallerThan
)

func (p Predicate) String() string {
	return [...]string{"Eq", "BiggerThan", "SmallerThan"}[p]
}
