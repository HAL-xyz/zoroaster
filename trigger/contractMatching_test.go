package trigger

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"zoroaster/config"
)

func TestMatchContract1(t *testing.T) {

	// () -> address
	tg, err := GetTriggerFromFile("../resources/triggers/wac1.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliTest, tg, 8387102)

	assert.True(t, isMatch)
	assert.Equal(t, "0x4a574510c7014e4ae985403536074abe582adfc8", matchingVals[0])
	assert.Equal(t, []interface{}{"0x4a574510c7014e4ae985403536074abe582adfc8"}, returnedVals)
}

func TestMatchContract2(t *testing.T) {

	// address -> uint256
	tg, err := GetTriggerFromFile("../resources/triggers/wac2.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliTest, tg, 8387679)

	assert.True(t, isMatch)
	assert.Equal(t, "3876846319093283908984", matchingVals[0])
	assert.Equal(t, []interface{}{"3876846319093283908984"}, returnedVals)
}

func TestMatchContract3(t *testing.T) {

	// () -> bool
	tg, err := GetTriggerFromFile("../resources/triggers/wac3.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliTest, tg, 8387102)

	assert.True(t, isMatch)
	assert.Equal(t, "true", matchingVals[0])
	assert.Equal(t, []interface{}{"true"}, returnedVals)
}

func TestMatchContract4(t *testing.T) {

	// uint256 -> address
	tg, err := GetTriggerFromFile("../resources/triggers/wac4.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliTest, tg, 8387102)

	assert.True(t, isMatch)
	assert.Equal(t, "0xd4fe7bc31cedb7bfb8a345f31e668033056b2728", matchingVals[0])
	assert.Equal(t, []interface{}{"0xd4fe7bc31cedb7bfb8a345f31e668033056b2728"}, returnedVals)
}

func TestMatchContract5(t *testing.T) {

	// uint16 -> address
	tg, err := GetTriggerFromFile("../resources/triggers/wac5.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliTest, tg, 8387102)

	assert.True(t, isMatch)
	assert.Equal(t, "0x02ca0dfabf5285b0b9d09dfaa241167013355c35", matchingVals[0])
	assert.Equal(t, []interface{}{"0x02ca0dfabf5285b0b9d09dfaa241167013355c35"}, returnedVals)
}

func TestMatchContract6(t *testing.T) {

	// () -> uint256[3]
	tg, err := GetTriggerFromFile("../resources/triggers/wac6.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliRinkeby, tg, 4974958)

	assert.True(t, isMatch)
	assert.Equal(t, "12", matchingVals[0])
	assert.Equal(t, []interface{}{[]string{"4", "8", "12"}}, returnedVals)
}

func TestMatchContract7(t *testing.T) {

	// () -> (int128, int128, int128)
	tg, err := GetTriggerFromFile("../resources/triggers/wac7.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliRinkeby, tg, 4974958)

	assert.True(t, isMatch)
	assert.Equal(t, "4", matchingVals[0])
	assert.Equal(t, []interface{}{"4", "8", "12"}, returnedVals)
}

func TestMatchContract8(t *testing.T) {

	// () -> (int128, string, string)
	tg, err := GetTriggerFromFile("../resources/triggers/wac8.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliRinkeby, tg, 4974958)

	assert.True(t, isMatch)
	assert.Equal(t, "moon", matchingVals[0])
	assert.Equal(t, []interface{}{"4", "sailor", "moon"}, returnedVals)
}

func TestMatchContract9(t *testing.T) {

	// () -> string[3]
	tg, err := GetTriggerFromFile("../resources/triggers/wac9.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliRinkeby, tg, 4974958)

	assert.True(t, isMatch)
	assert.Equal(t, "ciao", matchingVals[0])
	assert.Equal(t, []interface{}{[]string{"ciao", "come", "stai"}}, returnedVals)
}

func TestMatchContract10(t *testing.T) {

	// () -> string[3]
	tg, err := GetTriggerFromFile("../resources/triggers/wac10.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliRinkeby, tg, 4974958)

	exp := []interface{}{
		[]string{
			"0x4fed1fc4144c223ae3c1553be203cdfcbd38c581",
			"0x65d21616594825a738bcd08a5227358593a9aaf2",
			"0xd76f7d7d2ede0631ad23e4a01176c0e59878abda",
		}}

	assert.True(t, isMatch)
	assert.Equal(t, len(matchingVals), 2)
	assert.Equal(t, "[4fed1fc4144c223ae3c1553be203cdfcbd38c581 65d21616594825a738bcd08a5227358593a9aaf2 d76f7d7d2ede0631ad23e4a01176c0e59878abda]", matchingVals[0])
	assert.Equal(t, "0x4fed1fc4144c223ae3c1553be203cdfcbd38c581", matchingVals[1])
	assert.Equal(t, exp, returnedVals)
}

func TestMatchContract11(t *testing.T) {

	// () -> string[3]
	tg, err := GetTriggerFromFile("../resources/triggers/wac11.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliRinkeby, tg, 4974958)

	exp := []interface{}{
		[]string{
			"0x4fed1fc4144c223ae3c1553be203cdfcbd38c581",
			"0x65d21616594825a738bcd08a5227358593a9aaf2",
			"0xd76f7d7d2ede0631ad23e4a01176c0e59878abda",
		}}

	assert.False(t, isMatch)              // no match
	assert.Equal(t, 1, len(matchingVals)) // only the first Output matches
	assert.Equal(t, "[4fed1fc4144c223ae3c1553be203cdfcbd38c581 65d21616594825a738bcd08a5227358593a9aaf2 d76f7d7d2ede0631ad23e4a01176c0e59878abda]", matchingVals[0])
	assert.Equal(t, exp, returnedVals)
}

func TestMatchContract12(t *testing.T) {

	// int8 -> string
	tg, err := GetTriggerFromFile("../resources/triggers/wac12.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliRinkeby, tg, 4974958)

	assert.True(t, isMatch)
	assert.Equal(t, "99", matchingVals[0])
	assert.Equal(t, []interface{}{"99"}, returnedVals)
}

func TestMatchContract13(t *testing.T) {

	// int8[3] -> string
	tg, err := GetTriggerFromFile("../resources/triggers/wac13.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliRinkeby, tg, 4974958)

	assert.True(t, isMatch)
	assert.Equal(t, "20", matchingVals[0])
	assert.Equal(t, []interface{}{"20"}, returnedVals)
}

func TestMatchContract14(t *testing.T) {

	// int8[] -> string
	tg, err := GetTriggerFromFile("../resources/triggers/wac14.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliRinkeby, tg, 4974958)

	assert.True(t, isMatch)
	assert.Equal(t, "10", matchingVals[0])
	assert.Equal(t, []interface{}{"10"}, returnedVals)
}

func TestMatchContract15(t *testing.T) {

	// int8, int16[3], int32[] -> int256[3], bytes, int64
	tg, err := GetTriggerFromFile("../resources/triggers/wac15.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliRinkeby, tg, 5527743)

	exp := []interface{}{
		[]string{"20", "10", "333333333333333333333"},
		"6c61206c61206c612068656c6c6f20776f726c64",
		"110",
	}

	assert.True(t, isMatch)
	assert.Equal(t, "20", matchingVals[0])
	assert.Equal(t, "10", matchingVals[1])
	assert.Equal(t, "110", matchingVals[2])
	assert.Equal(t, "0x6c61206c61206c612068656c6c6f20776f726c64", matchingVals[3])
	assert.Equal(t, exp, returnedVals)

}

func TestMatchContract16(t *testing.T) {

	// address, address[3], address[] -> address, address[3], address[]
	tg, err := GetTriggerFromFile("../resources/triggers/wac16.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliRinkeby, tg, 5527743)

	exp := []interface{}{
		"0x6d64ca40f70cc4c54bbb2a32f1ea52a7d7d4ccec",
		[]string{"0x0000000000000000000000000000000000000000", "0x6d64ca40f70cc4c54bbb2a32f1ea52a7d7d4ccec", "0x6d64ca40f70cc4c54bbb2a32f1ea52a7d7d4ccec"},
		[]string{"0x0000000000000000000000000000000000000001", "0x6d64ca40f70cc4c54bbb2a32f1ea52a7d7d4ccec", "0x6d64ca40f70cc4c54bbb2a32f1ea52a7d7d4ccec"},
	}

	assert.True(t, isMatch)
	assert.Equal(t, "0x6d64ca40f70cc4c54bbb2a32f1ea52a7d7d4ccec", matchingVals[0])
	assert.Equal(t, "0x0000000000000000000000000000000000000000", matchingVals[1])
	assert.Equal(t, "0x0000000000000000000000000000000000000001", matchingVals[2])
	assert.Equal(t, exp, returnedVals)
}

func TestMatchContract17(t *testing.T) {

	// bytes -> bytes
	tg, err := GetTriggerFromFile("../resources/triggers/wac17.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliRinkeby, tg, 5527743)
	assert.True(t, isMatch)
	assert.Equal(t, "0x68656c6c6f20776f726c64", matchingVals[0])
	assert.Equal(t, 1, len(matchingVals))
	assert.Equal(t, []interface{}{"68656c6c6f20776f726c64"}, returnedVals)
}

func TestMatchContract18(t *testing.T) {

	// bytes32 -> bytes32
	tg, err := GetTriggerFromFile("../resources/triggers/wac18.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliRinkeby, tg, 5527743)
	assert.True(t, isMatch)
	assert.Equal(t, 1, len(matchingVals))
	assert.Equal(t, "0x68656c6c6f20776f726c64000000000000000000000000000000000000000000", matchingVals[0])
	assert.Equal(t, []interface{}{"68656c6c6f20776f726c64000000000000000000000000000000000000000000"}, returnedVals)
}

func TestMatchContract19(t *testing.T) {

	// byte16 -> byte16
	tg, err := GetTriggerFromFile("../resources/triggers/wac19.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliRinkeby, tg, 5527743)
	assert.True(t, isMatch)
	assert.Equal(t, 1, len(matchingVals))
	assert.Equal(t, "0x68656c6c6f20776f726c640000000000", matchingVals[0])
	assert.Equal(t, []interface{}{"68656c6c6f20776f726c640000000000"}, returnedVals)
}

func TestMatchContractUniswap(t *testing.T) {

	tg, err := GetTriggerFromFile("../resources/triggers/wac-uniswap.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(config.CliTest, tg, 8496486)
	assert.True(t, isMatch)
	assert.Equal(t, "177", matchingVals[0])
	assert.Equal(t, []interface{}{"177"}, returnedVals)
}
