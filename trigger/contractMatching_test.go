package trigger

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/onrik/ethrpc"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"zoroaster/config"
	"zoroaster/utils"
)

var zconf = config.Load("../config")
var client = ethrpc.New(zconf.EthNode)
var cliRinkeby = ethrpc.New("https://rinkebyshared.bdnodes.net?auth=dKvc9d7tXrOdmnKK9nsfl119I19PH4GZPbACnbH-QW0")

func TestMatchContract1(t *testing.T) {

	// () -> address
	tg, err := GetTriggerFromFile("../resources/triggers/wac1.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(client, tg, 8387102)
	exp := []interface{}{common.HexToAddress("0x4a574510c7014e4ae985403536074abe582adfc8")}

	assert.True(t, isMatch)
	assert.Equal(t, "0x4a574510c7014e4ae985403536074abe582adfc8", matchingVals[0])
	assert.Equal(t, exp, returnedVals)

}

func TestMatchContract2(t *testing.T) {

	// address -> uint256
	tg, err := GetTriggerFromFile("../resources/triggers/wac2.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(client, tg, 8387679)
	exp := []interface{}{utils.MakeBigInt("3876846319093283908984")}

	assert.True(t, isMatch)
	assert.Equal(t, "3876846319093283908984", matchingVals[0])
	assert.Equal(t, exp, returnedVals)
}

func TestMatchContract3(t *testing.T) {

	// () -> bool
	tg, err := GetTriggerFromFile("../resources/triggers/wac3.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(client, tg, 8387102)
	exp := []interface{}{true}

	assert.True(t, isMatch)
	assert.Equal(t, "true", matchingVals[0])
	assert.Equal(t, exp, returnedVals)
}

func TestMatchContract4(t *testing.T) {

	// uint256 -> address
	tg, err := GetTriggerFromFile("../resources/triggers/wac4.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(client, tg, 8387102)
	exp := []interface{}{common.HexToAddress("0xd4fe7bc31cedb7bfb8a345f31e668033056b2728")}

	assert.True(t, isMatch)
	assert.Equal(t, "0xd4fe7bc31cedb7bfb8a345f31e668033056b2728", matchingVals[0])
	assert.Equal(t, exp, returnedVals)
}

func TestMatchContract5(t *testing.T) {

	// uint16 -> address
	tg, err := GetTriggerFromFile("../resources/triggers/wac5.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(client, tg, 8387102)
	assert.True(t, isMatch)
	exp := []interface{}{common.HexToAddress("0x02ca0dfabf5285b0b9d09dfaa241167013355c35")}
	assert.Equal(t, "0x02ca0dfabf5285b0b9d09dfaa241167013355c35", matchingVals[0])
	assert.Equal(t, exp, returnedVals)
}

func TestMatchContract6(t *testing.T) {

	// () -> uint256[3]
	tg, err := GetTriggerFromFile("../resources/triggers/wac6.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)
	exp := []interface{}{
		[3]*big.Int{big.NewInt(4), big.NewInt(8), big.NewInt(12)}}

	assert.True(t, isMatch)
	assert.Equal(t, "12", matchingVals[0])
	assert.Equal(t, exp, returnedVals)
}

func TestMatchContract7(t *testing.T) {

	// () -> (int128, int128, int128)
	tg, err := GetTriggerFromFile("../resources/triggers/wac7.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)
	exp := []interface{}{big.NewInt(4), big.NewInt(8), big.NewInt(12)}

	assert.True(t, isMatch)
	assert.Equal(t, "4", matchingVals[0])
	assert.Equal(t, exp, returnedVals)
	assert.Equal(t, returnedVals[0], big.NewInt(4))
	assert.Equal(t, returnedVals[1], big.NewInt(8))
	assert.Equal(t, returnedVals[2], big.NewInt(12))
}

func TestMatchContract8(t *testing.T) {

	// () -> (int128, string, string)
	tg, err := GetTriggerFromFile("../resources/triggers/wac8.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)
	exp := []interface{}{
		big.NewInt(4), "sailor", "moon"}

	assert.True(t, isMatch)
	assert.Equal(t, "moon", matchingVals[0])
	assert.Equal(t, exp, returnedVals)
}

func TestMatchContract9(t *testing.T) {

	// () -> string[3]
	tg, err := GetTriggerFromFile("../resources/triggers/wac9.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)
	exp := []interface{}{[3]string{"ciao", "come", "stai"}}
	assert.True(t, isMatch)
	assert.Equal(t, "ciao", matchingVals[0])
	assert.Equal(t, exp, returnedVals)
}

func TestMatchContract10(t *testing.T) {

	// () -> string[3]
	tg, err := GetTriggerFromFile("../resources/triggers/wac10.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)
	exp := []interface{}{
		[3]common.Address{
			common.HexToAddress("0x4fed1fc4144c223ae3c1553be203cdfcbd38c581"),
			common.HexToAddress("0x65d21616594825a738bcd08a5227358593a9aaf2"),
			common.HexToAddress("0xd76f7d7d2ede0631ad23e4a01176c0e59878abda"),
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

	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)
	exp := []interface{}{
		[3]common.Address{
			common.HexToAddress("0x4fed1fc4144c223ae3c1553be203cdfcbd38c581"),
			common.HexToAddress("0x65d21616594825a738bcd08a5227358593a9aaf2"),
			common.HexToAddress("0xd76f7d7d2ede0631ad23e4a01176c0e59878abda"),
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

	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)

	exp := []interface{}{"99"}
	assert.True(t, isMatch)
	assert.Equal(t, "99", matchingVals[0])
	assert.Equal(t, exp, returnedVals)
}

func TestMatchContract13(t *testing.T) {

	// int8[3] -> string
	tg, err := GetTriggerFromFile("../resources/triggers/wac13.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)

	exp := []interface{}{"20"}
	assert.True(t, isMatch)
	assert.Equal(t, "20", matchingVals[0])
	assert.Equal(t, exp, returnedVals)
}

func TestMatchContract14(t *testing.T) {

	// int8[] -> string
	tg, err := GetTriggerFromFile("../resources/triggers/wac14.json")
	assert.NoError(t, err)

	isMatch, matchingVals, returnedVals := MatchContract(cliRinkeby, tg, 4974958)

	exp := []interface{}{"10"}
	assert.True(t, isMatch)
	assert.Equal(t, "10", matchingVals[0])
	assert.Equal(t, exp, returnedVals)
}

func TestMatchContract15(t *testing.T) {

	// int8, int16[3], int32[] -> int256[3], bytes, int64
	tg, err := GetTriggerFromFile("../resources/triggers/wac15.json")
	assert.NoError(t, err)

	isMatch, matchingVals, _ := MatchContract(cliRinkeby, tg, 5527743)
	assert.True(t, isMatch)
	assert.Equal(t, "20", matchingVals[0])
	assert.Equal(t, "10", matchingVals[1])
	assert.Equal(t, "110", matchingVals[2])
	assert.Equal(t, "0x6c61206c61206c612068656c6c6f20776f726c64", matchingVals[3])
}

func TestMatchContract16(t *testing.T) {

	// address, address[3], address[] -> address, address[3], address[]
	tg, err := GetTriggerFromFile("../resources/triggers/wac16.json")
	assert.NoError(t, err)

	isMatch, matchingVals, _ := MatchContract(cliRinkeby, tg, 5527743)
	assert.True(t, isMatch)
	assert.Equal(t, "0x6d64ca40f70cc4c54bbb2a32f1ea52a7d7d4ccec", matchingVals[0])
	assert.Equal(t, "0x0000000000000000000000000000000000000000", matchingVals[1])
	assert.Equal(t, "0x0000000000000000000000000000000000000001", matchingVals[2])
}

func TestMatchContract17(t *testing.T) {

	// bytes -> bytes
	tg, err := GetTriggerFromFile("../resources/triggers/wac17.json")
	assert.NoError(t, err)

	isMatch, matchingVals, _ := MatchContract(cliRinkeby, tg, 5527743)
	assert.True(t, isMatch)
	assert.Equal(t, "0x68656c6c6f20776f726c64", matchingVals[0])
	assert.Equal(t, 1, len(matchingVals))
}

func TestMatchContract18(t *testing.T) {

	// bytes32 -> bytes32
	tg, err := GetTriggerFromFile("../resources/triggers/wac18.json")
	assert.NoError(t, err)

	isMatch, matchingVals, _ := MatchContract(cliRinkeby, tg, 5527743)
	assert.True(t, isMatch)
	assert.Equal(t, 1, len(matchingVals))
	assert.Equal(t, "0x68656c6c6f20776f726c64000000000000000000000000000000000000000000", matchingVals[0])
}

func TestMatchContract19(t *testing.T) {

	// byte16 -> byte16
	tg, err := GetTriggerFromFile("../resources/triggers/wac19.json")
	assert.NoError(t, err)

	isMatch, matchingVals, _ := MatchContract(cliRinkeby, tg, 5527743)
	assert.True(t, isMatch)
	assert.Equal(t, 1, len(matchingVals))
	assert.Equal(t, "0x68656c6c6f20776f726c640000000000", matchingVals[0])
}

func TestMatchContractUniswap(t *testing.T) {

	tg, err := GetTriggerFromFile("../resources/triggers/wac-uniswap.json")
	assert.NoError(t, err)

	isMatch, matchingVals, _ := MatchContract(client, tg, 8496486)
	assert.True(t, isMatch)
	assert.Equal(t, "177", matchingVals[0])
}
