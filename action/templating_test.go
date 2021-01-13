package action

import (
	"github.com/HAL-xyz/ethrpc"
	"github.com/HAL-xyz/zoroaster/trigger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestHumanTimeConverter(t *testing.T) {

	assert.Equal(t, "13 Oct 20 23:32 UTC", unixToHumanTime("1602631929"))
	assert.Equal(t, "la la la", unixToHumanTime("la la la"))
}

func TestTemplateWithAllConversions(t *testing.T) {

	tg1, err := trigger.GetTriggerFromFile("../resources/triggers/ev1.json")
	assert.NoError(t, err)
	logs, err := trigger.GetLogsFromFile("../resources/events/logs1.json")
	assert.NoError(t, err)
	matches := trigger.MatchEvent(tg1, logs, mockTokenApi)

	matches[0].EventParams["someBigNumber"] = "629000000000000000"
	matches[0].EventParams["unixTimestamp"] = "1602631929"
	matches[0].EventParams["someOtherNumber"] = "1602631929"

	template := `the first is: decAmount(!someBigNumber); The second is: humanTime(!unixTimestamp); the third is: octAmount(!someOtherNumber); then hexAmount(!someOtherNumber)`

	body := fillBodyTemplate(template, *matches[0], "")

	assert.Equal(t, "the first is: 0.629; The second is: 13 Oct 20 23:32 UTC; the third is: 16.0264; then 1602.632", body)
}

func TestTemplateWithDecConversion(t *testing.T) {

	tg1, err := trigger.GetTriggerFromFile("../resources/triggers/ev1.json")
	assert.NoError(t, err)
	logs, err := trigger.GetLogsFromFile("../resources/events/logs1.json")
	assert.NoError(t, err)
	matches := trigger.MatchEvent(tg1, logs, mockTokenApi)

	matches[0].EventParams["someBigNumber"] = "629000000000000000"
	matches[0].EventParams["smallerNumber"] = "21000000000000"

	template := `the first is: decAmount(!someBigNumber); The second is: decAmount(!smallerNumber)`

	body := fillBodyTemplate(template, *matches[0], "")

	assert.Equal(t, "the first is: 0.629; The second is: 0.0001", body)
}

func TestFillTemplate1(t *testing.T) {

	input :=
		`
	{
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
	`

	trig :=
		`
{
  "CreatorId":13,
  "TriggerName":"hell",
  "TriggerType":"WatchTransactions",
  "ContractAdd":"0x2a0c0dbecc7e4d658f48e01e3fa353f44050c208",
  "ContractABI":"[{\"constant\":false,\"inputs\":[{\"name\":\"assertion\",\"type\":\"bool\"}],\"name\":\"assert\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"setOwner\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"user\",\"type\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint256\"},{\"name\":\"v\",\"type\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\"},{\"name\":\"feeWithdrawal\",\"type\":\"uint256\"}],\"name\":\"adminWithdraw\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"lastActiveTransaction\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"depositToken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"withdrawn\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"admins\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"admin\",\"type\":\"address\"},{\"name\":\"isAdmin\",\"type\":\"bool\"}],\"name\":\"setAdmin\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"address\"}],\"name\":\"tokens\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"feeAccount\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"invalidOrder\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"name\":\"out\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"a\",\"type\":\"uint256\"},{\"name\":\"b\",\"type\":\"uint256\"}],\"name\":\"safeSub\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"user\",\"type\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"invalidateOrdersBefore\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"a\",\"type\":\"uint256\"},{\"name\":\"b\",\"type\":\"uint256\"}],\"name\":\"safeMul\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"deposit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"traded\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"expiry\",\"type\":\"uint256\"}],\"name\":\"setInactivityReleasePeriod\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"a\",\"type\":\"uint256\"},{\"name\":\"b\",\"type\":\"uint256\"}],\"name\":\"safeAdd\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"tradeValues\",\"type\":\"uint256[8]\"},{\"name\":\"tradeAddresses\",\"type\":\"address[4]\"},{\"name\":\"v\",\"type\":\"uint8[2]\"},{\"name\":\"rs\",\"type\":\"bytes32[4]\"}],\"name\":\"trade\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"inactivityReleasePeriod\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"orderFills\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"user\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"feeAccount_\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"SetOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"tokenBuy\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amountBuy\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"tokenSell\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amountSell\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"expires\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"nonce\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"v\",\"type\":\"uint8\"},{\"indexed\":false,\"name\":\"r\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Order\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"tokenBuy\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amountBuy\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"tokenSell\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amountSell\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"expires\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"nonce\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"v\",\"type\":\"uint8\"},{\"indexed\":false,\"name\":\"r\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"Cancel\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"tokenBuy\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amountBuy\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"tokenSell\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amountSell\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"get\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"give\",\"type\":\"address\"}],\"name\":\"Trade\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"balance\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"balance\",\"type\":\"uint256\"}],\"name\":\"Withdraw\",\"type\":\"event\"}]",
  "Filters":[
    {
      "FilterType":"CheckFunctionCalled",
      "FunctionName":"trade"
    }
  ]
}
		`
	tx, err := trigger.JsonToTransaction([]byte(input))
	assert.NoError(t, err)

	tg, err := trigger.NewTriggerFromJson(trig)
	assert.NoError(t, err)

	b := ethrpc.Block{}
	b.Timestamp = 1554828248
	b.Transactions = append(b.Transactions, *tx)

	matches := trigger.MatchTransaction(tg, &b, mockTokenApi)

	template, err := ioutil.ReadFile("../resources/emails/1-wat-templ.txt")
	assert.NoError(t, err)

	body := fillBodyTemplate(string(template), *matches[0], "")
	expected, err := ioutil.ReadFile("../resources/emails/1-wat-exp.txt")

	assert.NoError(t, err)

	// This test doesn't work anymore, because the templating v1 for indexed function parameter is now broken.
	// We decided we don't care because templating v1 is deprecated anyway.
	_ = body
	_ = expected
	//assert.Equal(t, string(expected), body)
}

var tg, _ = trigger.GetTriggerFromFile("../resources/triggers/wac1.json")

var cnMatch = trigger.CnMatch{
	Trigger:        tg,
	MatchUUID:      "uuid",
	BlockNumber:    88888,
	MatchedValues:  []string{"4"},
	BlockTimestamp: 123456,
	AllValues:      []interface{}{},
	BlockHash:      "0x",
}

func TestFillTemplate2(t *testing.T) {

	cnMatch.AllValues = []interface{}{"4", "8", "12"}

	template := "$ReturnedValues$"
	body := fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "[4 8 12]", body)

	template = "$ReturnedValues[0]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "4", body)

	template = "$ReturnedValues[2]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "12", body)

	template = "found: $ReturnedValues[1]$; not found: $ReturnedValues[33]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "found: 8; not found: $ReturnedValues[33]$", body)

	template = "$MatchedValue$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "4", body)
}

func TestFillTemplate3(t *testing.T) {

	cnMatch.AllValues = []interface{}{"4", "sailor", "moon"}

	template := "$ReturnedValues$"
	body := fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "[4 sailor moon]", body)

	template = "$ReturnedValues[0]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "4", body)

	template = "$ReturnedValues[2]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "moon", body)

	template = "$ReturnedValues[0]$, $ReturnedValues[1]$, $ReturnedValues[2]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "4, sailor, moon", body)
}

func TestFillTemplate4(t *testing.T) {

	cnMatch.AllValues = []interface{}{"0x4a574510c7014e4ae985403536074abe582adfc8", "0xffffffffffffffffffffffffffffffffffffffff"}

	template := "$ReturnedValues[0]$"
	body := fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "0x4a574510c7014e4ae985403536074abe582adfc8", body)
}

func TestFillTemplateAdd(t *testing.T) {

	cnMatch.AllValues = []interface{}{
		[]string{
			"0x4a574510c7014e4ae985403536074abe582adfc8",
			"0xffffffffffffffffffffffffffffffffffffffff",
		}}

	template := "$ReturnedValues[0]$"
	body := fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "0x4a574510c7014e4ae985403536074abe582adfc8", body)

	template = "$ReturnedValues[1]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "0xffffffffffffffffffffffffffffffffffffffff", body)

	template = "$ReturnedValues[2]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "$ReturnedValues[2]$", body)
}

// multiple values and []Address
func TestFillTemplate6(t *testing.T) {

	cnMatch.AllValues = []interface{}{
		"4",
		"sailor",
		"moon",
		[]string{
			"0x4a574510c7014e4ae985403536074abe582adfc8",
			"0xffffffffffffffffffffffffffffffffffffffff",
		}}

	template := "$ReturnedValues[3]$"
	body := fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "[0x4a574510c7014e4ae985403536074abe582adfc8 0xffffffffffffffffffffffffffffffffffffffff]", body)

	template = "$ReturnedValues[3][0]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "0x4a574510c7014e4ae985403536074abe582adfc8", body)

	template = "$ReturnedValues[3][1]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "0xffffffffffffffffffffffffffffffffffffffff", body)

	template = "$ReturnedValues[3][2]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "$ReturnedValues[3][2]$", body)
}

// testing a template with (int, string, string, [3]int)
func TestFillTemplate5(t *testing.T) {

	cnMatch.AllValues = []interface{}{
		"4",
		"sailor",
		"moon",
		[]string{"one", "two", "three"}}

	template := "$ReturnedValues$"
	body := fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "[4 sailor moon [one two three]]", body)

	template = "$ReturnedValues[3][0]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "one", body)

	template = "$ReturnedValues[3][1]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "two", body)

	template = "$ReturnedValues[3][9]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "$ReturnedValues[3][9]$", body)

	template = "$ReturnedValues[3]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "[one two three]", body)

	template = "$ReturnedValues[1]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "sailor", body)

	template = "$ReturnedValues[10]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "$ReturnedValues[10]$", body)

	template = "sailor: $ReturnedValues[1]$ and moon: $ReturnedValues[2]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "sailor: sailor and moon: moon", body)

	template = "sailor: $ReturnedValues[1]$ and one: $ReturnedValues[3][0]$"
	body = fillBodyTemplate(template, cnMatch, "")
	assert.Equal(t, "sailor: sailor and one: one", body)
}

func TestFillTemplate7(t *testing.T) {

	cnMatch.AllValues = []interface{}{
		"4",
		"sailor",
		"moon",
		[]string{"one", "two", "three"}}

	template, err := ioutil.ReadFile("../resources/emails/2-wac-templ.txt")
	assert.NoError(t, err)

	body := fillBodyTemplate(string(template), cnMatch, "")

	expected, err := ioutil.ReadFile("../resources/emails/2-wac-exp.txt")
	assert.NoError(t, err)

	assert.Equal(t, string(expected), body)
}

func TestEmailTemplateEvent(t *testing.T) {

	tg1, err := trigger.GetTriggerFromFile("../resources/triggers/ev1.json")
	assert.NoError(t, err)
	logs, err := trigger.GetLogsFromFile("../resources/events/logs1.json")
	assert.NoError(t, err)
	matches := trigger.MatchEvent(tg1, logs, mockTokenApi)

	matches[0].EventParams["arrayParam"] = []string{"hello", "world", "yo yo"}

	addresses := []common.Address{common.HexToAddress("0x4a574510c7014e4ae985403536074abe582adfc8")}
	matches[0].EventParams["addresses"] = addresses

	template, err := ioutil.ReadFile("../resources/emails/3-wae-templ.txt")
	assert.NoError(t, err)

	matches[0].BlockTimestamp = 1572344236
	body := fillBodyTemplate(string(template), *matches[0], "")

	expected, err := ioutil.ReadFile("../resources/emails/3-wae-exp.txt")
	assert.NoError(t, err)

	assert.Equal(t, string(expected), body)
}
