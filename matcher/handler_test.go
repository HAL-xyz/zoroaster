package matcher

import (
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
	"zoroaster/aws"
	"zoroaster/trigger"
)
import log "github.com/sirupsen/logrus"

func init() {
	log.SetLevel(log.DebugLevel)
}

// HTTP Client mock
type mockHttpClient struct{}

func (m mockHttpClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	resp := http.Response{Status: "200 OK"}
	return &resp, nil
}

// SESAPI mock
type mockSESClient struct {
	sesiface.SESAPI
}

func (m *mockSESClient) SendEmail(*ses.SendEmailInput) (*ses.SendEmailOutput, error) {
	msg := "mock email success"
	return &ses.SendEmailOutput{MessageId: &msg}, nil
}

// IDB mock
type mockDB2 struct {
	aws.IDB
}

func (mockDB2) LogOutcome(outcome *trigger.Outcome, matchId int, watOrWac string) {
	// void
}

func (mockDB2) GetActions(tgId int, userId int) ([]string, error) {
	a1 := `
	{
  		"UserId": 1,
  		"TriggerId": 35,
  		"ActionType": "webhook_post",
  		"Attributes": {
			"URI": "https://webhook.site/4048fc82-5e5b-4095-8106-fa858f9d903d"
  		}
	}`
	a2 := `
	{
		"UserId": 1,
  		"TriggerId": 30,
  		"ActionType": "email",
  		"Attributes": {
    		"To": "hello@gmail.com",
    		"Body": "From tx $TransactionHash$",
    		"Subject": "Trigger 30"
  		}
	}`
	return []string{a1, a2}, nil
}

func TestProcessMatch(t *testing.T) {

	match := trigger.CnMatch{
		MatchId:        1,
		BlockNo:        999,
		BlockTimestamp: 1554828248,
		TgId:           1,
		TgUserId:       1,
		Value:          "0xfffffffffffff",
	}

	outcomes := ProcessMatch(&match, mockDB2{}, &mockSESClient{}, &mockHttpClient{})

	expectedPayload := `{"MatchId":1,"BlockNo":999,"ReturnValue":"0xfffffffffffff","BlockTimestamp":1554828248}`
	expectedOutcome := "200 OK"

	// web hook
	assert.Equal(t, outcomes[0].Payload, expectedPayload)
	assert.Equal(t, outcomes[0].Outcome, expectedOutcome)

	// email
	expPayload := "From tx $TransactionHash$"
	expOutcome := `{
  MessageId: "mock email success"
}`

	assert.Equal(t, outcomes[1].Payload, expPayload)
	assert.Equal(t, outcomes[1].Outcome, expOutcome)
}
