package action

import (
	"bytes"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"github.com/onrik/ethrpc"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"zoroaster/trigger"
	"zoroaster/utils"
)

// WEB HOOK TESTS

// HTTP Client mock
type mockHttpClient struct{}

func (m mockHttpClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	resp := http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Body:       ioutil.NopCloser(bytes.NewBufferString("Hello World"))}
	return &resp, nil
}

type mockHttpClient400 struct{}

func (m mockHttpClient400) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	resp := http.Response{
		StatusCode: 400,
		Status:     "400 BAD REQUEST",
		Body:       ioutil.NopCloser(bytes.NewBufferString("Hello World"))}
	return &resp, nil
}

func TestHandleWebHookPost(t *testing.T) {

	tg, _ := trigger.GetTriggerFromFile("../resources/triggers/wac1.json")
	url := AttributeWebhookPost{URI: "https://hal.xyz"}
	cnMatch := trigger.CnMatch{
		tg,
		8888,
		1554828248,
		"0x",
		"uuid",
		[]string{"true"},
		[]interface{}{"true"},
	}

	outcome := handleWebHookPost(url, cnMatch, mockHttpClient{})

	expectedPayload := `{
   "BlockNumber":8888,
   "BlockTimestamp":1554828248,
   "BlockHash":"0x",
   "ContractAdd":"0xbb9bc244d798123fde783fcc1c72d3bb8c189413",
   "FunctionName":"daoCreator",
   "ReturnedData":{
      "MatchedValues":"[\"true\"]",
      "AllValues":"[\"true\"]"
   },
   "TriggerName":"wac 1",
   "TriggerType":"WatchContracts",
   "TriggerUUID":""
}`
	areEq, err := utils.AreEqualJSON(outcome.Payload, expectedPayload)
	assert.NoError(t, err)
	assert.True(t, areEq)
	assert.Equal(t, `{"HttpCode":200,"Response":"200 OK"}`, outcome.Outcome)
	assert.Equal(t, true, outcome.Success)
}

func TestHandleWebhookPostWithTxMatch(t *testing.T) {
	url := AttributeWebhookPost{URI: "https://hal.xyz"}
	tg, _ := trigger.GetTriggerFromFile("../resources/triggers/t1.json")
	tg.ContractABI = "" // otherwise it's a pain to test
	tx, _ := trigger.GetTransactionFromFile("../resources/transactions/tx1.json")
	fnArgs := "{}"
	txMatch := trigger.TxMatch{
		MatchUUID:      "",
		Tg:             tg,
		BlockTimestamp: 1554828248,
		DecodedFnName:  &fnArgs,
		DecodedFnArgs:  &fnArgs,
		Tx:             tx,
	}
	outcome := handleWebHookPost(url, txMatch, mockHttpClient{})

	expectedPayload := `{
  "DecodedData": {
    "FunctionArguments": "{}",
    "FunctionName": "{}"
  },
  "Transaction": {
    "Hash": "0x0641bb18e73d9e874252d3de6993473d176200dc02f4482a64c6540749aecaff",
    "Nonce": 233172,
    "BlockHash": "0xc3fb1f0d4b36593bb2746086955c8c30727c62065e320602c93903ae080bf0af",
    "BlockNumber": 7669714,
	"BlockTimestamp":1554828248,
    "From": "0xabaf790eb22618275fdb47671fc6eab57b2ee04e",
    "To": "0x097b3b7cb01945ba7e76804ddc2fdda2cce6ef43",
    "Gas": 79068,
    "GasPrice": 5579104000,
	"Value":0,
	"InputData":"0x64887334000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000007507d00000000000000000000000000000000000000000000000000000000000000140000000000000000000000000000000000000000000000000000000000000000211170bfa274328fcc100121d00ed000000000000000000000000000000000000000b4e00f124e2110d0600fd00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002efe7f903e9c2d904340000e4001300000000000000000000000000000000000000f1b40008dd1ffdfbfc00020000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"
  },
  "TriggerName": "Basic/To, Basic/Nonce, FP/Address",
  "TriggerType": "WatchTransactions",
  "TriggerUUID": "" 
}`
	ok, err := utils.AreEqualJSON(outcome.Payload, expectedPayload)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, true, outcome.Success)
}

func TestHandleWebHookWrongStuff(t *testing.T) {
	tg, _ := trigger.GetTriggerFromFile("../resources/triggers/wac1.json")
	url := AttributeWebhookPost{URI: "https://foo.zyusfddsiu"}
	cnMatch := trigger.CnMatch{
		tg,
		8888,
		1554828248,
		"0x",
		"uuid",
		[]string{"true"},
		[]interface{}{"true"},
	}
	outcome := handleWebHookPost(url, cnMatch, &http.Client{})

	notFoundPattern := strings.HasPrefix(outcome.Outcome, `{"error":"Post https://foo.zyusfddsiu:`)
	assert.True(t, notFoundPattern)
	assert.Equal(t, false, outcome.Success)
}

type EthMock struct{}

func (cli EthMock) EthGetLogs(params ethrpc.FilterParams) ([]ethrpc.Log, error) {
	return trigger.GetLogsFromFile("../resources/events/logs1.json")
}

func TestHandleWebhookWithEvents(t *testing.T) {

	var client EthMock
	url := AttributeWebhookPost{URI: "https://hal.xyz"}
	tg1, err := trigger.GetTriggerFromFile("../resources/triggers/ev1.json")
	assert.NoError(t, err)
	matches1 := trigger.MatchEvent(client, tg1, 8496661, 1572344236)

	outcome := handleWebHookPost(url, matches1[0], mockHttpClient{})

	expectedPayload := `{
   "ContractAdd":"0xdac17f958d2ee523a2206206994597c13d831ec7",
   "EventName":"Transfer",
   "EventData":{
      "EventParameters":{
         "from":"0xf750f050e5596eb9480523eef7260b1535a689bd",
         "to":"0xcd95b32c98423172e04b1c76841e5a73f4532a7f",
         "value":"677420000"
      },
      "Data":"0x0000000000000000000000000000000000000000000000000000000028609be0",
      "Topics":[
         "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
         "0x000000000000000000000000f750f050e5596eb9480523eef7260b1535a689bd",
         "0x000000000000000000000000cd95b32c98423172e04b1c76841e5a73f4532a7f"
      ]
   },
   "Transaction":{
      "BlockHash":"0xf3d70d822816015f26843d378b8c1d5d5da62f5d346f3e86d91a0c2463d30543",
      "BlockNumber":8496661,
      "BlockTimestamp":1572344236,
      "Hash":"0xf44984a4b533ac0e7b608c881a856eff44ee8c17b9f4dcf8b4ee74e9c10c0455"
   },
   "TriggerName":"Watch an Event",
   "TriggerType":"WatchEvents",
   "TriggerUUID":""
}`
	ok, err := utils.AreEqualJSON(expectedPayload, outcome.Payload)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, true, outcome.Success)
}

// EMAIL TESTS

//SESAPI mock
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
		To:      []string{"manlio.poltronieri@gmail.com", "$ReturnedValues$"},
		Subject: "Hello World Test on block $BlockNumber$",
		Body:    "body",
	}

	tg, _ := trigger.GetTriggerFromFile("../resources/triggers/wac1.json")

	match := trigger.CnMatch{
		Trigger:        tg,
		MatchUUID:      "",
		BlockNumber:    777,
		MatchedValues:  []string{},
		AllValues:      []interface{}{"marco@atomic.eu.com"},
		BlockTimestamp: 123,
		BlockHash:      "0x",
	}
	outcome := handleEmail(email, match, &mockSESClient{})
	expectedPayload := `{
 "Recipients":[
    "manlio.poltronieri@gmail.com",
    "marco@atomic.eu.com"
 ],
 "Body":"body",
 "Subject":"Hello World Test on block 777"
}`
	ok, _ := utils.AreEqualJSON(expectedPayload, outcome.Payload)
	assert.True(t, ok)
	assert.Equal(t, true, outcome.Success)
}

func TestHandleEmail2(t *testing.T) {

	email := AttributeEmail{
		From:    "hello@wolrd.com",
		To:      []string{"manlio.poltronieri@gmail.com", "$ReturnedValues$"},
		Subject: "Matched value is $MatchedValue$",
		Body:    "body",
	}

	tg, _ := trigger.GetTriggerFromFile("../resources/triggers/wac1.json")

	match := trigger.CnMatch{
		Trigger:        tg,
		MatchUUID:      "",
		BlockNumber:    1,
		MatchedValues:  []string{"0x000"},
		AllValues:      []interface{}{"marco@atomic.eu.com", "matteo@atomic.eu.com", "not and address"},
		BlockTimestamp: 123,
		BlockHash:      "0x",
	}
	outcome := handleEmail(email, match, &mockSESClient{})

	expectedPayload := `{
  "Recipients":[
     "manlio.poltronieri@gmail.com",
     "marco@atomic.eu.com",
     "matteo@atomic.eu.com"
  ],
  "Body":"body",
  "Subject":"Matched value is [0x000]"
}`
	ok, _ := utils.AreEqualJSON(expectedPayload, outcome.Payload)
	assert.True(t, ok)
	assert.Equal(t, true, outcome.Success)
}

func TestHandleEmail3(t *testing.T) {

	email := AttributeEmail{
		From:    "hello@wolrd.com",
		To:      []string{"manlio.poltronieri@gmail.com", "$ReturnedValues$"},
		Subject: "Timestamp: $BlockTimestamp$",
		Body:    "body",
	}

	tg, _ := trigger.GetTriggerFromFile("../resources/triggers/wac1.json")

	match := trigger.CnMatch{
		Trigger:        tg,
		MatchUUID:      "",
		BlockNumber:    1,
		MatchedValues:  []string{},
		AllValues:      []interface{}{"manlio.poltronieri@gmail.com", "hello@world.com"},
		BlockTimestamp: 123,
		BlockHash:      "0x",
	}
	outcome := handleEmail(email, match, &mockSESClient{})
	expectedPayload := `{
  "Recipients":[
     "manlio.poltronieri@gmail.com",
     "hello@world.com"
  ],
  "Body":"body",
  "Subject":"Timestamp: 123"
}`
	ok, _ := utils.AreEqualJSON(expectedPayload, outcome.Payload)
	assert.True(t, ok)
	assert.Equal(t, true, outcome.Success)
}

func TestHandleEmailWithEvents(t *testing.T) {

	tg1, err := trigger.GetTriggerFromFile("../resources/triggers/ev1.json")
	assert.NoError(t, err)
	matches := trigger.MatchEvent(EthMock{}, tg1, 8496661, 1572344236)

	matches[0].EventParams["extraAddresses"] = []string{"yes@hal.xyz", "nope@hal.xyz"}

	email := AttributeEmail{
		From:    "hello@haz.xyz",
		To:      []string{"manlio.poltronieri@gmail.com", "!extraAddresses"},
		Subject: "Event email test",
		Body:    "body",
	}

	outcome := handleEmail(email, *matches[0], &mockSESClient{})
	expPayload := `{ 
   "Recipients":[ 
      "manlio.poltronieri@gmail.com",
      "yes@hal.xyz",
      "nope@hal.xyz"
   ],
   "Body":"body",
   "Subject":"Event email test"
}`
	ok, err := utils.AreEqualJSON(expPayload, outcome.Payload)
	assert.NoError(t, err)
	assert.True(t, ok)

	email.To = []string{"manlio.poltronieri@gmail.com", "!extraAddresses[0]"}
	expPayload = `{ 
   "Recipients":[ 
      "manlio.poltronieri@gmail.com",
      "yes@hal.xyz"
   ],
   "Body":"body",
   "Subject":"Event email test"
}`
	outcome = handleEmail(email, *matches[0], &mockSESClient{})

	ok, err = utils.AreEqualJSON(expPayload, outcome.Payload)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, true, outcome.Success)
}

func TestHandleSlackBot(t *testing.T) {

	slackMsg := AttributeSlackBot{
		URI:  "http://...",
		Body: "Hello World Test on block $BlockNumber$",
	}

	tg, _ := trigger.GetTriggerFromFile("../resources/triggers/wac1.json")

	match := trigger.CnMatch{
		Trigger:        tg,
		MatchUUID:      "",
		BlockNumber:    777,
		MatchedValues:  []string{},
		AllValues:      []interface{}{"marco@atomic.eu.com"},
		BlockTimestamp: 123,
		BlockHash:      "0x",
	}
	outcome := handleSlackBot(slackMsg, match, &mockHttpClient{})

	expectedPayload := `{"text":"Hello World Test on block 777"}`
	ok, _ := utils.AreEqualJSON(expectedPayload, outcome.Payload)
	assert.True(t, ok)
	assert.Equal(t, true, outcome.Success)
}

func TestHandleTelegramBot(t *testing.T) {

	payload := AttributeTelegramBot{
		Token:  "123...:xxxxxxxx",
		Body:   "block $BlockNumber$ and some *bold stuff*",
		ChatId: "-408369342",
		Format: "Markdown",
	}

	tg, _ := trigger.GetTriggerFromFile("../resources/triggers/wac1.json")

	match := trigger.CnMatch{
		Trigger:        tg,
		MatchUUID:      "",
		BlockNumber:    777,
		MatchedValues:  []string{},
		AllValues:      []interface{}{"marco@atomic.eu.com"},
		BlockTimestamp: 123,
		BlockHash:      "0x",
	}

	outcome := handleTelegramBot(payload, match, &mockHttpClient{})

	expectedPayload := `{"chat_id":"-408369342","text":"block 777 and some *bold stuff*","parse_mode":"Markdown"}`
	ok, _ := utils.AreEqualJSON(expectedPayload, outcome.Payload)
	assert.True(t, ok)
	assert.Equal(t, true, outcome.Success)

	// test some broken cases

	// 400
	outcomeBadRequest := handleTelegramBot(payload, match, &mockHttpClient400{})
	assert.Equal(t, false, outcomeBadRequest.Success)

	// wrong chat id
	brokenChatId := AttributeTelegramBot{
		Token:  "123...:xxxxxxxx",
		Body:   "Hello World Test on block $BlockNumber$",
		ChatId: "408369343", // missing `-` or `@` prefix
	}
	failedOutcome := handleTelegramBot(brokenChatId, match, &mockHttpClient{})
	assert.Equal(t, false, failedOutcome.Success)

	// wrong formatting
	brokenFormatting := AttributeTelegramBot{
		Token:  "123...:xxxxxxxx",
		Body:   "Hello World Test on block $BlockNumber$",
		ChatId: "-408369343",
		Format: "whoops", // wrong formatting option
	}
	anotherFail := handleTelegramBot(brokenFormatting, match, &mockHttpClient{})
	assert.Equal(t, false, anotherFail.Success)
}

// Mocking the twitter library isn't really worth it atm,
// so we just checked it works once in real life and call it a day.

//func TestHandleTweet(t *testing.T) {
//
//	payload := AttributeTweet{
//		Token:  "",
//		Secret: "",
//		Body:   "la la la la laaaaaa hey juuuude",
//	}
//
//	tg, _ := trigger.GetTriggerFromFile("../resources/triggers/wac1.json")
//
//	match := trigger.CnMatch{
//		Trigger:        tg,
//		MatchUUID:      "",
//		BlockNumber:    777,
//		MatchedValues:  []string{},
//		AllValues:      []interface{}{"marco@atomic.eu.com"},
//		BlockTimestamp: 123,
//		BlockHash:      "0x",
//	}
//
//	outcome := handleTweet(payload, match)
//	_ = outcome
//}
