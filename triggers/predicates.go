package trigger

import "math/big"

func validatePredBigInt(p Predicate, cv *big.Int, tv *big.Int) bool {
	switch p {
	case Eq:
		return cv.Cmp(tv) == 0
	case SmallerThan:
		return cv.Cmp(tv) == -1
	case BiggerThan:
		return cv.Cmp(tv) == 1
	}
	return false
}

func validatePredInt(p Predicate, cv int, tv int) bool {
	switch p {
	case Eq:
		return cv == tv
	case SmallerThan:
		return cv < tv
	case BiggerThan:
		return cv > tv
	}
	return false
}