package action

import (
	"github.com/HAL-xyz/ethrpc"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/HAL-xyz/zoroaster/trigger"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
	"io/ioutil"
	"math/big"
	"os"
	"testing"
)

func init() {
	defer gock.Off()
}

func TestTxMatching(t *testing.T) {

	block, _ := trigger.GetBlockFromFile("../resources/blocks/block1.json")
	tg, _ := trigger.GetTriggerFromFile("../resources/triggers/t2.json")

	matches := trigger.MatchTransaction(tg, block, mockTokenApi)

	templateText := `
Block Number is: {{ .Block.Number }}
Block Timestamp is: {{ .Block.Timestamp }}
Block Hash is: {{ .Block.Hash }}
Contract Address is: {{ .Contract.Address }}
Contract method name is: {{ .Contract.MethodName }}
Contract parameter "_to" is: {{index .Contract.MethodParameters "_to"}}
Contract parameter "_value" is: {{index .Contract.MethodParameters "_value"}}
Transaction from is: {{.Tx.From}}
Transaction gas is: {{.Tx.Gas}}
Transaction gas price is: {{.Tx.GasPrice}}
Transaction nonce is: {{.Tx.Nonce}}
Transaction to is: {{.Tx.To}}
Transaction hash is: {{.Tx.Hash}}
Transaction value is: {{.Tx.Value}}
Transaction input data is: {{.Tx.InputData}}
`

	expectedOutcome := `
Block Number is: 7535077
Block Timestamp is: 1554828248
Block Hash is: 0xb972fb8fe7a2aca471fa649e790ac51f59f920a2b71ec522aee606f1ccc99f6e
Contract Address is: 0x174bfa6600bf90c885c7c01c7031389ed1461ab9
Contract method name is: transfer
Contract parameter "_to" is: 0xfea2f9433058cd555fd67cdde8efd7e6031e56c0
Contract parameter "_value" is: 4000000000000000000
Transaction from is: 0x3d2339bf362a9b0f8ef3ca0867bd73f350ed66ac
Transaction gas is: 115960
Transaction gas price is: 7000000000
Transaction nonce is: 414
Transaction to is: 0x174bfa6600bf90c885c7c01c7031389ed1461ab9
Transaction hash is: 0x42c8de77ef5d76f36aea6e051b9059ece6e34619d9fb4a1d97f3224d5c990a67
Transaction value is: 0
Transaction input data is: 0xa9059cbb000000000000000000000000fea2f9433058cd555fd67cdde8efd7e6031e56c00000000000000000000000000000000000000000000000003782dace9d900000
`
	rendered, err := RenderTemplateWithData(templateText, matches[0].ToTemplateMatch())
	assert.NoError(t, err)
	assert.Equal(t, expectedOutcome, rendered)

	exampleUI :=
		`
Hello, welcome to the new template system. It will be soon wrapped up in a nice UI, but for the time being you can access fields manually using the Go template syntax.

You can also find the full documentation here: https://dev.hal.xyz/how-it-works/actions/templating

Contract Address is: {{.Contract.Address}}
Block Number is: {{.Block.Number}}
Block Timestamp is: {{.Block.Timestamp}}
Block Hash is: {{.Block.Hash }}
Transaction from is: {{.Tx.From}}
Transaction gas is: {{.Tx.Gas}}
Transaction gas price is: {{.Tx.GasPrice}}
Transaction nonce is: {{.Tx.Nonce}}
Transaction to is: {{.Tx.To}}
Transaction hash is: {{.Tx.Hash}}
Transaction value is: {{.Tx.Value}}
Transaction input data is: {{.Tx.InputData}}

If the transaction calls a function, you can access the function name like this: {{ .Contract.MethodName }}

You can also access the various function parameters like this: {{.Contract.MethodParameters.ParameterName}}

We also support a bunch of handy functions to manipulate different values:

{{ fromWei .Tx.Value 9 }} divides a value by 10^9.

{{ humanTime .Block.Timestamp }} prints a timestamp in some human readable format

{{ hexToASCII "0x4920686176652031303021" }} guess ;)

{{ hexToInt "0xea" }}

{{ etherscanTxLink .Tx.Hash }} creates an Etherscan transaction link

{{ etherscanTokenLink .Contract.Address }}

{{ etherscanAddressLink .Contract.Address }}

`
	_, err = RenderTemplateWithData(exampleUI, matches[0].ToTemplateMatch())
	assert.NoError(t, err)
}

func TestContractMatching(t *testing.T) {

	var cnMatch = trigger.CnMatch{
		Trigger:        tg,
		MatchUUID:      "uuid",
		BlockNumber:    88888,
		MatchedValues:  []string{},
		BlockTimestamp: 123456,
		AllValues:      []interface{}{},
		BlockHash:      "0x66666",
	}

	cnMatch.MatchedValues = []string{"hello", "world"}
	cnMatch.AllValues = []interface{}{"4", "8", "12", []string{"a", "b", "c"}}

	templateText := `
Block Number is: {{ .Block.Number }}
Matched Values are: {{ .Contract.MatchedValues }}
First matched value is {{ index .Contract.MatchedValues 0 }}
Returned Values are: {{ .Contract.ReturnedValues }}
First returned value is {{ index .Contract.ReturnedValues 0 }}
First returned value of inner array is {{ index (index .Contract.ReturnedValues 3) 0 }}
Testing uppercase function is {{ upperCase (index .Contract.MatchedValues 0) }}
Out of bound value is {{ index .Contract.MatchedValues 9 }}
`

	expectedOutcome := `
Block Number is: 88888
Matched Values are: [hello world]
First matched value is hello
Returned Values are: [4 8 12 [a b c]]
First returned value is 4
First returned value of inner array is a
Testing uppercase function is HELLO
Out of bound value is `

	rendered, err := RenderTemplateWithData(templateText, cnMatch.ToTemplateMatch())
	assert.Equal(t, expectedOutcome, rendered)
	assert.Error(t, err) // error isn't nil because of the out of bound indexing

	exampleUI :=
		`
Hello, welcome to the new template system. It will be soon wrapped up in a nice UI, but for the time being you can access fields manually using the Go template syntax.

You can also find the full documentation here: https://dev.hal.xyz/how-it-works/actions/templating

Block Number is: {{ .Block.Number }}
Block Timestamp is: {{.Block.Timestamp}}
Block Hash is: {{.Block.Hash }}

Contract Address is: {{.Contract.Address}}
All values returned by the function: {{ .Contract.ReturnedValues }}
Matched values only are: {{ .Contract.MatchedValues }}

If .Contract.ReturnedValues returns more than one value, you can access a specific value like this:
First returned value is {{ index .Contract.ReturnedValues 0 }}
Nesting also works: {{ index (index .Contract.ReturnedValues 3) 0 }}

We also support a bunch of handy functions to manipulate different values:

{{ humanTime .Block.Timestamp }} prints a timestamp in some human readable format

{{ hexToASCII "0x4920686176652031303021" }} guess ;)

{{ hexToInt "0xea" }}

{{ etherscanTokenLink .Contract.Address }}

{{ etherscanAddressLink .Contract.Address }}
`
	_, err = RenderTemplateWithData(exampleUI, cnMatch.ToTemplateMatch())
	assert.NoError(t, err)
}

func TestEventMatching(t *testing.T) {

	tg1, err := trigger.GetTriggerFromFile("../resources/triggers/ev1.json")
	assert.NoError(t, err)
	logs, err := trigger.GetLogsFromFile("../resources/events/logs1.json")
	assert.NoError(t, err)
	matches := trigger.MatchEvent(tg1, logs, []ethrpc.Transaction{}, mockTokenApi)

	matches[0].EventParams["arrayParam"] = []string{"hello", "world", "yo yo"}

	templateText := `
Block Number is: {{ .Block.Number }}
Event name is {{ .Contract.EventName }}
Event from param is: {{ .Contract.EventParameters.to }}
Event value param is: {{ .Contract.EventParameters.value }}
First element in array parameter is: {{ index (.Contract.EventParameters.arrayParam) 0 }}
Missing param is: {{ .Contract.EventParameters.missing }}
Transaction hash is {{ .Tx.Hash }}
`
	expectedOutcome := `
Block Number is: 8496661
Event name is Transfer
Event from param is: 0xcd95b32c98423172e04b1c76841e5a73f4532a7f
Event value param is: 677420000
First element in array parameter is: hello
Missing param is: <no value>
Transaction hash is 0xf44984a4b533ac0e7b608c881a856eff44ee8c17b9f4dcf8b4ee74e9c10c0455
`
	rendered, err := RenderTemplateWithData(templateText, matches[0].ToTemplateMatch())
	assert.NoError(t, err)
	assert.Equal(t, expectedOutcome, rendered)

	tmpl := `
{{ if eq .Contract.EventParameters.to "0xcd95b32c98423172e04b1c76841e5a73f4532a7f" }}
	the amount in DAI Is {{ fromWei .Contract.EventParameters.value 18 }}
{{ else }}
	{{ range .Contract.EventParameters.arrayParam }}
		looping through: {{ . }}
	{{ end }}
{{ end }}
`
	_, err = RenderTemplateWithData(tmpl, matches[0].ToTemplateMatch())
	assert.NoError(t, err)

	exampleUI :=
		`Hello, welcome to the new template system. It will be soon wrapped up in a nice UI, but for the time being you can access fields manually using the Go template syntax.

You can also find the full documentation here: https://dev.hal.xyz/how-it-works/actions/templating

Block Number is: {{ .Block.Number }}
Block Timestamp is: {{.Block.Timestamp}}
Block Hash is: {{.Block.Hash }}

Contract Address is: {{.Contract.Address}}

Event name is {{ .Contract.EventName }}

To access specific parameters of an event, such as "from" and "value":
Event from param is: {{ .Contract.EventParameters.to }}
Event value param is: {{ .Contract.EventParameters.value }}

If the parameter of an event is an array, you can access specific values like this:
First element in array parameter is: {{ index (.Contract.EventParameters.arrayParam) 0 }}

We also support a bunch of handy functions to manipulate different values:

{{ fromWei .Tx.Value 9 }} divides a value by 10^9.

{{ humanTime .Block.Timestamp }} prints a timestamp in some human readable format

{{ hexToASCII "0x4920686176652031303021" }} guess ;)

{{ hexToInt "0xea" }}

{{ etherscanTxLink "0x..." }} creates an Etherscan transaction link

{{ etherscanTokenLink .Contract.Address }}

{{ etherscanAddressLink .Contract.Address }}
`
	_, err = RenderTemplateWithData(exampleUI, matches[0].ToTemplateMatch())
	assert.NoError(t, err)
}

func TestTemplateFunctions(t *testing.T) {

	template := "{{ hexToASCII . }}"
	rendered, err := RenderTemplateWithData(template, "0x4920686176652031303021")
	assert.NoError(t, err)
	assert.Equal(t, "I have 100!", rendered)

	template = "{{ hexToASCII . }}"
	rendered, err = RenderTemplateWithData(template, "0x534e580000000000000000000000000000000000000000000000000000000000")
	assert.NoError(t, err)
	assert.Equal(t, "SNX", rendered)

	template = "{{ hexToInt . }}"
	rendered, err = RenderTemplateWithData(template, "0xEA")
	assert.NoError(t, err)
	assert.Equal(t, "234", rendered)

	template = "{{ hexToInt . }}"
	rendered, err = RenderTemplateWithData(template, "100")
	assert.NoError(t, err)
	assert.Equal(t, "100", rendered)

	template = "{{ etherscanTxLink . }}"
	rendered, err = RenderTemplateWithData(template, "0xfdb96f7387559ebfc41e88e21962414eb527484f578ce87996f8733352ab2ee7")
	assert.NoError(t, err)
	assert.Equal(t, "https://etherscan.io/tx/0xfdb96f7387559ebfc41e88e21962414eb527484f578ce87996f8733352ab2ee7", rendered)

	template = "{{ etherscanAddressLink . }}"
	rendered, err = RenderTemplateWithData(template, "0xfdb96f7387559ebfc41e88e21962414eb527484f578ce87996f8733352ab2ee7")
	assert.NoError(t, err)
	assert.Equal(t, "https://etherscan.io/address/0xfdb96f7387559ebfc41e88e21962414eb527484f578ce87996f8733352ab2ee7", rendered)

	template = "{{ etherscanTokenLink . }}"
	rendered, err = RenderTemplateWithData(template, "0xfdb96f7387559ebfc41e88e21962414eb527484f578ce87996f8733352ab2ee7")
	assert.NoError(t, err)
	assert.Equal(t, "https://etherscan.io/token/0xfdb96f7387559ebfc41e88e21962414eb527484f578ce87996f8733352ab2ee7", rendered)

	template = "{{ fromWei . 18 }}"
	rendered, err = RenderTemplateWithData(template, "629700000000000000")
	assert.NoError(t, err)
	assert.Equal(t, "0.6297", rendered)

	template = "{{ fromWei . 6 }}"
	rendered, err = RenderTemplateWithData(template, "629000000000000000")
	assert.NoError(t, err)
	assert.Equal(t, "629000000000", rendered)

	template = "{{ fromWei . 6 }}"
	rendered, err = RenderTemplateWithData(template, big.NewInt(629000000000000000))
	assert.NoError(t, err)
	assert.Equal(t, "629000000000", rendered)

	template = "{{ fromWei . 6 }}"
	rendered, err = RenderTemplateWithData(template, 629000000000000000)
	assert.NoError(t, err)
	assert.Equal(t, "629000000000", rendered)

	template = `{{ fromWei . "6" }}`
	rendered, err = RenderTemplateWithData(template, 629000000000000000)
	assert.NoError(t, err)
	assert.Equal(t, "629000000000", rendered)

	template = "{{ humanTime . }}"
	rendered, err = RenderTemplateWithData(template, "1602631929")
	assert.NoError(t, err)
	assert.Equal(t, "13 Oct 20 23:32 UTC", rendered)

	template = "{{ humanTime . }}"
	rendered, err = RenderTemplateWithData(template, 1602631929)
	assert.NoError(t, err)
	assert.Equal(t, "13 Oct 20 23:32 UTC", rendered)

	template = `{{ humanTime . "3:04:05 PM" }}`
	rendered, err = RenderTemplateWithData(template, 1602631929)
	assert.NoError(t, err)
	assert.Equal(t, "11:32:09 PM", rendered)

	template = `{{ formatNumber "10000" 2 }}`
	rendered, err = RenderTemplateWithData(template, nil)
	assert.NoError(t, err)
	assert.Equal(t, "10,000.00", rendered)

	// stringified floating point numbers are converted in a strange way so that this happens:
	template = `{{ if ge "100" "100.0" }} GE {{ else }} Not-GE {{ end }}`
	rendered, err = RenderTemplateWithData(template, nil)
	assert.NoError(t, err)
	assert.Equal(t, " Not-GE ", rendered)
	// we use floatToInt to truncate floats and compare correctly
	template = `{{ if ge 100 (floatToInt "100.0") }} GE {{ else }} Not-GE {{ end }}`
	rendered, err = RenderTemplateWithData(template, nil)
	assert.NoError(t, err)
	assert.Equal(t, " GE ", rendered)
}

func TestMathFunctions(t *testing.T) {
	template := `{{ add 10 "20" 30 }}`
	rendered, err := RenderTemplateWithData(template, nil)
	assert.NoError(t, err)
	assert.Equal(t, "60", rendered)

	template = `{{ sub 100 "20" }}`
	rendered, err = RenderTemplateWithData(template, nil)
	assert.NoError(t, err)
	assert.Equal(t, "80", rendered)

	template = `{{ mul 10.55 "4" }}`
	rendered, err = RenderTemplateWithData(template, nil)
	assert.NoError(t, err)
	assert.Equal(t, "42.2", rendered)

	template = `{{ div 2 "3" }}`
	rendered, err = RenderTemplateWithData(template, nil)
	assert.NoError(t, err)
	assert.Equal(t, "0.6666666666666666667", rendered)

	template = `{{ round (div 2 "3") 2 }}`
	rendered, err = RenderTemplateWithData(template, nil)
	assert.NoError(t, err)
	assert.Equal(t, "0.67", rendered)

	template = `{{ round (mul (div 2 3) 100) 3 }}`
	rendered, err = RenderTemplateWithData(template, nil)
	assert.NoError(t, err)
	assert.Equal(t, "66.667", rendered)

	template = `{{ pow 2 8 }}`
	rendered, err = RenderTemplateWithData(template, nil)
	assert.NoError(t, err)
	assert.Equal(t, "256", rendered)

	template = `{{ pow "2" "8" }}`
	rendered, err = RenderTemplateWithData(template, nil)
	assert.NoError(t, err)
	assert.Equal(t, "256", rendered)

	template = `{{ pow 1.0000560291 355 }}`
	rendered, err = RenderTemplateWithData(template, nil)
	assert.NoError(t, err)
	assert.Equal(t, "1.02008889279936419637271196253370647336257590813508215370058066207976589819699", rendered)

	template = `{{ round (mul (sub (pow (add (mul (div 9727274683 1000000000000000000) 5760) 1) 364) 1) 100) 2 }}`
	rendered, err = RenderTemplateWithData(template, nil)
	assert.NoError(t, err)
	assert.Equal(t, "2.06", rendered)

	template = `{{ percentageVariation 200 100 }}`
	rendered, err = RenderTemplateWithData(template, nil)
	assert.NoError(t, err)
	assert.Equal(t, "100.00%", rendered)

	template = `{{ percentageVariation 100 150 }}`
	rendered, err = RenderTemplateWithData(template, nil)
	assert.NoError(t, err)
	assert.Equal(t, "-33.33%", rendered)
}

func TestERC20Functions(t *testing.T) {

	template := "{{ symbol . }}"
	rendered, err := RenderTemplateWithData(template, "0x6b175474e89094c44da98b954eedeac495271d0f")
	assert.NoError(t, err)
	assert.Equal(t, "DAI", rendered)

	template = "{{ symbol . }}"
	rendered, err = RenderTemplateWithData(template, "0x0000000000000000000000000000000000000000")
	assert.NoError(t, err)
	assert.Equal(t, "ETH", rendered)

	template = "{{ symbol . }}"
	rendered, err = RenderTemplateWithData(template, "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee")
	assert.NoError(t, err)
	assert.Equal(t, "ETH", rendered)

	template = "{{ decimals . }}"
	rendered, err = RenderTemplateWithData(template, "0x6b175474e89094c44da98b954eedeac495271d0f")
	assert.NoError(t, err)
	assert.Equal(t, "18", rendered)

	template = "{{ decimals . }}"
	rendered, err = RenderTemplateWithData(template, "0x0000000000000000000000000000000000000000")
	assert.NoError(t, err)
	assert.Equal(t, "18", rendered)

	template = "{{ decimals . }}"
	assert.NoError(t, err)
	rendered, err = RenderTemplateWithData(template, "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee")
	assert.Equal(t, "18", rendered)

	template = `{{ balanceOf "0x9f8f72aa9304c8b593d555f12ef6589cc3a579a2" "0x6b175474e89094c44da98b954eedeac495271d0f" }}`
	rendered, err = RenderTemplateWithData(template, nil)
	assert.NoError(t, err)
	assert.Equal(t, "100000000000000", rendered)
}

func TestConversionRates(t *testing.T) {
	template := `this should not {{ toFiat "0x" "usd"}} completely break the parsing`
	rendered, err := RenderTemplateWithData(template, nil)
	assert.NoError(t, err) // we've hidden the error
	assert.Equal(t, "this should not 0 completely break the parsing", rendered)
}

func TestERC20Snapshot(t *testing.T) {

	template := `{{ ERC20Snapshot . }}`
	data := []interface{}{[]string{"100", "0", "99"}}

	rendered, err := RenderTemplateWithData(template, data)
	assert.NoError(t, err)
	assert.Equal(t, "map[0x0000000000000000000000000000000000000000:100 0x0000000000004946c0e9f43f4dee607b0ef1fa1c:99]", rendered)

	template = `{{ index (ERC20Snapshot .) "0x0000000000000000000000000000000000000000" }}`
	rendered, err = RenderTemplateWithData(template, data)
	assert.NoError(t, err)
	assert.Equal(t, "100", rendered)
}

func TestEthCall(t *testing.T) {
	blockNo, err := tokenapi.GetTokenAPI().GetRPCCli().EthBlockNumber()
	assert.NoError(t, err)

	template := `{{ ethCall "0x1f9840a85d5af5bf1d1762f925bdaddc4201f984" . 0 "balanceOf" "0x41ac4e73e8dE10E9A902785989Fbc28E7cdc5abC" }}`
	rendered, err := RenderTemplateWithData(template, blockNo)
	assert.NoError(t, err)
	assert.Equal(t, "300000000000000000000", rendered)

	//template = `{{ ethCall "0x1f9840a85d5af5bf1d1762f925bdaddc4201f984" . 0 "name" }}`
	//rendered, err = RenderTemplateWithData(template, blockNo)
	//assert.NoError(t, err)
	//assert.Equal(t, "Uniswap", rendered)
}

func setupGock(filename, url, path string) error {
	testJSON, err := os.Open(filename)
	defer testJSON.Close()
	if err != nil {
		return err
	}
	testByte, err := ioutil.ReadAll(testJSON)
	if err != nil {
		return err
	}
	gock.New(url).
		Get(path).
		Reply(200).
		JSON(testByte)

	return nil
}
