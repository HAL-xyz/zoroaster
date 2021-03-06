package tokenapi

import (
	"fmt"
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
	"io/ioutil"
	"os"
	"testing"
)

var tapi = New(NewZRPC(config.Zconf.EthNode, "test"))

func setupGock(filename, url, path, method string) error {
	testJSON, err := os.Open(filename)
	defer testJSON.Close()
	if err != nil {
		return err
	}
	testByte, err := ioutil.ReadAll(testJSON)
	if err != nil {
		return err
	}
	switch method {
	case "GET":
		gock.New(url).Get(path).Reply(200).JSON(testByte)
	case "POST":
		gock.New(url).Post(path).Reply(200).JSON(testByte)
	default:
		return fmt.Errorf("cannot use %s as http method", method)
	}

	return nil
}

func TestTokenAPI_GetFiatCacheCount(t *testing.T) {

	assert.Equal(t, 0, tapi.fiatCache.ItemCount())

	_, err := tapi.GetExchangeRate("0x1f573d6fb3f13d689ff844b4ce37794d79a7ff1c", "usd")
	assert.NoError(t, err)
	assert.Equal(t, 1, tapi.fiatCache.ItemCount())

	_, err = tapi.GetExchangeRate("0x1f573d6fb3f13d689ff844b4ce37794d79a7ff1c", "USD")
	assert.NoError(t, err)
	assert.Equal(t, 1, tapi.fiatCache.ItemCount())

	_, err = tapi.GetExchangeRate("0x1f573d6fb3f13d689ff844b4ce37794d79a7ff1c", "xxx")
	assert.Error(t, err)
	assert.Equal(t, 1, tapi.fiatCache.ItemCount())

	_, err = tapi.GetExchangeRate("0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", "usd")
	assert.NoError(t, err)
	assert.Equal(t, 2, tapi.fiatCache.ItemCount())

	_, err = tapi.GetExchangeRate("0x0000000000000000000000000000000000000000", "usd")
	assert.NoError(t, err)
	assert.Equal(t, 2, tapi.fiatCache.ItemCount())

	// token on Binance
	config.Zconf.Network = "4_binance_mainnet"
	_, err = tapi.GetExchangeRate("0xe9e7cea3dedca5984780bafc599bd69add087d56", "usd")
	assert.NoError(t, err)
	assert.Equal(t, 3, tapi.fiatCache.ItemCount())

	// token on Polygon
	config.Zconf.Network = "5_polygon_mainnet"
	_, err = tapi.GetExchangeRate("0xb33eaad8d922b1083446dc23f610c2567fb5180f", "usd")
	assert.NoError(t, err)
	assert.Equal(t, 4, tapi.fiatCache.ItemCount())
}

func TestTokenAPI_GetExchangeRate(t *testing.T) {

	_, err := tapi.GetExchangeRate("0x9cb9d5429a93174566efa5b5a73cf71e1ca1a8ab", "usd")
	_, ok := err.(ApiNotFoundErr)
	assert.True(t, ok)

	_, err = tapi.GetExchangeRate("0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", "xxx")
	_, ok = err.(ApiNotFoundErr)
}

func TestTokenAPI_GetExchangeRateAtDate(t *testing.T) {

	const baseUrl = "https://api.coingecko.com"

	_ = setupGock("resources/coin_list.json", baseUrl, "/api/v3/coins/list", "GET")
	_ = setupGock("resources/history.json", baseUrl, "/api/v3/coins/usd-coin/history", "GET")
	_ = setupGock("resources/history.json", baseUrl, "/api/v3/coins/usdex-2/history", "GET")

	assert.Equal(t, 0, tapi.fiatCacheHistory.ItemCount())

	res, err := tapi.GetExchangeRateAtDate("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", "usd", "yesterday")
	assert.NoError(t, err)

	assert.Equal(t, 4, len(tapi.coingeckoIdsMap))
	assert.Equal(t, "ethereum", tapi.coingeckoIdsMap["0x0000000000000000000000000000000000000000"])
	assert.Equal(t, "ethereum", tapi.coingeckoIdsMap["0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"])
	assert.Equal(t, "usd-coin", tapi.coingeckoIdsMap["0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"])
	assert.Equal(t, float32(1.0013702), res)
	assert.Equal(t, 1, tapi.fiatCacheHistory.ItemCount())

	// let's call it again
	res, err = tapi.GetExchangeRateAtDate("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", "usd", "yesterday")
	assert.NoError(t, err)

	assert.Equal(t, float32(1.0013702), res)
	assert.Equal(t, 1, tapi.fiatCacheHistory.ItemCount())

	// different token
	res, err = tapi.GetExchangeRateAtDate("0x4726e9de74573255ea41e0d00b49b833c77a671e", "usd", "yesterday")
	assert.NoError(t, err)

	assert.Equal(t, float32(1.0013702), res)
	assert.Equal(t, 2, tapi.fiatCacheHistory.ItemCount())
}
func TestCallERC20API(t *testing.T) {

	_ = setupGock("resources/tokenLookup.json", tapi.TokenEndpoint, "token", "GET")

	token, err := tapi.callERC20api("0x6b175474e89094c44da98b954eedeac495271d0f")
	assert.NoError(t, err)
	assert.Equal(t, "DAI", token.Symbol)
	assert.Equal(t, 18, token.Decimals)
}
