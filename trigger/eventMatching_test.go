package trigger

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/onrik/ethrpc"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"zoroaster/config"
	"zoroaster/utils"
)

func TestValidateFilterLog(t *testing.T) {

	logs, err := GetLogsFromFile("../resources/events/logs1.json")
	if err != nil {
		t.Error(err)
	}
	tg, err := GetTriggerFromFile("../resources/triggers/ev1.json")
	if err != nil {
		t.Error(err)
	}
	abiObj, err := abi.JSON(strings.NewReader(tg.ContractABI))

	res, err := validateFilterLog(&logs[0], &tg.Filters[0], &abiObj, tg.Filters[0].EventName)
	assert.NoError(t, err)
	assert.True(t, res)

	res2, err := validateFilterLog(&logs[0], &tg.Filters[1], &abiObj, tg.Filters[0].EventName)
	assert.NoError(t, err)
	assert.True(t, res2)

	res3 := validateTriggerLog(&logs[0], tg, &abiObj, "Transfer")
	assert.True(t, res3)

	res4 := validateTriggerLog(&logs[1], tg, &abiObj, "Transfer")
	assert.False(t, res4)
}

type EthMock struct{}

func (cli EthMock) EthGetLogs(params ethrpc.FilterParams) ([]ethrpc.Log, error) {
	return GetLogsFromFile("../resources/events/logs1.json")
}

type EthMock2 struct{}

func (cli EthMock2) EthGetLogs(params ethrpc.FilterParams) ([]ethrpc.Log, error) {
	return GetLogsFromFile("../resources/events/logs2.json")
}

func TestMatchEvent7(t *testing.T) {
	js := `{
    "Filters": [
        {
            "Condition": {
                "Attribute": "0x875c04fcadcd0ae4b369679d6d8eefaf3080de016142bddfabdfe543b430ffac",
                "Predicate": "Eq"
            },
            "EventName": "address_event",
            "FilterType": "CheckEventParameter",
            "ParameterName": "v4",
            "ParameterType": "address[]"
        }
    ],
    "ContractABI": "[{\"constant\":false,\"inputs\":[],\"name\":\"bool_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"address_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"string_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"bytes_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"int_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"string\",\"name\":\"v1\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"string[]\",\"name\":\"v2\",\"type\":\"string[]\"},{\"indexed\":true,\"internalType\":\"string[3]\",\"name\":\"v3\",\"type\":\"string[3]\"}],\"name\":\"string_event\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address[]\",\"name\":\"v4\",\"type\":\"address[]\"},{\"indexed\":true,\"internalType\":\"address[3]\",\"name\":\"v5\",\"type\":\"address[3]\"}],\"name\":\"address_event\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"int256[]\",\"name\":\"v6\",\"type\":\"int256[]\"},{\"indexed\":true,\"internalType\":\"int256[3]\",\"name\":\"v7\",\"type\":\"int256[3]\"}],\"name\":\"int_event\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bool\",\"name\":\"v8\",\"type\":\"bool\"},{\"indexed\":true,\"internalType\":\"bool[]\",\"name\":\"v9\",\"type\":\"bool[]\"},{\"indexed\":true,\"internalType\":\"bool[3]\",\"name\":\"v10\",\"type\":\"bool[3]\"}],\"name\":\"bool_event\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes16\",\"name\":\"v11\",\"type\":\"bytes16\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"v12\",\"type\":\"bytes\"}],\"name\":\"bytes_event\",\"type\":\"event\"}]",
    "ContractAdd": "0x63cbf20c5e2a2a6599627fdce8b9f0cc3b782be1",
    "TriggerName": "test event",
    "TriggerType": "WatchEvents"
}`
	tg, err := NewTriggerFromJson(js)
	assert.NoError(t, err)

	matches := MatchEvent(config.CliRinkeby, tg, 5693736, 1572344236)

	assert.Equal(t, 1, len(matches))
	assert.Equal(t, 5693736, matches[0].Log.BlockNumber)
}

func TestMatchEvent6(t *testing.T) {
	js := `{
    "Filters": [
        {
            "Condition": {
                "Attribute": "true",
                "Predicate": "Eq"
            },
            "EventName": "bool_event",
            "FilterType": "CheckEventParameter",
            "ParameterName": "v8",
            "ParameterType": "bool"
        },
        {
            "Condition": {
                "Attribute": "0xa6eef7e35abe7026729641147f7915573c7e97b47efa546f5f6e3230263bcb49",
                "Predicate": "Eq"
            },
            "EventName": "int_event",
            "FilterType": "CheckEventParameter",
            "ParameterName": "v9",
            "ParameterType": "bool[]"
        }
    ],
    "ContractABI": "[{\"constant\":false,\"inputs\":[],\"name\":\"bool_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"address_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"string_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"bytes_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"int_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"string\",\"name\":\"v1\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"string[]\",\"name\":\"v2\",\"type\":\"string[]\"},{\"indexed\":true,\"internalType\":\"string[3]\",\"name\":\"v3\",\"type\":\"string[3]\"}],\"name\":\"string_event\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address[]\",\"name\":\"v4\",\"type\":\"address[]\"},{\"indexed\":true,\"internalType\":\"address[3]\",\"name\":\"v5\",\"type\":\"address[3]\"}],\"name\":\"address_event\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"int256[]\",\"name\":\"v6\",\"type\":\"int256[]\"},{\"indexed\":true,\"internalType\":\"int256[3]\",\"name\":\"v7\",\"type\":\"int256[3]\"}],\"name\":\"int_event\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bool\",\"name\":\"v8\",\"type\":\"bool\"},{\"indexed\":true,\"internalType\":\"bool[]\",\"name\":\"v9\",\"type\":\"bool[]\"},{\"indexed\":true,\"internalType\":\"bool[3]\",\"name\":\"v10\",\"type\":\"bool[3]\"}],\"name\":\"bool_event\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes16\",\"name\":\"v11\",\"type\":\"bytes16\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"v12\",\"type\":\"bytes\"}],\"name\":\"bytes_event\",\"type\":\"event\"}]",
    "ContractAdd": "0x63cbf20c5e2a2a6599627fdce8b9f0cc3b782be1",
    "TriggerName": "test event",
    "TriggerType": "WatchEvents"
}`
	tg, err := NewTriggerFromJson(js)
	assert.NoError(t, err)

	matches := MatchEvent(config.CliRinkeby, tg, 5693736, 1572344236)

	assert.Equal(t, 1, len(matches))
	assert.Equal(t, 5693736, matches[0].Log.BlockNumber)
}

func TestMatchEvent5(t *testing.T) {
	js := `{
    "Filters": [
        {
            "Condition": {
                "Attribute": "0x7250aa92a65150fcccca5852ea1a09d977f749a84405ea8fcbfbc6727ec4f515",
                "Predicate": "Eq"
            },
            "EventName": "int_event",
            "FilterType": "CheckEventParameter",
            "ParameterName": "v6",
            "ParameterType": "int[]"
        }
    ],
    "ContractABI": "[{\"constant\":false,\"inputs\":[],\"name\":\"bool_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"address_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"string_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"bytes_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"int_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"string\",\"name\":\"v1\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"string[]\",\"name\":\"v2\",\"type\":\"string[]\"},{\"indexed\":true,\"internalType\":\"string[3]\",\"name\":\"v3\",\"type\":\"string[3]\"}],\"name\":\"string_event\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address[]\",\"name\":\"v4\",\"type\":\"address[]\"},{\"indexed\":true,\"internalType\":\"address[3]\",\"name\":\"v5\",\"type\":\"address[3]\"}],\"name\":\"address_event\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"int256[]\",\"name\":\"v6\",\"type\":\"int256[]\"},{\"indexed\":true,\"internalType\":\"int256[3]\",\"name\":\"v7\",\"type\":\"int256[3]\"}],\"name\":\"int_event\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bool\",\"name\":\"v8\",\"type\":\"bool\"},{\"indexed\":true,\"internalType\":\"bool[]\",\"name\":\"v9\",\"type\":\"bool[]\"},{\"indexed\":true,\"internalType\":\"bool[3]\",\"name\":\"v10\",\"type\":\"bool[3]\"}],\"name\":\"bool_event\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes16\",\"name\":\"v11\",\"type\":\"bytes16\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"v12\",\"type\":\"bytes\"}],\"name\":\"bytes_event\",\"type\":\"event\"}]",
    "ContractAdd": "0x63cbf20c5e2a2a6599627fdce8b9f0cc3b782be1",
    "TriggerName": "test event",
    "TriggerType": "WatchEvents"
}`
	tg, err := NewTriggerFromJson(js)
	assert.NoError(t, err)

	matches := MatchEvent(config.CliRinkeby, tg, 5693738, 1572344236)

	assert.Equal(t, 1, len(matches))
	assert.Equal(t, 5693738, matches[0].Log.BlockNumber)
}

func TestMatchEvent4(t *testing.T) {
	js := `{
    "Filters": [
        {
            "Condition": {
                "Attribute": "0x9c22ff5f21f0b81b113e63f7db6da94fedef11b2119b4088b89664fb9a3cb658",
                "Predicate": "Eq"
            },
            "EventName": "string_event",
            "FilterType": "CheckEventParameter",
            "ParameterName": "v1",
            "ParameterType": "string"
        },
        {
            "Condition": {
                "Attribute": "0x4c314d3ec7fe572b5f0ca8d4231464e89602faea100b49f287ef5148bfb5b776",
                "Predicate": "Eq"
            },
            "EventName": "string_event",
            "FilterType": "CheckEventParameter",
            "ParameterName": "v2",
            "ParameterType": "string[]"
        },
        {
            "Condition": {
                "Attribute": "0xe7cc7564e647aae3b7253c8ab67ae03afc76f838e01ce364433dba9960e50afd",
                "Predicate": "Eq"
            },
            "EventName": "string_event",
            "FilterType": "CheckEventParameter",
            "ParameterName": "v3",
            "ParameterType": "string[3]"
        }
    ],
    "ContractABI": "[{\"constant\":false,\"inputs\":[],\"name\":\"bool_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"address_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"string_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"bytes_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"int_event_f\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"string\",\"name\":\"v1\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"string[]\",\"name\":\"v2\",\"type\":\"string[]\"},{\"indexed\":true,\"internalType\":\"string[3]\",\"name\":\"v3\",\"type\":\"string[3]\"}],\"name\":\"string_event\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address[]\",\"name\":\"v4\",\"type\":\"address[]\"},{\"indexed\":true,\"internalType\":\"address[3]\",\"name\":\"v5\",\"type\":\"address[3]\"}],\"name\":\"address_event\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"int256[]\",\"name\":\"v6\",\"type\":\"int256[]\"},{\"indexed\":true,\"internalType\":\"int256[3]\",\"name\":\"v7\",\"type\":\"int256[3]\"}],\"name\":\"int_event\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bool\",\"name\":\"v8\",\"type\":\"bool\"},{\"indexed\":true,\"internalType\":\"bool[]\",\"name\":\"v9\",\"type\":\"bool[]\"},{\"indexed\":true,\"internalType\":\"bool[3]\",\"name\":\"v10\",\"type\":\"bool[3]\"}],\"name\":\"bool_event\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes16\",\"name\":\"v11\",\"type\":\"bytes16\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"v12\",\"type\":\"bytes\"}],\"name\":\"bytes_event\",\"type\":\"event\"}]",
    "ContractAdd": "0x63cbf20c5e2a2a6599627fdce8b9f0cc3b782be1",
    "TriggerName": "test event",
    "TriggerType": "WatchEvents"
}`
	tg, err := NewTriggerFromJson(js)
	assert.NoError(t, err)

	matches := MatchEvent(config.CliRinkeby, tg, 5693738, 1572344236)

	assert.Equal(t, 1, len(matches))
	assert.Equal(t, 5693738, matches[0].Log.BlockNumber)
}

func TestMatchEvent3(t *testing.T) {
	js := `{
    "Filters": [
        {
            "Condition": {
                "Attribute": "0xa52e014b3f5cc48287c2d483a3e026c32cc76e6d",
                "Predicate": "Eq"
            },
            "FilterType": "BasicFilter",
            "ParameterName": "To"
        },
        {
            "Condition": {
                "Attribute": "335632",
                "Predicate": "Eq"
            },
            "EventName": "LogResult",
            "FilterType": "CheckEventParameter",
            "ParameterName": "ResultSerialNumber",
            "ParameterType": "uint256"
        },
        {
            "Condition": {
                "Attribute": "0x49c2381f46efd87fbc3e6662593bf4992a6e027ca569bddd35b3dce2c2f9ec23",
                "Predicate": "Eq"
            },
            "EventName": "LogResult",
            "FilterType": "CheckEventParameter",
            "ParameterName": "BetID",
            "ParameterType": "bytes32"
        },
        {
            "Condition": {
                "Attribute": "0x4236daa27a262fe6baf9bb43ade5e41f8f7498f9",
                "Predicate": "Eq"
            },
            "EventName": "LogResult",
            "FilterType": "CheckEventParameter",
            "ParameterName": "PlayerAddress",
            "ParameterType": "address"
        }
    ],
    "ContractABI": "[{\"constant\":false,\"inputs\":[{\"name\":\"newCallbackGasPrice\",\"type\":\"uint256\"}],\"name\":\"ownerSetCallbackGasPrice\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalWeiWon\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"maxProfitAsPercentOfHouse\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newHouseEdge\",\"type\":\"uint256\"}],\"name\":\"ownerSetHouseEdge\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"myid\",\"type\":\"bytes32\"},{\"name\":\"result\",\"type\":\"string\"}],\"name\":\"__callback\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"payoutsPaused\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newTreasury\",\"type\":\"address\"}],\"name\":\"ownerSetTreasury\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"myid\",\"type\":\"bytes32\"},{\"name\":\"result\",\"type\":\"string\"},{\"name\":\"proof\",\"type\":\"bytes\"}],\"name\":\"__callback\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"maxNumber\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addressToCheck\",\"type\":\"address\"}],\"name\":\"playerGetPendingTxByAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newContractBalanceInWei\",\"type\":\"uint256\"}],\"name\":\"ownerUpdateContractBalance\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"maxProfitDivisor\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newPayoutStatus\",\"type\":\"bool\"}],\"name\":\"ownerPausePayouts\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"ownerChangeOwner\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"minNumber\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newMaxProfitAsPercent\",\"type\":\"uint256\"}],\"name\":\"ownerSetMaxProfitAsPercentOfHouse\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"treasury\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalWeiWagered\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newMinimumBet\",\"type\":\"uint256\"}],\"name\":\"ownerSetMinBet\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newStatus\",\"type\":\"bool\"}],\"name\":\"ownerPauseGame\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"gasForOraclize\",\"outputs\":[{\"name\":\"\",\"type\":\"uint32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"sendTo\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ownerTransferEther\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"contractBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"minBet\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"playerWithdrawPendingTransactions\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"maxProfit\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalBets\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"randomQueryID\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"gamePaused\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"originalPlayerBetId\",\"type\":\"bytes32\"},{\"name\":\"sendTo\",\"type\":\"address\"},{\"name\":\"originalPlayerProfit\",\"type\":\"uint256\"},{\"name\":\"originalPlayerBetValue\",\"type\":\"uint256\"}],\"name\":\"ownerRefundPlayer\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newSafeGasToOraclize\",\"type\":\"uint32\"}],\"name\":\"ownerSetOraclizeSafeGas\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"ownerkill\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"houseEdge\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"rollUnder\",\"type\":\"uint256\"}],\"name\":\"playerRollDice\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"houseEdgeDivisor\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"maxPendingPayouts\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"BetID\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"PlayerAddress\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"RewardValue\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"ProfitValue\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"BetValue\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"PlayerNumber\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"RandomQueryID\",\"type\":\"uint256\"}],\"name\":\"LogBet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"ResultSerialNumber\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"BetID\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"PlayerAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"PlayerNumber\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"DiceResult\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"Value\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"Status\",\"type\":\"int256\"},{\"indexed\":false,\"name\":\"Proof\",\"type\":\"bytes\"}],\"name\":\"LogResult\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"BetID\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"PlayerAddress\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"RefundValue\",\"type\":\"uint256\"}],\"name\":\"LogRefund\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"SentToAddress\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"AmountTransferred\",\"type\":\"uint256\"}],\"name\":\"LogOwnerTransfer\",\"type\":\"event\"}]",
    "ContractAdd": "0xa52e014b3f5cc48287c2d483a3e026c32cc76e6d",
    "TriggerName": "WAE MATTEO",
    "TriggerType": "WatchEvents"
}`
	tg, err := NewTriggerFromJson(js)
	assert.NoError(t, err)

	matches := MatchEvent(EthMock2{}, tg, 9098826, 1572344236)

	assert.Equal(t, 1, len(matches))
	assert.Equal(t, 9098826, matches[0].Log.BlockNumber)
}

func TestMatchEvent2(t *testing.T) {
	js := `{
    "Filters": [
        {
            "Condition": {
                "Attribute": "0xd3a6fdd4408f5fd15623abbae9041025a337314d",
                "Predicate": "Eq"
            },
            "EventName": "Fill",
            "FilterType": "CheckEventParameter",
            "ParameterName": "makerAddress",
            "ParameterType": "address"
        },
        {
            "Condition": {
                "Attribute": "0x0000000000000000000000000d056bb17ad4df5593b93a1efc29cb35ba4aa38d",
                "Predicate": "Eq"
            },
            "EventName": "Fill",
            "FilterType": "CheckEventParameter",
            "ParameterName": "feeRecipientAddress",
            "ParameterType": "address"
        }
    ],
    "ContractABI": "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"filled\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"name\":\"makerAddress\",\"type\":\"address\"},{\"name\":\"takerAddress\",\"type\":\"address\"},{\"name\":\"feeRecipientAddress\",\"type\":\"address\"},{\"name\":\"senderAddress\",\"type\":\"address\"},{\"name\":\"makerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"makerFee\",\"type\":\"uint256\"},{\"name\":\"takerFee\",\"type\":\"uint256\"},{\"name\":\"expirationTimeSeconds\",\"type\":\"uint256\"},{\"name\":\"salt\",\"type\":\"uint256\"},{\"name\":\"makerAssetData\",\"type\":\"bytes\"},{\"name\":\"takerAssetData\",\"type\":\"bytes\"}],\"name\":\"orders\",\"type\":\"tuple[]\"},{\"name\":\"takerAssetFillAmounts\",\"type\":\"uint256[]\"},{\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"batchFillOrders\",\"outputs\":[{\"components\":[{\"name\":\"makerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"makerFeePaid\",\"type\":\"uint256\"},{\"name\":\"takerFeePaid\",\"type\":\"uint256\"}],\"name\":\"totalFillResults\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"cancelled\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"},{\"name\":\"signerAddress\",\"type\":\"address\"},{\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"preSign\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"name\":\"makerAddress\",\"type\":\"address\"},{\"name\":\"takerAddress\",\"type\":\"address\"},{\"name\":\"feeRecipientAddress\",\"type\":\"address\"},{\"name\":\"senderAddress\",\"type\":\"address\"},{\"name\":\"makerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"makerFee\",\"type\":\"uint256\"},{\"name\":\"takerFee\",\"type\":\"uint256\"},{\"name\":\"expirationTimeSeconds\",\"type\":\"uint256\"},{\"name\":\"salt\",\"type\":\"uint256\"},{\"name\":\"makerAssetData\",\"type\":\"bytes\"},{\"name\":\"takerAssetData\",\"type\":\"bytes\"}],\"name\":\"leftOrder\",\"type\":\"tuple\"},{\"components\":[{\"name\":\"makerAddress\",\"type\":\"address\"},{\"name\":\"takerAddress\",\"type\":\"address\"},{\"name\":\"feeRecipientAddress\",\"type\":\"address\"},{\"name\":\"senderAddress\",\"type\":\"address\"},{\"name\":\"makerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"makerFee\",\"type\":\"uint256\"},{\"name\":\"takerFee\",\"type\":\"uint256\"},{\"name\":\"expirationTimeSeconds\",\"type\":\"uint256\"},{\"name\":\"salt\",\"type\":\"uint256\"},{\"name\":\"makerAssetData\",\"type\":\"bytes\"},{\"name\":\"takerAssetData\",\"type\":\"bytes\"}],\"name\":\"rightOrder\",\"type\":\"tuple\"},{\"name\":\"leftSignature\",\"type\":\"bytes\"},{\"name\":\"rightSignature\",\"type\":\"bytes\"}],\"name\":\"matchOrders\",\"outputs\":[{\"components\":[{\"components\":[{\"name\":\"makerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"makerFeePaid\",\"type\":\"uint256\"},{\"name\":\"takerFeePaid\",\"type\":\"uint256\"}],\"name\":\"left\",\"type\":\"tuple\"},{\"components\":[{\"name\":\"makerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"makerFeePaid\",\"type\":\"uint256\"},{\"name\":\"takerFeePaid\",\"type\":\"uint256\"}],\"name\":\"right\",\"type\":\"tuple\"},{\"name\":\"leftMakerAssetSpreadAmount\",\"type\":\"uint256\"}],\"name\":\"matchedFillResults\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"name\":\"makerAddress\",\"type\":\"address\"},{\"name\":\"takerAddress\",\"type\":\"address\"},{\"name\":\"feeRecipientAddress\",\"type\":\"address\"},{\"name\":\"senderAddress\",\"type\":\"address\"},{\"name\":\"makerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"makerFee\",\"type\":\"uint256\"},{\"name\":\"takerFee\",\"type\":\"uint256\"},{\"name\":\"expirationTimeSeconds\",\"type\":\"uint256\"},{\"name\":\"salt\",\"type\":\"uint256\"},{\"name\":\"makerAssetData\",\"type\":\"bytes\"},{\"name\":\"takerAssetData\",\"type\":\"bytes\"}],\"name\":\"order\",\"type\":\"tuple\"},{\"name\":\"takerAssetFillAmount\",\"type\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"fillOrderNoThrow\",\"outputs\":[{\"components\":[{\"name\":\"makerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"makerFeePaid\",\"type\":\"uint256\"},{\"name\":\"takerFeePaid\",\"type\":\"uint256\"}],\"name\":\"fillResults\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes4\"}],\"name\":\"assetProxies\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"name\":\"makerAddress\",\"type\":\"address\"},{\"name\":\"takerAddress\",\"type\":\"address\"},{\"name\":\"feeRecipientAddress\",\"type\":\"address\"},{\"name\":\"senderAddress\",\"type\":\"address\"},{\"name\":\"makerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"makerFee\",\"type\":\"uint256\"},{\"name\":\"takerFee\",\"type\":\"uint256\"},{\"name\":\"expirationTimeSeconds\",\"type\":\"uint256\"},{\"name\":\"salt\",\"type\":\"uint256\"},{\"name\":\"makerAssetData\",\"type\":\"bytes\"},{\"name\":\"takerAssetData\",\"type\":\"bytes\"}],\"name\":\"orders\",\"type\":\"tuple[]\"}],\"name\":\"batchCancelOrders\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"name\":\"makerAddress\",\"type\":\"address\"},{\"name\":\"takerAddress\",\"type\":\"address\"},{\"name\":\"feeRecipientAddress\",\"type\":\"address\"},{\"name\":\"senderAddress\",\"type\":\"address\"},{\"name\":\"makerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"makerFee\",\"type\":\"uint256\"},{\"name\":\"takerFee\",\"type\":\"uint256\"},{\"name\":\"expirationTimeSeconds\",\"type\":\"uint256\"},{\"name\":\"salt\",\"type\":\"uint256\"},{\"name\":\"makerAssetData\",\"type\":\"bytes\"},{\"name\":\"takerAssetData\",\"type\":\"bytes\"}],\"name\":\"orders\",\"type\":\"tuple[]\"},{\"name\":\"takerAssetFillAmounts\",\"type\":\"uint256[]\"},{\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"batchFillOrKillOrders\",\"outputs\":[{\"components\":[{\"name\":\"makerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"makerFeePaid\",\"type\":\"uint256\"},{\"name\":\"takerFeePaid\",\"type\":\"uint256\"}],\"name\":\"totalFillResults\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"targetOrderEpoch\",\"type\":\"uint256\"}],\"name\":\"cancelOrdersUpTo\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"name\":\"makerAddress\",\"type\":\"address\"},{\"name\":\"takerAddress\",\"type\":\"address\"},{\"name\":\"feeRecipientAddress\",\"type\":\"address\"},{\"name\":\"senderAddress\",\"type\":\"address\"},{\"name\":\"makerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"makerFee\",\"type\":\"uint256\"},{\"name\":\"takerFee\",\"type\":\"uint256\"},{\"name\":\"expirationTimeSeconds\",\"type\":\"uint256\"},{\"name\":\"salt\",\"type\":\"uint256\"},{\"name\":\"makerAssetData\",\"type\":\"bytes\"},{\"name\":\"takerAssetData\",\"type\":\"bytes\"}],\"name\":\"orders\",\"type\":\"tuple[]\"},{\"name\":\"takerAssetFillAmounts\",\"type\":\"uint256[]\"},{\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"batchFillOrdersNoThrow\",\"outputs\":[{\"components\":[{\"name\":\"makerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"makerFeePaid\",\"type\":\"uint256\"},{\"name\":\"takerFeePaid\",\"type\":\"uint256\"}],\"name\":\"totalFillResults\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"assetProxyId\",\"type\":\"bytes4\"}],\"name\":\"getAssetProxy\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"transactions\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"name\":\"makerAddress\",\"type\":\"address\"},{\"name\":\"takerAddress\",\"type\":\"address\"},{\"name\":\"feeRecipientAddress\",\"type\":\"address\"},{\"name\":\"senderAddress\",\"type\":\"address\"},{\"name\":\"makerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"makerFee\",\"type\":\"uint256\"},{\"name\":\"takerFee\",\"type\":\"uint256\"},{\"name\":\"expirationTimeSeconds\",\"type\":\"uint256\"},{\"name\":\"salt\",\"type\":\"uint256\"},{\"name\":\"makerAssetData\",\"type\":\"bytes\"},{\"name\":\"takerAssetData\",\"type\":\"bytes\"}],\"name\":\"order\",\"type\":\"tuple\"},{\"name\":\"takerAssetFillAmount\",\"type\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"fillOrKillOrder\",\"outputs\":[{\"components\":[{\"name\":\"makerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"makerFeePaid\",\"type\":\"uint256\"},{\"name\":\"takerFeePaid\",\"type\":\"uint256\"}],\"name\":\"fillResults\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"validatorAddress\",\"type\":\"address\"},{\"name\":\"approval\",\"type\":\"bool\"}],\"name\":\"setSignatureValidatorApproval\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"address\"}],\"name\":\"allowedValidators\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"name\":\"makerAddress\",\"type\":\"address\"},{\"name\":\"takerAddress\",\"type\":\"address\"},{\"name\":\"feeRecipientAddress\",\"type\":\"address\"},{\"name\":\"senderAddress\",\"type\":\"address\"},{\"name\":\"makerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"makerFee\",\"type\":\"uint256\"},{\"name\":\"takerFee\",\"type\":\"uint256\"},{\"name\":\"expirationTimeSeconds\",\"type\":\"uint256\"},{\"name\":\"salt\",\"type\":\"uint256\"},{\"name\":\"makerAssetData\",\"type\":\"bytes\"},{\"name\":\"takerAssetData\",\"type\":\"bytes\"}],\"name\":\"orders\",\"type\":\"tuple[]\"},{\"name\":\"takerAssetFillAmount\",\"type\":\"uint256\"},{\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"marketSellOrders\",\"outputs\":[{\"components\":[{\"name\":\"makerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"makerFeePaid\",\"type\":\"uint256\"},{\"name\":\"takerFeePaid\",\"type\":\"uint256\"}],\"name\":\"totalFillResults\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"components\":[{\"name\":\"makerAddress\",\"type\":\"address\"},{\"name\":\"takerAddress\",\"type\":\"address\"},{\"name\":\"feeRecipientAddress\",\"type\":\"address\"},{\"name\":\"senderAddress\",\"type\":\"address\"},{\"name\":\"makerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"makerFee\",\"type\":\"uint256\"},{\"name\":\"takerFee\",\"type\":\"uint256\"},{\"name\":\"expirationTimeSeconds\",\"type\":\"uint256\"},{\"name\":\"salt\",\"type\":\"uint256\"},{\"name\":\"makerAssetData\",\"type\":\"bytes\"},{\"name\":\"takerAssetData\",\"type\":\"bytes\"}],\"name\":\"orders\",\"type\":\"tuple[]\"}],\"name\":\"getOrdersInfo\",\"outputs\":[{\"components\":[{\"name\":\"orderStatus\",\"type\":\"uint8\"},{\"name\":\"orderHash\",\"type\":\"bytes32\"},{\"name\":\"orderTakerAssetFilledAmount\",\"type\":\"uint256\"}],\"name\":\"\",\"type\":\"tuple[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"},{\"name\":\"\",\"type\":\"address\"}],\"name\":\"preSigned\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"},{\"name\":\"signerAddress\",\"type\":\"address\"},{\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"isValidSignature\",\"outputs\":[{\"name\":\"isValid\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"name\":\"makerAddress\",\"type\":\"address\"},{\"name\":\"takerAddress\",\"type\":\"address\"},{\"name\":\"feeRecipientAddress\",\"type\":\"address\"},{\"name\":\"senderAddress\",\"type\":\"address\"},{\"name\":\"makerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"makerFee\",\"type\":\"uint256\"},{\"name\":\"takerFee\",\"type\":\"uint256\"},{\"name\":\"expirationTimeSeconds\",\"type\":\"uint256\"},{\"name\":\"salt\",\"type\":\"uint256\"},{\"name\":\"makerAssetData\",\"type\":\"bytes\"},{\"name\":\"takerAssetData\",\"type\":\"bytes\"}],\"name\":\"orders\",\"type\":\"tuple[]\"},{\"name\":\"makerAssetFillAmount\",\"type\":\"uint256\"},{\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"marketBuyOrdersNoThrow\",\"outputs\":[{\"components\":[{\"name\":\"makerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"makerFeePaid\",\"type\":\"uint256\"},{\"name\":\"takerFeePaid\",\"type\":\"uint256\"}],\"name\":\"totalFillResults\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"name\":\"makerAddress\",\"type\":\"address\"},{\"name\":\"takerAddress\",\"type\":\"address\"},{\"name\":\"feeRecipientAddress\",\"type\":\"address\"},{\"name\":\"senderAddress\",\"type\":\"address\"},{\"name\":\"makerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"makerFee\",\"type\":\"uint256\"},{\"name\":\"takerFee\",\"type\":\"uint256\"},{\"name\":\"expirationTimeSeconds\",\"type\":\"uint256\"},{\"name\":\"salt\",\"type\":\"uint256\"},{\"name\":\"makerAssetData\",\"type\":\"bytes\"},{\"name\":\"takerAssetData\",\"type\":\"bytes\"}],\"name\":\"order\",\"type\":\"tuple\"},{\"name\":\"takerAssetFillAmount\",\"type\":\"uint256\"},{\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"fillOrder\",\"outputs\":[{\"components\":[{\"name\":\"makerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"makerFeePaid\",\"type\":\"uint256\"},{\"name\":\"takerFeePaid\",\"type\":\"uint256\"}],\"name\":\"fillResults\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"salt\",\"type\":\"uint256\"},{\"name\":\"signerAddress\",\"type\":\"address\"},{\"name\":\"data\",\"type\":\"bytes\"},{\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"executeTransaction\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"assetProxy\",\"type\":\"address\"}],\"name\":\"registerAssetProxy\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"components\":[{\"name\":\"makerAddress\",\"type\":\"address\"},{\"name\":\"takerAddress\",\"type\":\"address\"},{\"name\":\"feeRecipientAddress\",\"type\":\"address\"},{\"name\":\"senderAddress\",\"type\":\"address\"},{\"name\":\"makerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"makerFee\",\"type\":\"uint256\"},{\"name\":\"takerFee\",\"type\":\"uint256\"},{\"name\":\"expirationTimeSeconds\",\"type\":\"uint256\"},{\"name\":\"salt\",\"type\":\"uint256\"},{\"name\":\"makerAssetData\",\"type\":\"bytes\"},{\"name\":\"takerAssetData\",\"type\":\"bytes\"}],\"name\":\"order\",\"type\":\"tuple\"}],\"name\":\"getOrderInfo\",\"outputs\":[{\"components\":[{\"name\":\"orderStatus\",\"type\":\"uint8\"},{\"name\":\"orderHash\",\"type\":\"bytes32\"},{\"name\":\"orderTakerAssetFilledAmount\",\"type\":\"uint256\"}],\"name\":\"orderInfo\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"name\":\"makerAddress\",\"type\":\"address\"},{\"name\":\"takerAddress\",\"type\":\"address\"},{\"name\":\"feeRecipientAddress\",\"type\":\"address\"},{\"name\":\"senderAddress\",\"type\":\"address\"},{\"name\":\"makerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"makerFee\",\"type\":\"uint256\"},{\"name\":\"takerFee\",\"type\":\"uint256\"},{\"name\":\"expirationTimeSeconds\",\"type\":\"uint256\"},{\"name\":\"salt\",\"type\":\"uint256\"},{\"name\":\"makerAssetData\",\"type\":\"bytes\"},{\"name\":\"takerAssetData\",\"type\":\"bytes\"}],\"name\":\"order\",\"type\":\"tuple\"}],\"name\":\"cancelOrder\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"address\"}],\"name\":\"orderEpoch\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"ZRX_ASSET_DATA\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"name\":\"makerAddress\",\"type\":\"address\"},{\"name\":\"takerAddress\",\"type\":\"address\"},{\"name\":\"feeRecipientAddress\",\"type\":\"address\"},{\"name\":\"senderAddress\",\"type\":\"address\"},{\"name\":\"makerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"makerFee\",\"type\":\"uint256\"},{\"name\":\"takerFee\",\"type\":\"uint256\"},{\"name\":\"expirationTimeSeconds\",\"type\":\"uint256\"},{\"name\":\"salt\",\"type\":\"uint256\"},{\"name\":\"makerAssetData\",\"type\":\"bytes\"},{\"name\":\"takerAssetData\",\"type\":\"bytes\"}],\"name\":\"orders\",\"type\":\"tuple[]\"},{\"name\":\"takerAssetFillAmount\",\"type\":\"uint256\"},{\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"marketSellOrdersNoThrow\",\"outputs\":[{\"components\":[{\"name\":\"makerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"makerFeePaid\",\"type\":\"uint256\"},{\"name\":\"takerFeePaid\",\"type\":\"uint256\"}],\"name\":\"totalFillResults\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"EIP712_DOMAIN_HASH\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"name\":\"makerAddress\",\"type\":\"address\"},{\"name\":\"takerAddress\",\"type\":\"address\"},{\"name\":\"feeRecipientAddress\",\"type\":\"address\"},{\"name\":\"senderAddress\",\"type\":\"address\"},{\"name\":\"makerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetAmount\",\"type\":\"uint256\"},{\"name\":\"makerFee\",\"type\":\"uint256\"},{\"name\":\"takerFee\",\"type\":\"uint256\"},{\"name\":\"expirationTimeSeconds\",\"type\":\"uint256\"},{\"name\":\"salt\",\"type\":\"uint256\"},{\"name\":\"makerAssetData\",\"type\":\"bytes\"},{\"name\":\"takerAssetData\",\"type\":\"bytes\"}],\"name\":\"orders\",\"type\":\"tuple[]\"},{\"name\":\"makerAssetFillAmount\",\"type\":\"uint256\"},{\"name\":\"signatures\",\"type\":\"bytes[]\"}],\"name\":\"marketBuyOrders\",\"outputs\":[{\"components\":[{\"name\":\"makerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"takerAssetFilledAmount\",\"type\":\"uint256\"},{\"name\":\"makerFeePaid\",\"type\":\"uint256\"},{\"name\":\"takerFeePaid\",\"type\":\"uint256\"}],\"name\":\"totalFillResults\",\"type\":\"tuple\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"currentContextAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"VERSION\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_zrxAssetData\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"signerAddress\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"validatorAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"SignatureValidatorApproval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"makerAddress\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"feeRecipientAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"takerAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"senderAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"makerAssetFilledAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"takerAssetFilledAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"makerFeePaid\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"takerFeePaid\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"orderHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"makerAssetData\",\"type\":\"bytes\"},{\"indexed\":false,\"name\":\"takerAssetData\",\"type\":\"bytes\"}],\"name\":\"Fill\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"makerAddress\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"feeRecipientAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"senderAddress\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"orderHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"makerAssetData\",\"type\":\"bytes\"},{\"indexed\":false,\"name\":\"takerAssetData\",\"type\":\"bytes\"}],\"name\":\"Cancel\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"makerAddress\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"senderAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"orderEpoch\",\"type\":\"uint256\"}],\"name\":\"CancelUpTo\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"id\",\"type\":\"bytes4\"},{\"indexed\":false,\"name\":\"assetProxy\",\"type\":\"address\"}],\"name\":\"AssetProxyRegistered\",\"type\":\"event\"}]",
    "ContractAdd": "0x080bf510fcbf18b91105470639e9561022937712",
    "TriggerName": "WAE MATTEO",
    "TriggerType": "WatchEvents"
}`
	tg, err := NewTriggerFromJson(js)
	assert.NoError(t, err)

	matches := MatchEvent(config.CliMain, tg, 9099675, 1572344236)
	assert.Equal(t, 1, len(matches))
	assert.Equal(t, 9099675, matches[0].Log.BlockNumber)
}

func TestMatchEvent1(t *testing.T) {

	var client EthMock

	tg1, err := GetTriggerFromFile("../resources/triggers/ev1.json")
	assert.NoError(t, err)
	matches1 := MatchEvent(client, tg1, 8496661, 1572344236)

	assert.Equal(t, 1, len(matches1))
	assert.Equal(t, "677420000", matches1[0].EventParams["value"])

	tg2, err := GetTriggerFromFile("../resources/triggers/ev2.json")
	assert.NoError(t, err)
	matches2 := MatchEvent(client, tg2, 8496661, 1572344236)

	assert.Equal(t, 3, len(matches2))
	assert.Equal(t, "677420000", matches2[0].EventParams["value"])
	assert.Equal(t, "771470000", matches2[1].EventParams["value"])
	assert.Equal(t, "607760000", matches2[2].EventParams["value"])

	assert.Equal(t, 3, len(matches2[0].EventParams))
	assert.Equal(t, "0xf750f050e5596eb9480523eef7260b1535a689bd", matches2[0].EventParams["from"])
	assert.Equal(t, "0xcd95b32c98423172e04b1c76841e5a73f4532a7f", matches2[0].EventParams["to"])
	assert.Equal(t, "677420000", matches2[0].EventParams["value"])

	// testing ToPersistent()
	persistentJson, err := utils.GimmePrettyJson(matches1[0].ToPersistent())
	expectedJsn := `{
  "ContractAdd": "0xdac17f958d2ee523a2206206994597c13d831ec7",
  "EventName": "Transfer",
  "EventData": {
    "EventParameters": {
      "from": "0xf750f050e5596eb9480523eef7260b1535a689bd",
      "to": "0xcd95b32c98423172e04b1c76841e5a73f4532a7f",
      "value": "677420000"
    },
    "Data": "0x000000000000000000000000000000000000000000000000000000002439ae80",
    "Topics": [
      "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
      "0x000000000000000000000000f3272a8f1da1f23979c63e328e4dfb35bdf5ff36",
      "0x000000000000000000000000110f0bffb53c82a172edaf007fcaa3f56ed360b0"
    ]
  },
  "Transaction": {
    "BlockHash": "0xf3d70d822816015f26843d378b8c1d5d5da62f5d346f3e86d91a0c2463d30543",
    "BlockNumber": 8496661,
    "BlockTimestamp": 1572344236,
    "Hash": "0xab5e7b8ec9eaf3aaffff797a7992780e9c1c717dfdb5dca2b76b0b71cf182f52"
  }
}`
	ok, err := utils.AreEqualJSON(persistentJson, expectedJsn)
	assert.NoError(t, err)
	assert.True(t, ok)
}
