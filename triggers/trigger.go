package trigger

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
	I()
}

type Condition struct {
}

// Implements Conditioner interface
func (Condition) I() {}

type Predicate int

const (
	Eq Predicate = iota
	BiggerThan
	SmallerThan
)

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
