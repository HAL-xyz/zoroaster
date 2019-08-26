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

func fillEmailTemplate(body string, payload trigger.IMatch) string {
	switch m := payload.(type) {
	case *trigger.TxMatch:
		return templateTransaction(body, m.ZTx)
	case *trigger.CnMatch:
		return templateContract(body, m)
	default:
		return body
	}
}

func templateContract(body string, match *trigger.CnMatch) string {
	// standard fields
	blockNumber := fmt.Sprintf("%v", match.BlockNo)
	blockTimestamp := fmt.Sprintf("%v", match.BlockTimestamp)

	body = strings.ReplaceAll(body, "$BlockNumber$", blockNumber)
	body = strings.ReplaceAll(body, "$BlockTimestamp$", blockTimestamp)

	// return value
	body = strings.ReplaceAll(body, "$ReturnValues$", match.Value)

	// TODO: support array indexing

	return body
}

func templateTransaction(body string, ztx *trigger.ZTransaction) string {
	// standard fields
	blockNumber := fmt.Sprintf("%v", *ztx.Tx.BlockNumber)
	blockTimestamp := fmt.Sprintf("%v", ztx.BlockTimestamp)
	gas := fmt.Sprintf("%v", ztx.Tx.Gas)
	gasPrice := fmt.Sprintf("%v", &ztx.Tx.GasPrice)
	nonce := fmt.Sprintf("%v", ztx.Tx.Nonce)

	body = strings.ReplaceAll(body, "$BlockNumber$", blockNumber)
	body = strings.ReplaceAll(body, "$BlockHash$", ztx.Tx.BlockHash)
	body = strings.ReplaceAll(body, "$TransactionHash$", ztx.Tx.Hash)
	body = strings.ReplaceAll(body, "$BlockTimestamp$", blockTimestamp)
	body = strings.ReplaceAll(body, "$From$", ztx.Tx.From)
	body = strings.ReplaceAll(body, "$To$", ztx.Tx.To)
	body = strings.ReplaceAll(body, "$Value$", ztx.Tx.Value.String())
	body = strings.ReplaceAll(body, "$Gas$", gas)
	body = strings.ReplaceAll(body, "$GasPrice$", gasPrice)
	body = strings.ReplaceAll(body, "$Nonce$", nonce)

	// function name
	if ztx.DecodedFnName != nil {
		body = strings.ReplaceAll(body, "!MethodName", *ztx.DecodedFnName)
	}

	// function params
	if ztx.DecodedFnArgs != nil {
		var args map[string]json.RawMessage
		err := json.Unmarshal([]byte(*ztx.DecodedFnArgs), &args)
		if err != nil {
			log.Error(err)
			return body
		}

		// replace !functionParams
		for key, rawJson := range args {
			old := fmt.Sprintf("!%s", key)
			new := fmt.Sprintf("%s", rawJson)
			body = strings.ReplaceAll(body, old, new)
		}

		// replace a function parameter with its indexed value, e.g. given
		// ["0x0df721639ca2f7ff0e1f618b918a65ffb199ac4e",...][0] we want
		// "0x0df721639ca2f7ff0e1f618b918a65ffb199ac4e"

		indexedRgx := regexp.MustCompile(`\[\S*]\[\d*]`)
		indexedParams := indexedRgx.FindAllString(body, -1)

		arrayRgx := regexp.MustCompile(`]\[\d*]`)
		for _, param := range indexedParams {
			array := arrayRgx.FindString(param)   // matches ...][N]
			array = removeCharacters(array, "[]") // N
			index, err := strconv.Atoi(array)
			if err != nil {
				return body
			}
			splitElements := strings.Split(param, ",")
			for i, e := range splitElements {
				splitElements[i] = removeCharacters(e, "[]")
			}
			if index < len(splitElements) {
				body = strings.Replace(body, param, splitElements[index], 1)
			}
		}
	}
	return body
}

func removeCharacters(input string, characters string) string {
	filter := func(r rune) rune {
		if strings.IndexRune(characters, r) < 0 {
			return r
		}
		return -1
	}
	return strings.Map(filter, input)
}
