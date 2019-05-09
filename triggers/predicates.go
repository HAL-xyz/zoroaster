package trigger

import (
	"math/big"
	"strconv"
	"strings"
)

// TODO return InvalidPredicateError instead of failing silently

func validatePredStringArray(p Predicate, cv []string, tv string) bool {
	// lowercase
	tv = strings.ToLower(tv)
	for i, v := range cv {
		cv[i] = strings.ToLower(v)
	}

	// remove hex prefix
	if strings.HasPrefix(tv, "0x") {
		tv = tv[2:]
	}
	for i, v := range cv {
		if strings.HasPrefix(v, "0x") {
			cv[i] = v[2:]
		}
	}

	switch p {
	case IsIn:
		for _, v := range cv {
			if v == tv {
				return true
			}
		}
		return false
	case SmallerThan:
		v, err := strconv.Atoi(tv)
		if err != nil {
			return false
		}
		return len(cv) < v
	case BiggerThan:
		v, err := strconv.Atoi(tv)
		if err != nil {
			return false
		}
		return len(cv) > v
	default:
		return false
	}
}

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

func validatePredBool(p Predicate, cv bool, tv string) bool {
	if p != Eq {
		return false
	}
	ctVal := "false"
	if cv {
		ctVal = "true"
	}
	return strings.ToLower(tv) == ctVal
}
