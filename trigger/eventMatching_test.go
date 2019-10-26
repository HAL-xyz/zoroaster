package trigger

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/onrik/ethrpc"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"testing"
)

func TestValidateFilterLog(t *testing.T) {

	logs, err := GetLogsFromFile("../resources/events/logs1.json")
	if err != nil {
		t.Error(err)
	}
	tg, err := NewTriggerFromFile("../resources/triggers/ev1.json")
	if err != nil {
		t.Error(err)
	}
	abiObj, err := abi.JSON(strings.NewReader(tg.ContractABI))
	namedTopics := getNamedTopics(abiObj, tg.Filters[0].EventName)

	res := validateFilterLog(&logs[0], &tg.Filters[0], &abiObj, tg.Filters[0].EventName, namedTopics)
	assert.True(t, res)

	res2 := validateFilterLog(&logs[0], &tg.Filters[1], &abiObj, tg.Filters[0].EventName, namedTopics)
	assert.True(t, res2)

	eventSignature, err := getEventSignature(tg.ContractABI, tg.Filters[0].EventName)
	res3 := validateTriggerLog(&logs[0], tg, &abiObj, tg.Filters[0].EventName, eventSignature, namedTopics)
	assert.True(t, res3)

	res4 := validateTriggerLog(&logs[1], tg, &abiObj, tg.Filters[0].EventName, eventSignature, namedTopics)
	assert.False(t, res4)
}

type EthMock struct{}

func (cli EthMock) EthGetLogs(params ethrpc.FilterParams) ([]ethrpc.Log, error) {
	return GetLogsFromFile("../resources/events/logs1.json")
}

func TestMatchEvent(t *testing.T) {

	var client EthMock

	tg1, _ := NewTriggerFromFile("../resources/triggers/ev1.json")
	matches1 := MatchEvent(client, tg1, 8496661)

	assert.Equal(t, 1, len(matches1))
	assert.Equal(t, big.NewInt(677420000), matches1[0].decodedData["value"])

	tg2, _ := NewTriggerFromFile("../resources/triggers/ev2.json")
	matches2 := MatchEvent(client, tg2, 8496661)

	assert.Equal(t, 3, len(matches2))
	assert.Equal(t, big.NewInt(677420000), matches2[0].decodedData["value"])
	assert.Equal(t, big.NewInt(771470000), matches2[1].decodedData["value"])
	assert.Equal(t, big.NewInt(607760000), matches2[2].decodedData["value"])
}
