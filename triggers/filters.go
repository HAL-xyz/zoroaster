package trigger

// Filters

type Filter interface{}

type GreaterThan struct {
	value int
}

type SmallerThan struct {
	value int
}

// TODO implement this
type InBetween struct {
	lowerBound int
	upperBound int
}
