package trigger

import (
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

var lastBlockRinkeby int
var lastBlockMainnet int

func init() {
	var err error
	lastBlockRinkeby, err = config.CliRinkeby.EthBlockNumber()
	if err != nil {
		logrus.Fatal(err)
	}
	lastBlockMainnet, err = config.CliMain.EthBlockNumber()
	if err != nil {
		logrus.Fatal(err)
	}
}

func TestMatchContract1(t *testing.T) {

	// () -> address
	tg, err := GetTriggerFromFile("../resources/triggers/wac1.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliMain, tg, lastBlockMainnet)

	assert.NotNil(t, match)
	assert.Equal(t, "0x4a574510c7014e4ae985403536074abe582adfc8", match.MatchedValues[0])
	assert.Equal(t, []interface{}{"0x4a574510c7014e4ae985403536074abe582adfc8"}, match.AllValues)
}

func TestMatchContract2(t *testing.T) {

	// address -> uint256
	tg, err := GetTriggerFromFile("../resources/triggers/wac2.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliMain, tg, lastBlockMainnet)

	assert.NotNil(t, match)
	assert.Equal(t, "3876846319093283908984", match.MatchedValues[0])
	assert.Equal(t, []interface{}{"3876846319093283908984"}, match.AllValues)
}

func TestMatchContract3(t *testing.T) {

	// () -> bool
	tg, err := GetTriggerFromFile("../resources/triggers/wac3.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliMain, tg, lastBlockMainnet)

	assert.NotNil(t, match)
	assert.Equal(t, "true", match.MatchedValues[0])
	assert.Equal(t, []interface{}{"true"}, match.AllValues)
}

func TestMatchContract4(t *testing.T) {

	// uint256 -> address
	tg, err := GetTriggerFromFile("../resources/triggers/wac4.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliMain, tg, lastBlockMainnet)

	assert.NotNil(t, match)
	assert.Equal(t, "0xd4fe7bc31cedb7bfb8a345f31e668033056b2728", match.MatchedValues[0])
	assert.Equal(t, []interface{}{"0xd4fe7bc31cedb7bfb8a345f31e668033056b2728"}, match.AllValues)
}

func TestMatchContract5(t *testing.T) {

	// uint16 -> address
	tg, err := GetTriggerFromFile("../resources/triggers/wac5.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliMain, tg, lastBlockMainnet)

	assert.NotNil(t, match)
	assert.Equal(t, "0x02ca0dfabf5285b0b9d09dfaa241167013355c35", match.MatchedValues[0])
	assert.Equal(t, []interface{}{"0x02ca0dfabf5285b0b9d09dfaa241167013355c35"}, match.AllValues)
}

func TestMatchContract6(t *testing.T) {

	// () -> uint256[3]
	tg, err := GetTriggerFromFile("../resources/triggers/wac6.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliRinkeby, tg, lastBlockRinkeby)

	assert.NotNil(t, match)
	assert.Equal(t, "12", match.MatchedValues[0])
	assert.Equal(t, []interface{}{[]string{"4", "8", "12"}}, match.AllValues)
}

func TestMatchContract7(t *testing.T) {

	// () -> (int128, int128, int128)
	tg, err := GetTriggerFromFile("../resources/triggers/wac7.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliRinkeby, tg, lastBlockRinkeby)

	assert.NotNil(t, match)
	assert.Equal(t, "4", match.MatchedValues[0])
	assert.Equal(t, []interface{}{"4", "8", "12"}, match.AllValues)
}

func TestMatchContract8(t *testing.T) {

	// () -> (int128, string, string)
	tg, err := GetTriggerFromFile("../resources/triggers/wac8.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliRinkeby, tg, lastBlockRinkeby)

	assert.NotNil(t, match)
	assert.Equal(t, "moon", match.MatchedValues[0])
	assert.Equal(t, []interface{}{"4", "sailor", "moon"}, match.AllValues)
}

func TestMatchContract9(t *testing.T) {

	// () -> string[3]
	tg, err := GetTriggerFromFile("../resources/triggers/wac9.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliRinkeby, tg, lastBlockRinkeby)

	assert.NotNil(t, match)
	assert.Equal(t, "ciao", match.MatchedValues[0])
	assert.Equal(t, []interface{}{[]string{"ciao", "come", "stai"}}, match.AllValues)
}

func TestMatchContract10(t *testing.T) {

	// () -> string[3]
	tg, err := GetTriggerFromFile("../resources/triggers/wac10.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliRinkeby, tg, lastBlockRinkeby)

	exp := []interface{}{
		[]string{
			"0x4fed1fc4144c223ae3c1553be203cdfcbd38c581",
			"0x65d21616594825a738bcd08a5227358593a9aaf2",
			"0xd76f7d7d2ede0631ad23e4a01176c0e59878abda",
		}}

	assert.NotNil(t, match)
	assert.Equal(t, len(match.MatchedValues), 2)
	assert.Equal(t, "[4fed1fc4144c223ae3c1553be203cdfcbd38c581 65d21616594825a738bcd08a5227358593a9aaf2 d76f7d7d2ede0631ad23e4a01176c0e59878abda]", match.MatchedValues[0])
	assert.Equal(t, "0x4fed1fc4144c223ae3c1553be203cdfcbd38c581", match.MatchedValues[1])
	assert.Equal(t, exp, match.AllValues)
}

func TestMatchContract11(t *testing.T) {

	// () -> string[3]
	tg, err := GetTriggerFromFile("../resources/triggers/wac11.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliRinkeby, tg, lastBlockRinkeby)

	assert.Nil(t, match) // no match
}

func TestMatchContract12(t *testing.T) {

	// int8 -> string
	tg, err := GetTriggerFromFile("../resources/triggers/wac12.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliRinkeby, tg, lastBlockRinkeby)

	assert.NotNil(t, match)
	assert.Equal(t, "99", match.MatchedValues[0])
	assert.Equal(t, []interface{}{"99"}, match.AllValues)
}

func TestMatchContract13(t *testing.T) {

	// int8[3] -> string
	tg, err := GetTriggerFromFile("../resources/triggers/wac13.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliRinkeby, tg, lastBlockRinkeby)

	assert.NotNil(t, match)
	assert.Equal(t, "20", match.MatchedValues[0])
	assert.Equal(t, []interface{}{"20"}, match.AllValues)
}

func TestMatchContract14(t *testing.T) {

	// int8[] -> string
	tg, err := GetTriggerFromFile("../resources/triggers/wac14.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliRinkeby, tg, lastBlockRinkeby)

	assert.NotNil(t, match)
	assert.Equal(t, "10", match.MatchedValues[0])
	assert.Equal(t, []interface{}{"10"}, match.AllValues)
}

func TestMatchContract15(t *testing.T) {

	// int8, int16[3], int32[] -> int256[3], bytes, int64
	tg, err := GetTriggerFromFile("../resources/triggers/wac15.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliRinkeby, tg, lastBlockRinkeby)

	exp := []interface{}{
		[]string{"20", "10", "333333333333333333333"},
		"6c61206c61206c612068656c6c6f20776f726c64",
		"110",
	}

	assert.NotNil(t, match)
	assert.Equal(t, "20", match.MatchedValues[0])
	assert.Equal(t, "10", match.MatchedValues[1])
	assert.Equal(t, "110", match.MatchedValues[2])
	assert.Equal(t, "0x6c61206c61206c612068656c6c6f20776f726c64", match.MatchedValues[3])
	assert.Equal(t, exp, match.AllValues)

}

func TestMatchContract16(t *testing.T) {

	// address, address[3], address[] -> address, address[3], address[]
	tg, err := GetTriggerFromFile("../resources/triggers/wac16.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliRinkeby, tg, lastBlockRinkeby)

	exp := []interface{}{
		"0x6d64ca40f70cc4c54bbb2a32f1ea52a7d7d4ccec",
		[]string{"0x0000000000000000000000000000000000000000", "0x6d64ca40f70cc4c54bbb2a32f1ea52a7d7d4ccec", "0x6d64ca40f70cc4c54bbb2a32f1ea52a7d7d4ccec"},
		[]string{"0x0000000000000000000000000000000000000001", "0x6d64ca40f70cc4c54bbb2a32f1ea52a7d7d4ccec", "0x6d64ca40f70cc4c54bbb2a32f1ea52a7d7d4ccec"},
	}

	assert.NotNil(t, match)
	assert.Equal(t, "0x6d64ca40f70cc4c54bbb2a32f1ea52a7d7d4ccec", match.MatchedValues[0])
	assert.Equal(t, "0x0000000000000000000000000000000000000000", match.MatchedValues[1])
	assert.Equal(t, "0x0000000000000000000000000000000000000001", match.MatchedValues[2])
	assert.Equal(t, exp, match.AllValues)
}

func TestMatchContract17(t *testing.T) {

	// bytes -> bytes
	tg, err := GetTriggerFromFile("../resources/triggers/wac17.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliRinkeby, tg, lastBlockRinkeby)
	assert.NotNil(t, match)
	assert.Equal(t, "0x68656c6c6f20776f726c64", match.MatchedValues[0])
	assert.Equal(t, 1, len(match.MatchedValues))
	assert.Equal(t, []interface{}{"68656c6c6f20776f726c64"}, match.AllValues)
}

func TestMatchContract18(t *testing.T) {

	// bytes32 -> bytes32
	tg, err := GetTriggerFromFile("../resources/triggers/wac18.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliRinkeby, tg, lastBlockRinkeby)
	assert.NotNil(t, match)
	assert.Equal(t, 1, len(match.MatchedValues))
	assert.Equal(t, "0x68656c6c6f20776f726c64000000000000000000000000000000000000000000", match.MatchedValues[0])
	assert.Equal(t, []interface{}{"68656c6c6f20776f726c64000000000000000000000000000000000000000000"}, match.AllValues)
}

func TestMatchContract19(t *testing.T) {

	// byte16 -> byte16
	tg, err := GetTriggerFromFile("../resources/triggers/wac19.json")
	assert.NoError(t, err)

	match, err := MatchContract(config.CliRinkeby, tg, lastBlockRinkeby)
	assert.NotNil(t, match)
	assert.Equal(t, 1, len(match.MatchedValues))
	assert.Equal(t, "0x68656c6c6f20776f726c640000000000", match.MatchedValues[0])
	assert.Equal(t, []interface{}{"68656c6c6f20776f726c640000000000"}, match.AllValues)
}
