package action

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"zoroaster/trigger"
)

func TestTxMatching(t *testing.T) {

	// TODO big int json unmarshaling is broken

	block, _ := trigger.GetBlockFromFile("../resources/blocks/block1.json")
	tg, _ := trigger.GetTriggerFromFile("../resources/triggers/t2.json")

	matches := trigger.MatchTransaction(tg, block)

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
Contract parameter "_value" is: 4e&#43;18
Transaction from is: 0x3d2339bf362a9b0f8ef3ca0867bd73f350ed66ac
Transaction gas is: 115960
Transaction gas price is: 7000000000
Transaction nonce is: 414
Transaction to is: 0x174bfa6600bf90c885c7c01c7031389ed1461ab9
Transaction hash is: 0x42c8de77ef5d76f36aea6e051b9059ece6e34619d9fb4a1d97f3224d5c990a67
Transaction value is: 0
Transaction input data is: 0xa9059cbb000000000000000000000000fea2f9433058cd555fd67cdde8efd7e6031e56c00000000000000000000000000000000000000000000000003782dace9d900000
`
	rendered, err := renderTemplateWithData(templateText, matches[0].ToTemplateMatch())
	assert.NoError(t, err)
	assert.Equal(t, expectedOutcome, rendered)
}

func TestContractMatching(t *testing.T) {

	var cnMatch = trigger.CnMatch{
		Trigger:        tg,
		MatchUUID:      "uuid",
		BlockNumber:    88888,
		MatchedValues:  []string{},
		BlockTimestamp: 123456,
		AllValues:      []interface{}{},
		BlockHash:      "0x",
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

	rendered, err := renderTemplateWithData(templateText, cnMatch.ToTemplateMatch())
	assert.Equal(t, expectedOutcome, rendered)
	assert.Error(t, err) // error isn't nil because of the out of bound indexing
}

func TestEventMatching(t *testing.T) {

	tg1, err := trigger.GetTriggerFromFile("../resources/triggers/ev1.json")
	assert.NoError(t, err)
	matches := trigger.MatchEvent(EthMock{}, tg1, 8496661, 1572344236)

	matches[0].EventParams["arrayParam"] = []string{"hello", "world", "yo yo"}

	templateText := `
Block Number is: {{ .Block.Number }}
Event name is {{ .Contract.EventName }}
Event from param is: {{ index .Contract.EventParameters "to" }}
Event value param is: {{ index .Contract.EventParameters "value" }}
First element in array parameter is: {{ index (index .Contract.EventParameters "arrayParam") 0 }}
Missing param is: {{ index .Contract.EventParameters "missing" }}
`
	expectedOutcome := `
Block Number is: 8496661
Event name is Transfer
Event from param is: 0xcd95b32c98423172e04b1c76841e5a73f4532a7f
Event value param is: 677420000
First element in array parameter is: hello
Missing param is: 
`
	rendered, err := renderTemplateWithData(templateText, matches[0].ToTemplateMatch())
	assert.NoError(t, err)
	assert.Equal(t, expectedOutcome, rendered)
}

func TestTemplateFunctions(t *testing.T) {

	template := "{{ hexToASCII . }}"
	rendered, err := renderTemplateWithData(template, "0x4920686176652031303021")
	assert.NoError(t, err)
	assert.Equal(t, "I have 100!", rendered)

	template = "{{ hexToInt . }}"
	rendered, err = renderTemplateWithData(template, "0xEA")
	assert.NoError(t, err)
	assert.Equal(t, "234", rendered)

	template = "{{ etherscanTxLink . }}"
	rendered, err = renderTemplateWithData(template, "0xfdb96f7387559ebfc41e88e21962414eb527484f578ce87996f8733352ab2ee7")
	assert.NoError(t, err)
	assert.Equal(t, "https://etherscan.io/tx/0xfdb96f7387559ebfc41e88e21962414eb527484f578ce87996f8733352ab2ee7", rendered)

	template = "{{ etherscanAddressLink . }}"
	rendered, err = renderTemplateWithData(template, "0xfdb96f7387559ebfc41e88e21962414eb527484f578ce87996f8733352ab2ee7")
	assert.NoError(t, err)
	assert.Equal(t, "https://etherscan.io/address/0xfdb96f7387559ebfc41e88e21962414eb527484f578ce87996f8733352ab2ee7", rendered)

	template = "{{ etherscanTokenLink . }}"
	rendered, err = renderTemplateWithData(template, "0xfdb96f7387559ebfc41e88e21962414eb527484f578ce87996f8733352ab2ee7")
	assert.NoError(t, err)
	assert.Equal(t, "https://etherscan.io/token/0xfdb96f7387559ebfc41e88e21962414eb527484f578ce87996f8733352ab2ee7", rendered)

	template = "{{ fromWei . 18 }}"
	rendered, err = renderTemplateWithData(template, "629000000000000000")
	assert.NoError(t, err)
	assert.Equal(t, "0.63", rendered)

	template = "{{ fromWei . 6 }}"
	rendered, err = renderTemplateWithData(template, "629000000000000000")
	assert.NoError(t, err)
	assert.Equal(t, "629000000000", rendered)

	template = "{{ humanTime . }}"
	rendered, err = renderTemplateWithData(template, "1602631929")
	assert.NoError(t, err)
	assert.Equal(t, "14 Oct 20 00:32 BST", rendered)
}
