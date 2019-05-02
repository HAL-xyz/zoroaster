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
	default:
		return false
	}
}

func validatePredBigIntArray(p Predicate, cvs []*big.Int, tv *big.Int) bool {
	switch p {
	case SmallerThan:
		return int64(len(cvs)) < tv.Int64()
	case BiggerThan:
		return int64(len(cvs)) > tv.Int64()
	case IsIn:
		for _, v := range cvs {
			if v.Cmp(tv) == 0 {
				return true
			}
		}
		return false
	default:
		return false
	}
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