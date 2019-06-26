package actions

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ses"
	log "github.com/sirupsen/logrus"
	"strings"
	"zoroaster/triggers"
)

const (
	Sender  = "hello@hal.xyz"
	CharSet = "UTF-8"
)

func sendEmail(svc *ses.SES, recipient, subject, body string) (*ses.SendEmailOutput, error) {

	input := assembleEmail(recipient, subject, body)

	// Attempt to send the email.
	result, err := svc.SendEmail(input)

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

func assembleEmail(recipient, subject, body string) *ses.SendEmailInput {
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(recipient),
			},
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

func FillEmailTemplate(template string, ztx *trigger.ZTransaction) string {
	body := template

	// standard fields
	body = strings.ReplaceAll(body, "$BlockNumber", string(*ztx.Tx.BlockNumber))
	body = strings.ReplaceAll(body, "$BlockHash", ztx.Tx.BlockHash)
	body = strings.ReplaceAll(body, "$TransactionHash", ztx.Tx.Hash)
	body = strings.ReplaceAll(body, "$BlockTimestamp", string(ztx.BlockTimestamp))
	body = strings.ReplaceAll(body, "$From", ztx.Tx.From)
	body = strings.ReplaceAll(body, "$To", ztx.Tx.To)
	body = strings.ReplaceAll(body, "$Value", ztx.Tx.Value.String())
	body = strings.ReplaceAll(body, "$Gas", string(ztx.Tx.Gas))
	body = strings.ReplaceAll(body, "$GasPrice", ztx.Tx.GasPrice.String())
	body = strings.ReplaceAll(body, "$Nonce", string(ztx.Tx.Nonce))

	// function name
	if ztx.DecodedFnName != nil {
		body = strings.ReplaceAll(body, "!MethodName", *ztx.DecodedFnName)
	}

	// function params
	if ztx.DecodedFnArgs != nil {
		var args map[string]interface{}
		err := json.Unmarshal([]byte(*ztx.DecodedFnArgs), &args)
		if err != nil {
			log.Error(err)
			return body
		}
		for key, value := range args {
			old := fmt.Sprintf("!%s", key)
			new := fmt.Sprintf("%v", value)
			body = strings.ReplaceAll(body, old, new)
		}
	}
	return body
}
