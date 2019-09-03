package action

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
	"strings"
	"zoroaster/trigger"
	"zoroaster/utils"
)

const (
	Sender  = "hello@hal.xyz"
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
				Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(body),
				},
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
	case *trigger.TxMatch:
		return templateTransaction(text, m.ZTx)
	case *trigger.CnMatch:
		return templateContract(text, m)
	default:
		log.Warnf("Invalid match type %T", payload)
		return text
	}
}

func templateContract(text string, match *trigger.CnMatch) string {
	// standard fields
	blockNumber := fmt.Sprintf("%v", match.BlockNo)
	blockTimestamp := fmt.Sprintf("%v", match.BlockTimestamp)

	text = strings.ReplaceAll(text, "$BlockNumber$", blockNumber)
	text = strings.ReplaceAll(text, "$BlockTimestamp$", blockTimestamp)

	// all values
	cleanAllValues := strings.ReplaceAll(match.AllValues, "#END#", "")
	cleanAllValues = strings.TrimPrefix(cleanAllValues, "[")
	cleanAllValues = strings.TrimSuffix(cleanAllValues, "]")
	cleanAllValues = strings.ReplaceAll(cleanAllValues, "\"", "")
	text = strings.ReplaceAll(text, "$AllValues$", fmt.Sprintf("%s", cleanAllValues))

	// matched value
	text = strings.ReplaceAll(text, "$MatchedValue$", fmt.Sprintf("%s", match.Value))

	// array indexing

	// Arrays are split by commas, multiple returns objects by #END# token
	var allValues []string
	if strings.HasPrefix(match.AllValues, "[[") {
		allValues = strings.Split(match.AllValues, ",")
	} else {
		allValues = strings.Split(fmt.Sprintf(";;%s;;", match.AllValues), "#END# ")
	}
	for i := range allValues {
		allValues[i] = strings.ReplaceAll(allValues[i], "[[", "")
		allValues[i] = strings.ReplaceAll(allValues[i], "]]", "")
		allValues[i] = strings.ReplaceAll(allValues[i], ";;[", "")
		allValues[i] = strings.ReplaceAll(allValues[i], "];;", "")
		allValues[i] = strings.ReplaceAll(allValues[i], "\"", "")
	}

	// figure out the positions, like [!AllValues[0], AllValues[1]...]
	indexedValueRgx := regexp.MustCompile(`!AllValues\[\d+]`)
	indexedValues := indexedValueRgx.FindAllString(text, -1)
	for _, e := range indexedValues {
		index := utils.GetOnlyNumbers(e)
		position, _ := strconv.Atoi(index)
		if position < len(allValues) {
			text = strings.ReplaceAll(text, e, allValues[position])
		}
	}
	return text
}

func templateTransaction(text string, ztx *trigger.ZTransaction) string {
	// standard fields
	blockNumber := fmt.Sprintf("%v", *ztx.Tx.BlockNumber)
	blockTimestamp := fmt.Sprintf("%v", ztx.BlockTimestamp)
	gas := fmt.Sprintf("%v", ztx.Tx.Gas)
	gasPrice := fmt.Sprintf("%v", &ztx.Tx.GasPrice)
	nonce := fmt.Sprintf("%v", ztx.Tx.Nonce)

	text = strings.ReplaceAll(text, "$BlockNumber$", blockNumber)
	text = strings.ReplaceAll(text, "$BlockHash$", ztx.Tx.BlockHash)
	text = strings.ReplaceAll(text, "$TransactionHash$", ztx.Tx.Hash)
	text = strings.ReplaceAll(text, "$BlockTimestamp$", blockTimestamp)
	text = strings.ReplaceAll(text, "$From$", ztx.Tx.From)
	text = strings.ReplaceAll(text, "$To$", ztx.Tx.To)
	text = strings.ReplaceAll(text, "$Value$", ztx.Tx.Value.String())
	text = strings.ReplaceAll(text, "$Gas$", gas)
	text = strings.ReplaceAll(text, "$GasPrice$", gasPrice)
	text = strings.ReplaceAll(text, "$Nonce$", nonce)

	// function name
	if ztx.DecodedFnName != nil {
		text = strings.ReplaceAll(text, "!MethodName", *ztx.DecodedFnName)
	}

	// function params
	if ztx.DecodedFnArgs != nil {
		var args map[string]json.RawMessage
		err := json.Unmarshal([]byte(*ztx.DecodedFnArgs), &args)
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
