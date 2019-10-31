package action

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"zoroaster/trigger"
)

func TestFillEmailTemplate1(t *testing.T) {

	input :=
		`
	{
	  "BlockTimestamp": 1559195951,
	  "DecodedFnArgs": "{\"rs\":[[195,255,1,32,106,90,218,108,77,64,219,113,93,84,74,170,108,77,7,248,209,223,89,192,154,253,182,96,26,237,30,100],[10,102,14,116,173,69,66,21,19,20,28,12,254,158,130,24,163,247,35,225,115,142,47,57,99,103,75,103,10,155,124,246],[71,120,212,176,174,19,184,1,135,161,215,8,198,35,155,6,153,225,208,184,249,142,213,234,82,247,16,251,53,223,47,12],[20,169,70,31,201,12,82,156,97,208,160,81,143,235,64,116,48,247,6,106,51,11,127,45,235,36,172,57,48,185,44,81]],\"tradeAddresses\":[\"0x0df721639ca2f7ff0e1f618b918a65ffb199ac4e\",\"0x0000000000000000000000000000000000000000\",\"0x000331657a1c8752e10d883b885fef46dec0ef84\",\"0xde62454e1f6f7ef04a70a79edd44936aaa5259ae\"],\"tradeValues\":[2854124180013133621850,185809192367215025,10000,6764712445,2000000000000000000000,129,1000000000000000,34641086295351909],\"v\":[27,27]}",
	  "DecodedFnName": "trade",
	  "Tx": {
		"Hash": "0xad07416eb7c8344a385cf32db1377596601a093b15adea1dfc475f0781308912",
		"Nonce": 3325711,
		"BlockHash": "0xeff67a4d650d6f282f7b8bd74ab47c833a95b4f7dad07b89541c759ca44bf852",
		"BlockNumber": 7859205,
		"TransactionIndex": 48,
		"From": "0xa7a7899d944fe658c4b0a1803bab2f490bd3849e",
		"To": "0x2a0c0dbecc7e4d658f48e01e3fa353f44050c208",
		"Value": 333,
		"Gas": 400000,
		"GasPrice": 21000000000,
		"Input": "0xef34358800000000000000000000000000000000000000000000009ab8ee03ec87e06a5a00000000000000000000000000000000000000000000000002942079db0cd1b1000000000000000000000000000000000000000000000000000000000000271000000000000000000000000000000000000000000000000000000001933551fd00000000000000000000000000000000000000000000006c6b935b8bbd400000000000000000000000000000000000000000000000000000000000000000008100000000000000000000000000000000000000000000000000038d7ea4c68000000000000000000000000000000000000000000000000000007b11e26b44a2650000000000000000000000000df721639ca2f7ff0e1f618b918a65ffb199ac4e0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000331657a1c8752e10d883b885fef46dec0ef84000000000000000000000000de62454e1f6f7ef04a70a79edd44936aaa5259ae000000000000000000000000000000000000000000000000000000000000001b000000000000000000000000000000000000000000000000000000000000001bc3ff01206a5ada6c4d40db715d544aaa6c4d07f8d1df59c09afdb6601aed1e640a660e74ad45421513141c0cfe9e8218a3f723e1738e2f3963674b670a9b7cf64778d4b0ae13b80187a1d708c6239b0699e1d0b8f98ed5ea52f710fb35df2f0c14a9461fc90c529c61d0a0518feb407430f7066a330b7f2deb24ac3930b92c51"
	  }
	}
	`
	var ztx trigger.ZTransaction
	err := json.Unmarshal([]byte(input), &ztx)
	assert.NoError(t, err)

	match := trigger.TxMatch{"uuid", nil, &ztx}

	template, err := ioutil.ReadFile("../resources/emails/1-wat-templ.txt")
	assert.NoError(t, err)

	body := fillEmailTemplate(string(template), match)

	expected, err := ioutil.ReadFile("../resources/emails/1-wat-exp.txt")

	assert.NoError(t, err)
	assert.Equal(t, body, string(expected))
}

func TestFillEmailTemplate2(t *testing.T) {
	tg, _ := trigger.NewTriggerFromFile("../resources/triggers/wac1.json")

	match := trigger.CnMatch{
		Trigger:        tg,
		MatchUUID:      "uuid",
		BlockNo:        88888,
		MatchedValues:  "4",
		BlockTimestamp: 123456,
		AllValues:      "[[4,8,12]]",
		BlockHash:      "0x",
	}
	template, err := ioutil.ReadFile("../resources/emails/2-wac-templ.txt")
	assert.NoError(t, err)

	body := fillEmailTemplate(string(template), match)

	expected, err := ioutil.ReadFile("../resources/emails/2-wac-exp.txt")
	assert.NoError(t, err)
	assert.Equal(t, body, string(expected))
}

func TestFillEmailTemplate3(t *testing.T) {
	tg, _ := trigger.NewTriggerFromFile("../resources/triggers/wac1.json")

	match := trigger.CnMatch{
		Trigger:        tg,
		MatchUUID:      "uuid",
		BlockNo:        88888,
		MatchedValues:  "4",
		BlockTimestamp: 123456,
		AllValues:      "[4#END# \"sailor\"#END# \"moon\"]",
		BlockHash:      "0x",
	}
	template, err := ioutil.ReadFile("../resources/emails/3-wac-templ.txt")
	assert.NoError(t, err)

	body := fillEmailTemplate(string(template), match)

	expected, err := ioutil.ReadFile("../resources/emails/3-wac-exp.txt")
	assert.NoError(t, err)
	assert.Equal(t, body, string(expected))
}

func TestFillEmailTemplate4(t *testing.T) {
	tg, _ := trigger.NewTriggerFromFile("../resources/triggers/wac1.json")

	match := trigger.CnMatch{
		Trigger:        tg,
		MatchUUID:      "uuid",
		BlockNo:        88888,
		MatchedValues:  "4",
		BlockTimestamp: 123456,
		AllValues:      "[\"0x4a574510c7014e4ae985403536074abe582adfc8\"]",
		BlockHash:      "0x",
	}

	template := "$AllValues$"
	body := fillEmailTemplate(template, match)

	assert.Equal(t, body, "0x4a574510c7014e4ae985403536074abe582adfc8")
}

// Actually send an email. Commented out bc we only want
// to run it manually
//func TestSendEmail(t *testing.T) {
//
//	sesSession := aws.GetSESSession()
//	to := []string{"manlio.poltronieri@gmail.com", "marco@atomic.eu.com"}
//	subject := "hello from Zoroaster to both of you :)"
//	body := "bla bla"
//	res, err := sendEmail(sesSession, to, subject, body)
//	if err != nil {
//		t.Error(err)
//	}
//	fmt.Println(res)
//}
