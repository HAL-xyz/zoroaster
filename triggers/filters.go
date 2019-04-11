package trigger

// Filters

type Filter interface{}

type GreaterThan struct {
	value int
}

type SmallerThan struct {
	value int
}

type InBetween struct {
	lowerBound int
	upperBound int
}
