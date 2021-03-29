package tokenapi

import (
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

var tapi = New(NewZRPC(config.Zconf.EthNode, "test"))

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

}

func TestTokenAPI_GetExchangeRate(t *testing.T) {

	_, err := tapi.GetExchangeRate("0x9cb9d5429a93174566efa5b5a73cf71e1ca1a8ab", "usd")
	_, ok := err.(ApiNotFoundErr)
	assert.True(t, ok)

	_, err = tapi.GetExchangeRate("0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", "xxx")
	_, ok = err.(ApiNotFoundErr)
	assert.True(t, ok)

}
