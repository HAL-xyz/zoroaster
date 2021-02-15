package config

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"log"
)

// SES session
func GetSESSession() *ses.SES {
	awsSess := getSession()
	svc := ses.New(awsSess)
	return svc
}

// AWS session
func getSession() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1"), // SES is only available in Ireland
	})
	_, err = sess.Config.Credentials.Get()
	if err != nil {
		log.Fatal(err)
	}
	return sess
}
