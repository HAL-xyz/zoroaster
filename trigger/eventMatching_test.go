package trigger

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/onrik/ethrpc"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"testing"
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

	res := validateFilterLog(&logs[0], &tg.Filters[0], &abiObj, tg.Filters[0].EventName)
	assert.True(t, res)

	res2 := validateFilterLog(&logs[0], &tg.Filters[1], &abiObj, tg.Filters[0].EventName)
	assert.True(t, res2)

	res3 := validateTriggerLog(&logs[0], tg, &abiObj)
	assert.True(t, res3)

	res4 := validateTriggerLog(&logs[1], tg, &abiObj)
	assert.False(t, res4)
}

type EthMock struct{}

func (cli EthMock) EthGetLogs(params ethrpc.FilterParams) ([]ethrpc.Log, error) {
	return GetLogsFromFile("../resources/events/logs1.json")
}

func TestMatchEvent(t *testing.T) {

	var client EthMock

	tg1, err := GetTriggerFromFile("../resources/triggers/ev1.json")
	assert.NoError(t, err)
	matches1 := MatchEvent(client, tg1, 8496661, 1572344236)

	assert.Equal(t, 1, len(matches1))
	assert.Equal(t, big.NewInt(677420000), matches1[0].EventParams["value"])

	tg2, err := GetTriggerFromFile("../resources/triggers/ev2.json")
	assert.NoError(t, err)
	matches2 := MatchEvent(client, tg2, 8496661, 1572344236)

	assert.Equal(t, 3, len(matches2))
	assert.Equal(t, big.NewInt(677420000), matches2[0].EventParams["value"])
	assert.Equal(t, big.NewInt(771470000), matches2[1].EventParams["value"])
	assert.Equal(t, big.NewInt(607760000), matches2[2].EventParams["value"])

	assert.Equal(t, 3, len(matches2[0].EventParams))
	assert.Equal(t, "0x000000000000000000000000f750f050e5596eb9480523eef7260b1535a689bd", matches2[0].EventParams["from"])
	assert.Equal(t, "0x000000000000000000000000cd95b32c98423172e04b1c76841e5a73f4532a7f", matches2[0].EventParams["to"])
	assert.Equal(t, big.NewInt(677420000), matches2[0].EventParams["value"])

	// testing ToPersistent()
	persistentJson, err := utils.GimmePrettyJson(matches1[0].ToPersistent())
	expectedJsn := `{
  "ContractAdd": "0xdac17f958d2ee523a2206206994597c13d831ec7",
  "EventName": "Transfer",
  "EventData": {
    "EventParameters": {
      "from": "0x000000000000000000000000f750f050e5596eb9480523eef7260b1535a689bd",
      "to": "0x000000000000000000000000cd95b32c98423172e04b1c76841e5a73f4532a7f",
      "value": 677420000
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
