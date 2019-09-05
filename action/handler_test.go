package action

import (
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"github.com/magiconair/properties/assert"
	"testing"
	"zoroaster/trigger"
)

// SESAPI mock
type mockSESClient struct {
	sesiface.SESAPI
}

func (m *mockSESClient) SendEmail(*ses.SendEmailInput) (*ses.SendEmailOutput, error) {
	msg := "mock email success"
	return &ses.SendEmailOutput{MessageId: &msg}, nil
}

func TestHandleEmail1(t *testing.T) {

	email := AttributeEmail{
		From:    "hello@wolrd.com",
		To:      []string{"manlio.poltronieri@gmail.com", "$AllValues$"},
		Subject: "Hello World Test",
		Body:    "body",
	}

	payload := &trigger.CnMatch{
		MatchId:        1,
		BlockNo:        1,
		TgId:           1,
		TgUserId:       1,
		MatchedValues:  "",
		AllValues:      "[\"marco@atomic.eu.com\"",
		BlockTimestamp: 123,
	}
	outcome := handleEmail(email, payload, &mockSESClient{})
	expectedPayload := `{"Recipients":["manlio.poltronieri@gmail.com","marco@atomic.eu.com"],"Body":"body"}`
	assert.Equal(t, outcome.Payload, expectedPayload)
}

func TestHandleEmail2(t *testing.T) {

	email := AttributeEmail{
		From:    "hello@wolrd.com",
		To:      []string{"manlio.poltronieri@gmail.com", "$AllValues$"},
		Subject: "Hello World Test",
		Body:    "body",
	}

	payload := &trigger.CnMatch{
		MatchId:        1,
		BlockNo:        1,
		TgId:           1,
		TgUserId:       1,
		MatchedValues:  "",
		AllValues:      "[[\"marco@atomic.eu.com\",\"matteo@atomic.eu.com\",\"not and address\"]]",
		BlockTimestamp: 123,
	}
	outcome := handleEmail(email, payload, &mockSESClient{})
	expectedPayload := `{"Recipients":["manlio.poltronieri@gmail.com","marco@atomic.eu.com","matteo@atomic.eu.com"],"Body":"body"}`
	assert.Equal(t, outcome.Payload, expectedPayload)
}

func TestHandleEmail3(t *testing.T) {

	email := AttributeEmail{
		From:    "hello@wolrd.com",
		To:      []string{"manlio.poltronieri@gmail.com", "$AllValues$"},
		Subject: "Hello World Test",
		Body:    "body",
	}

	payload := &trigger.CnMatch{
		MatchId:        1,
		BlockNo:        1,
		TgId:           1,
		TgUserId:       1,
		MatchedValues:  "",
		AllValues:      "[4#END# \"manlio.poltronieri@gmail.com\"#END# \"hello@world.com\"]",
		BlockTimestamp: 123,
	}
	outcome := handleEmail(email, payload, &mockSESClient{})
	expectedPayload := `{"Recipients":["manlio.poltronieri@gmail.com","hello@world.com"],"Body":"body"}`
	assert.Equal(t, outcome.Payload, expectedPayload)
}
