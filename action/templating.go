package action

import (
	"fmt"
	"github.com/HAL-xyz/zoroaster/trigger"
	"github.com/HAL-xyz/zoroaster/utils"
	"github.com/sirupsen/logrus"
	"math"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var applyAllTemplateConversions = utils.ComposeStringFns(scaleAmounts, fillHumanTime)

func fillBodyTemplate(text string, payload trigger.IMatch, templateVersion string) string {
	// new template system
	if templateVersion == "v2" {
		rendered, _ := renderTemplateWithData(text, payload.ToTemplateMatch())
		return rendered
	}
	// legacy template system
	switch m := payload.(type) {
	case trigger.TxMatch:
		return applyAllTemplateConversions(templateTransaction(text, m))
	case trigger.CnMatch:
		return applyAllTemplateConversions(templateContract(text, m))
	case trigger.EventMatch:
		return applyAllTemplateConversions(templateEvent(text, m))
	default:
		logrus.Warnf("Invalid match type %T", payload)
		return text
	}
}

func scaleBy(text, scaleBy string) string {
	v := new(big.Float)
	v, ok := v.SetString(text)
	if !ok {
		return text
	}
	scale, _ := new(big.Float).SetString(scaleBy)
	res, _ := new(big.Float).Quo(v, scale).Float64()

	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.4f", math.Ceil(res*10000)/10000), "0"), ".")
}

func scaleAmounts(text string) string {
	r := regexp.MustCompile(`(.{3})Amount\((\d*)(\))`)
	matches := r.FindAllStringSubmatch(text, -1)

	for _, g := range matches {
		switch g[1] {
		case "dec":
			text = strings.ReplaceAll(text, g[0], scaleBy(g[2], "1000000000000000000"))
		case "nin":
			text = strings.ReplaceAll(text, g[0], scaleBy(g[2], "1000000000"))
		case "oct":
			text = strings.ReplaceAll(text, g[0], scaleBy(g[2], "100000000"))
		case "hex":
			text = strings.ReplaceAll(text, g[0], scaleBy(g[2], "1000000"))
		default:
			continue
		}
	}
	return text
}

func fillHumanTime(text string) string {
	r := regexp.MustCompile(`humanTime\((\d*)(\))`)
	matches := r.FindAllStringSubmatch(text, -1)

	for _, g := range matches {
		text = strings.ReplaceAll(text, g[0], unixToHumanTime(g[1]))
	}
	return text
}

func unixToHumanTime(timestamp string) string {
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return timestamp
	}
	unixTimeUTC := time.Unix(i, 0).UTC()
	return unixTimeUTC.Format(time.RFC822)
}

func templateEvent(text string, match trigger.EventMatch) string {
	// standard fields
	blockNumber := fmt.Sprintf("%v", match.Log.BlockNumber)
	blockTimestamp := fmt.Sprintf("%v", match.BlockTimestamp)
	text = strings.ReplaceAll(text, "$BlockNumber$", blockNumber)
	text = strings.ReplaceAll(text, "$BlockTimestamp$", blockTimestamp)
	text = strings.ReplaceAll(text, "$BlockHash$", match.Log.BlockHash)
	text = strings.ReplaceAll(text, "$TransactionHash$", match.Log.TransactionHash)
	text = strings.ReplaceAll(text, "$ContractAddress$", match.Tg.ContractAdd)

	// custom fields
	text = strings.ReplaceAll(text, "$MethodName$", match.Tg.Filters[0].EventName)

	// arrays, such as !ParamName[K]
	arrayRgx := regexp.MustCompile(`!\w+\[\d+]`)
	for _, templateToken := range arrayRgx.FindAllString(text, -1) {
		pos := utils.GetOnlyNumbers(templateToken)
		index, _ := strconv.Atoi(pos)

		cleanToken := strings.Split(templateToken, "[")[0][1:]
		actualVal := match.EventParams[cleanToken]
		if actualVal != nil {
			if reflect.TypeOf(actualVal).Kind() == reflect.Array || reflect.TypeOf(actualVal).Kind() == reflect.Slice {
				if index < reflect.ValueOf(actualVal).Len() {
					text = strings.ReplaceAll(text, fmt.Sprintf("%s", templateToken), fmt.Sprintf("%s", reflect.ValueOf(actualVal).Index(index)))
				}
			}
		}
	}

	// all other param names
	for k, v := range match.EventParams {
		text = strings.ReplaceAll(text, fmt.Sprintf("!%s", k), fmt.Sprintf("%s", v))
	}
	return text
}

func templateContract(text string, match trigger.CnMatch) string {
	// standard fields
	blockNumber := fmt.Sprintf("%v", match.BlockNumber)
	blockTimestamp := fmt.Sprintf("%v", match.BlockTimestamp)
	text = strings.ReplaceAll(text, "$BlockNumber$", blockNumber)
	text = strings.ReplaceAll(text, "$BlockTimestamp$", blockTimestamp)
	text = strings.ReplaceAll(text, "$BlockHash$", match.BlockHash)
	text = strings.ReplaceAll(text, "$ContractAddress$", match.Trigger.ContractAdd)

	// matched value
	text = strings.ReplaceAll(text, "$MatchedValue$", utils.RemoveCharacters(fmt.Sprintf("%s", match.MatchedValues), "[]"))

	// all values
	text = strings.ReplaceAll(text, "$ReturnedValues$", fmt.Sprintf("%s", match.AllValues))

	// indexed value, multiple returns (i.e. $ReturnedValues[K][N]$)
	multIndexedValueRgx := regexp.MustCompile(`\$ReturnedValues\[\d+]\[\d+]\$`)
	multIndexedValues := multIndexedValueRgx.FindAllString(text, -1)
	for _, e := range multIndexedValues {
		positionS := utils.GetOnlyNumbers(strings.Split(e, "][")[0])
		indexS := utils.GetOnlyNumbers(strings.Split(e, "][")[1])
		position, _ := strconv.Atoi(positionS)
		index, _ := strconv.Atoi(indexS)
		if position < len(match.AllValues) {
			switch reflect.TypeOf(match.AllValues[position]).Kind() {
			case reflect.Array, reflect.Slice:
				if index < reflect.ValueOf(match.AllValues[position]).Len() {
					text = strings.ReplaceAll(text, e, fmt.Sprintf("%s", reflect.ValueOf(match.AllValues[position]).Index(index)))
				}
			}
		}
	}

	// indexed value, single return (i.e. $ReturnedValues[N]$)
	indexedValueRgx := regexp.MustCompile(`\$ReturnedValues\[\d+]\$`)
	indexedValues := indexedValueRgx.FindAllString(text, -1)
	for _, e := range indexedValues {
		indexS := utils.GetOnlyNumbers(e)
		index, _ := strconv.Atoi(indexS)
		// single value with array / slice
		if len(match.AllValues) == 1 {
			rt := reflect.TypeOf(match.AllValues[0])
			switch rt.Kind() {
			case reflect.Array, reflect.Slice:
				if index < reflect.ValueOf(match.AllValues[0]).Len() {
					text = strings.ReplaceAll(text, e, fmt.Sprintf("%s", reflect.ValueOf(match.AllValues[0]).Index(index)))
					continue
				}
			}
		}
		// multiple value
		if index < len(match.AllValues) {
			text = strings.ReplaceAll(text, e, fmt.Sprintf("%s", match.AllValues[index]))
		}
	}

	return text
}

func templateTransaction(text string, match trigger.TxMatch) string {
	// standard fields
	blockNumber := fmt.Sprintf("%v", *match.Tx.BlockNumber)
	blockTimestamp := fmt.Sprintf("%v", match.BlockTimestamp)
	gas := fmt.Sprintf("%v", match.Tx.Gas)
	gasPrice := fmt.Sprintf("%v", &match.Tx.GasPrice)
	nonce := fmt.Sprintf("%v", match.Tx.Nonce)

	text = strings.ReplaceAll(text, "$BlockNumber$", blockNumber)
	text = strings.ReplaceAll(text, "$BlockHash$", match.Tx.BlockHash)
	text = strings.ReplaceAll(text, "$TransactionHash$", match.Tx.Hash)
	text = strings.ReplaceAll(text, "$BlockTimestamp$", blockTimestamp)
	text = strings.ReplaceAll(text, "$From$", match.Tx.From)
	text = strings.ReplaceAll(text, "$To$", match.Tx.To)
	text = strings.ReplaceAll(text, "$Value$", match.Tx.Value.String())
	text = strings.ReplaceAll(text, "$Gas$", gas)
	text = strings.ReplaceAll(text, "$GasPrice$", gasPrice)
	text = strings.ReplaceAll(text, "$Nonce$", nonce)

	// function name
	if match.DecodedFnName != nil {
		text = strings.ReplaceAll(text, "$MethodName$", *match.DecodedFnName)
	}

	// replace !functionParams
	for key, value := range match.DecodedFnArgs {
		old := fmt.Sprintf("!%s", key)
		new := fmt.Sprintf("%s", value)
		text = strings.ReplaceAll(text, old, new)
	}

	// replace a function parameter with its indexed value, e.g. given
	// ["0x0df721639ca2f7ff0e1f618b918a65ffb199ac4e",...][0] we want
	// "0x0df721639ca2f7ff0e1f618b918a65ffb199ac4e"

	indexedRgx := regexp.MustCompile(`\[\S*]\[\d*]`)
	indexedParams := indexedRgx.FindAllString(text, -1)

	arrayRgx := regexp.MustCompile(`]\[\d*]`)
	for _, param := range indexedParams {
		array := arrayRgx.FindString(param)         // matches ...][N]
		array = utils.RemoveCharacters(array, "[]") // N
		index, err := strconv.Atoi(array)
		if err != nil {
			return text
		}
		splitElements := strings.Split(param, ",")
		for i, e := range splitElements {
			splitElements[i] = utils.RemoveCharacters(e, "[]")
		}
		if index < len(splitElements) {
			text = strings.Replace(text, param, splitElements[index], 1)
		}
	}

	return text
}
