package action

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"zoroaster/trigger"
	"zoroaster/utils"
)

const (
	Sender  = `"HAL" <hello@hal.xyz>`
	CharSet = "UTF-8"
)

func sendEmail(iemail sesiface.SESAPI, recipient []string, subject, body string) (*ses.SendEmailOutput, error) {

	input := assembleEmail(recipient, subject, body)

	// Attempt to send the email.
	result, err := iemail.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				log.Error(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				log.Error(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				log.Error(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				log.Error(aerr.Error())
			}
		} else {
			// cast err to awserr.Error to get the Code and Message from an error.
			log.Error(err.Error())
		}
		return nil, err
	}
	log.Debug("\temail sent to: ", recipient)
	return result, nil
}

func assembleEmail(recipients []string, subject, body string) *ses.SendEmailInput {

	toAddresses := make([]*string, len(recipients))
	for i := range recipients {
		toAddresses[i] = aws.String(recipients[i])
	}

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: toAddresses,
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(body),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(Sender),
	}
	return input
}

func fillEmailTemplate(text string, payload trigger.IMatch) string {
	switch m := payload.(type) {
	case trigger.TxMatch:
		return templateTransaction(text, m)
	case trigger.CnMatch:
		return templateContract(text, m)
	case trigger.EventMatch:
		return templateEvent(text, m)
	default:
		log.Warnf("Invalid match type %T", payload)
		return text
	}
}

// ffs...
func ifcPrintf(in interface{}) string {
	switch v := in.(type) {
	case common.Address:
		return strings.ToLower(v.String())
	case []common.Address:
		out := make([]string, len(v))
		for i, a := range v {
			out[i] = strings.ToLower(a.String())
		}
		return fmt.Sprintf("%s", out)
	case []interface{}:
		out := make([]string, len(v))
		for i, inner := range v {
			switch innerV := inner.(type) {
			case common.Address:
				out[i] = strings.ToLower(innerV.String())
			default:
				out[i] = fmt.Sprintf("%v", innerV)
			}
		}
		return fmt.Sprintf("%s", out)
	case reflect.Value:
		a, ok := v.Interface().(common.Address)
		if ok {
			return strings.ToLower(a.String())
		} else {
			return fmt.Sprintf("%v", v)
		}
	case bool:
		return fmt.Sprintf("%v", in)
	default:
		return utils.NormalizeAddress(fmt.Sprintf("%s", in))
	}
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
					text = strings.ReplaceAll(text, fmt.Sprintf("%s", templateToken), ifcPrintf(reflect.ValueOf(actualVal).Index(index)))
				}
			}
		}
	}

	// all other param names
	for k, v := range match.EventParams {
		text = strings.ReplaceAll(text, fmt.Sprintf("!%s", k), ifcPrintf(v))
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
	text = strings.ReplaceAll(text, "$MatchedValue$", fmt.Sprintf("%s", match.MatchedValues))

	// all values
	text = strings.ReplaceAll(text, "$ReturnedValues$", ifcPrintf(match.AllValues))

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
					text = strings.ReplaceAll(text, e, ifcPrintf(reflect.ValueOf(match.AllValues[position]).Index(index)))
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
					text = strings.ReplaceAll(text, e, ifcPrintf(reflect.ValueOf(match.AllValues[0]).Index(index)))
					continue
				}
			}
		}
		// multiple value
		if index < len(match.AllValues) {
			text = strings.ReplaceAll(text, e, ifcPrintf(match.AllValues[index]))
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

	// function params
	if match.DecodedFnArgs != nil {
		var args map[string]json.RawMessage
		err := json.Unmarshal([]byte(*match.DecodedFnArgs), &args)
		if err != nil {
			log.Error(err)
			return text
		}

		// replace !functionParams
		for key, rawJson := range args {
			old := fmt.Sprintf("!%s", key)
			new := fmt.Sprintf("%s", rawJson)
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
	}
	return text
}
