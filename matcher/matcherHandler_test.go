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
	"zoroaster/utils"
)
import log "github.com/sirupsen/logrus"

func init() {
	log.SetLevel(log.DebugLevel)
}

// HTTP Client mock
type mockHttpClient struct{}

func (m mockHttpClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	resp := http.Response{StatusCode: 200, Status: "200 OK"}
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

func (mockDB2) LogOutcome(outcome *trigger.Outcome, matchUUID string) {
	// void
}

func (mockDB2) GetActions(tgUUID string, userUUID string) ([]string, error) {
	a1 := `
	{
  		"UserUUID": 1,
  		"TriggerUUID": 35,
  		"ActionType": "webhook_post",
  		"Attributes": {
			"URI": "https://webhook.site/4048fc82-5e5b-4095-8106-fa858f9d903d"
  		}
	}`
	a2 := `
	{
		"UserUUID": 1,
  		"TriggerUUID": 30,
  		"ActionType": "email",
  		"Attributes": {
    		"To": [
				"hello@gmail.com"
			],
    		"Body": "$BlockNumber$",
    		"Subject": "Trigger 30"
  		}
	}`
	return []string{a1, a2}, nil
}

func TestProcessMatch(t *testing.T) {

	tg, _ := trigger.NewTriggerFromFile("../resources/triggers/wac1.json")

	match := trigger.CnMatch{
		Trigger:        tg,
		MatchUUID:      "",
		BlockNo:        999,
		BlockTimestamp: 1554828248,
		MatchedValues:  "0xfffffffffffff",
		BlockHash:      "0x",
	}

	outcomes := ProcessMatch(match, mockDB2{}, &mockSESClient{}, &mockHttpClient{})

	// web hook
	expPayload := `{
   "BlockNo":999,
   "BlockTimestamp":1554828248,
   "BlockHash":"0x",
   "ContractAdd":"0xbb9bc244d798123fde783fcc1c72d3bb8c189413",
   "FunctionName":"daoCreator",
   "ReturnedData":{
      "MatchedValues":"0xfffffffffffff",
      "AllValues":"null"
   },
   "TriggerName":"wac 1",
   "TriggerType":"WatchContracts",
   "TriggerUUID":""
}`
	expOutcome := `{"HttpCode":200,"Response":"200 OK"}`

	ok, err := utils.AreEqualJSON(expPayload, outcomes[0].Payload)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, expOutcome, outcomes[0].Outcome)

	// email
	expEmailPayload := "{\"Recipients\":[\"hello@gmail.com\"],\"Body\":\"999\"}"
	expEmailOutcome := `{"MessageId":"mock email success"}`
	assert.Equal(t, expEmailPayload, outcomes[1].Payload)
	assert.Equal(t, expEmailOutcome, outcomes[1].Outcome)
}
