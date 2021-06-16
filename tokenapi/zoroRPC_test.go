package tokenapi

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
	"testing"
)

func TestWithRetriesBlockNumber(t *testing.T) {
	defer gock.Off()

	gock.New("https://testenv.com").Post("/").Reply(200).JSON(map[string]string{"jsonrpc": "2.0", "result": ""})         // error
	gock.New("https://testenv.com").Post("/").Reply(200).JSON(map[string]string{"jsonrpc": "2.0", "result": ""})         // error
	gock.New("https://testenv.com").Post("/").Reply(200).JSON(map[string]string{"jsonrpc": "2.0", "result": "0xc18d99"}) // 12684697

	cli := NewZRPC("https://testenv.com", "label", WithRetries(3))
	res, err := cli.EthBlockNumber()
	assert.NoError(t, err)
	assert.Equal(t, 12684697, res)
	assert.Equal(t, 3, cli.calls)
}

func TestWithRetriesLogsByHash(t *testing.T) {
	defer gock.Off()

	// error
	gock.New("https://testenv.com").Post("/").Reply(200).JSON(map[string]string{"jsonrpc": "2.0", "result": ""})
	// success
	err := setupGock("resources/getLogs.json", "https://testenv.com", "/", "POST")
	assert.NoError(t, err)

	cli := NewZRPC("https://testenv.com", "label", WithRetries(2))
	res, err := cli.EthGetLogsByHash("0x22e56dabd801d35788391fe257413ebe744e274884983fa071248949840f7193")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(res))
	assert.Equal(t, 2, cli.calls)
}

func TestWithRetriesLogsByNumber(t *testing.T) {
	defer gock.Off()

	// error
	gock.New("https://testenv.com").Post("/").Reply(200).JSON(map[string]string{"jsonrpc": "2.0", "result": ""})
	// error
	gock.New("https://testenv.com").Post("/").Reply(200).JSON(map[string]string{"jsonrpc": "2.0", "result": ""})
	// success
	err := setupGock("resources/getLogs.json", "https://testenv.com", "/", "POST")
	assert.NoError(t, err)

	cli := NewZRPC("https://testenv.com", "label", WithRetries(3))
	res, err := cli.EthGetLogsByNumber(123, "0x")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(res))
	assert.Equal(t, 3, cli.calls)
}

func TestWithRetriesGetBlockByNumber(t *testing.T) {
	defer gock.Off()

	// error
	gock.New("https://testenv.com").Post("/").Reply(200).JSON(map[string]string{"jsonrpc": "2.0", "result": ""})

	// success
	err := setupGock("resources/getBlock.json", "https://testenv.com", "/", "POST")
	assert.NoError(t, err)

	cli := NewZRPC("https://testenv.com", "label", WithRetries(2))
	res, err := cli.EthGetBlockByNumber(12690259, false)

	assert.NoError(t, err)
	assert.Equal(t, 12690259, res.Number)

}
