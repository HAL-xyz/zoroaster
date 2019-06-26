package actions

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	log "github.com/sirupsen/logrus"
)

const (
	Sender  = "hello@hal.xyz"
	CharSet = "UTF-8"
)

func sendEmail(recipient, subject, body string) (*ses.SendEmailOutput, error) {

	input := assembleEmail(recipient, subject, body)

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("eu-west-1"),
		Credentials: credentials.NewSharedCredentials("", "atomic"),
	})
	_, err = sess.Config.Credentials.Get()
	if err != nil {
		log.Fatal(err)
	}

	// Create an SES session.
	svc := ses.New(sess)

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
