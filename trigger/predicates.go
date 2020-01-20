package trigger

import (
	"math/big"
	"strconv"
	"strings"
)

func validatePredStringArray(p Predicate, cv []string, tv string, index *int) bool {
	// lowercase
	tv = strings.ToLower(tv)
	for i, v := range cv {
		cv[i] = strings.ToLower(v)
	}
	// remove hex prefix
	tv = strings.TrimPrefix(tv, "0x")
	for i, v := range cv {
		cv[i] = strings.TrimPrefix(v, "0x")
	}
	if index != nil {
		if *index > len(cv) {
			return false
		}
		if p == Eq {
			return cv[*index] == tv
		} else {
			return false
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
	case Eq:
		v, err := strconv.Atoi(tv)
		if err != nil {
			return false
		}
		return len(cv) == v
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

func validatePredBigIntArray(p Predicate, cvs []*big.Int, tv *big.Int, index *int) bool {
	if index != nil {
		if *index > len(cvs) {
			return false
		}
		return validatePredBigInt(p, cvs[*index], tv)
	}
	switch p {
	case SmallerThan:
		return int64(len(cvs)) < tv.Int64()
	case BiggerThan:
		return int64(len(cvs)) > tv.Int64()
	case Eq:
		return int64(len(cvs)) == tv.Int64()
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

func validatePredBoolArray(p Predicate, cvs []bool, tv string, index *int) bool {
	if index != nil {
		if *index > len(cvs) {
			return false
		}
		return validatePredBool(p, cvs[*index], tv)
	}

	switch p {
	case IsIn:
		triggerVal := false
		if strings.ToLower(tv) == "true " {
			triggerVal = true
		}
		for _, v := range cvs {
			if v == triggerVal {
				return true
			}
		}
		return false
	case SmallerThan:
		trigNumericVal, err := strconv.Atoi(tv)
		if err != nil {
			return false
		}
		return len(cvs) < trigNumericVal
	case BiggerThan:
		trigNumericVal, err := strconv.Atoi(tv)
		if err != nil {
			return false
		}
		return len(cvs) > trigNumericVal
	case Eq:
		trigNumericVal, err := strconv.Atoi(tv)
		if err != nil {
			return false
		}
		return len(cvs) == trigNumericVal
	default:
		return false
	}
}

func validatePredUIntArray(p Predicate, cvs []uint8, tv int, index *int) bool {
	if index != nil {
		if *index > len(cvs) {
			return false
		}
		return validatePredInt(p, int(cvs[*index]), tv)
	}
	switch p {
	case SmallerThan:
		return len(cvs) < tv
	case BiggerThan:
		return len(cvs) > tv
	case Eq:
		return len(cvs) == tv
	case IsIn:
		for _, v := range cvs {
			if int(v) == tv {
				return true
			}
		}
		return false
	default:
		return false
	}
}
