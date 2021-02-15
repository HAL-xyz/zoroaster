package tokenapi

import (
	"encoding/json"
	"fmt"
	"github.com/HAL-xyz/ethrpc"
	"github.com/HAL-xyz/zoroaster/abidec"
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/HAL-xyz/zoroaster/rpc"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ITokenAPI interface {
	GetRPCCli() rpc.IEthRpc
	Symbol(address string) string
	Decimals(address string) string
	GetCacheTokenPriceCacheCount() int
	BalanceOf(token string, user string) string
	FromWei(wei interface{}, units interface{}) string
	GetExchangeRate(tokenAddress, fiatCurrency string) (float32, error)
	EncodeMethod(methodName, cntABI string, inputs []Input) (string, error)
	MakeEthRpcCall(cntAddress, data string, blockNumber int) (string, error)
}

type TokenAPI struct {
	tokenAddToPrice *cache.Cache
	httpCli         *http.Client
	rpcCli          rpc.IEthRpc
	tokenCache      *cache.Cache
}

// package-level singleton accessed through GetTokenAPI()
// some day it would be nice to pass it explicitly as a dependency of the templating system
var tokenApi = &TokenAPI{
	tokenAddToPrice: cache.New(5*time.Minute, 5*time.Minute),
	httpCli:         &http.Client{},
	rpcCli:          rpc.New(ethrpc.New(config.Zconf.EthNode), "templating client"),
	tokenCache:      cache.New(12*time.Hour, 12*time.Hour),
}

func GetTokenAPI() *TokenAPI {
	return tokenApi
}

// returns a new TokenAPI
func New(cli rpc.IEthRpc) *TokenAPI {
	return &TokenAPI{
		tokenAddToPrice: cache.New(5*time.Minute, 5*time.Minute),
		httpCli:         &http.Client{},
		rpcCli:          cli,
		tokenCache:      cache.New(12*time.Hour, 12*time.Hour),
	}
}

func (t TokenAPI) GetCacheTokenPriceCacheCount() int {
	return t.tokenAddToPrice.ItemCount()
}

func (t TokenAPI) GetRPCCli() rpc.IEthRpc {
	return t.rpcCli
}

func (t TokenAPI) ResetETHRPCstats(blockNo int) {
	t.rpcCli.ResetCounterAndLogStats(blockNo)
}

const erc20abi = `[ { "constant": true, "inputs": [], "name": "name", "outputs": [ { "name": "", "type": "string" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": false, "inputs": [ { "name": "_spender", "type": "address" }, { "name": "_value", "type": "uint256" } ], "name": "approve", "outputs": [ { "name": "", "type": "bool" } ], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [], "name": "totalSupply", "outputs": [ { "name": "", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": false, "inputs": [ { "name": "_from", "type": "address" }, { "name": "_to", "type": "address" }, { "name": "_value", "type": "uint256" } ], "name": "transferFrom", "outputs": [ { "name": "", "type": "bool" } ], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [], "name": "decimals", "outputs": [ { "name": "", "type": "uint8" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": true, "inputs": [ { "name": "_owner", "type": "address" } ], "name": "balanceOf", "outputs": [ { "name": "balance", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": true, "inputs": [], "name": "symbol", "outputs": [ { "name": "", "type": "string" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": false, "inputs": [ { "name": "_to", "type": "address" }, { "name": "_value", "type": "uint256" } ], "name": "transfer", "outputs": [ { "name": "", "type": "bool" } ], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [ { "name": "_owner", "type": "address" }, { "name": "_spender", "type": "address" } ], "name": "allowance", "outputs": [ { "name": "", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "payable": true, "stateMutability": "payable", "type": "fallback" }, { "anonymous": false, "inputs": [ { "indexed": true, "name": "owner", "type": "address" }, { "indexed": true, "name": "spender", "type": "address" }, { "indexed": false, "name": "value", "type": "uint256" } ], "name": "Approval", "type": "event" }, { "anonymous": false, "inputs": [ { "indexed": true, "name": "from", "type": "address" }, { "indexed": true, "name": "to", "type": "address" }, { "indexed": false, "name": "value", "type": "uint256" } ], "name": "Transfer", "type": "event" } ]`

func (t TokenAPI) callERC20(address, methodHash, methodName string) string {

	lastBlock, err := t.rpcCli.EthBlockNumber()
	if err != nil {
		return err.Error()
	}
	rawData, err := t.MakeEthRpcCall(address, methodHash, lastBlock)
	if err != nil {
		return err.Error()
	}
	result, err := abidec.DecodeParamsIntoList(strings.TrimPrefix(rawData, "0x"), erc20abi, methodName)
	if err != nil {
		return err.Error()
	}
	if len(result) == 1 {
		return fmt.Sprintf("%v", result[0])
	}
	return address
}

func (t TokenAPI) Symbol(address string) string {
	if isEthereumAddress(address) {
		return "ETH"
	}
	return t.callERC20(address, "0x95d89b41", "symbol")
}

func (t TokenAPI) Decimals(address string) string {
	if isEthereumAddress(address) {
		return "18"
	}

	dec, found := t.tokenCache.Get(address)
	if found {
		return dec.(string)
	}
	dec = t.callERC20(address, "0x313ce567", "decimals")
	t.tokenCache.Set(address, dec, cache.DefaultExpiration)
	return dec.(string)
}

func (t TokenAPI) BalanceOf(token string, user string) string {
	if isEthereumAddress(token) {
		return "0"
	}

	paramInput := Input{
		ParameterType:  "address",
		ParameterValue: user,
	}

	methodHash, err := t.EncodeMethod("balanceOf", erc20abi, []Input{paramInput})

	if err != nil {
		return err.Error()
	}

	return t.callERC20(token, methodHash, "balanceOf")
}

func (t TokenAPI) MakeEthRpcCall(cntAddress, data string, blockNumber int) (string, error) {
	params := ethrpc.T{
		To:   cntAddress,
		From: "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		Data: data,
	}
	return t.rpcCli.EthCall(params, fmt.Sprintf("0x%x", blockNumber))
}

func (t TokenAPI) FromWei(wei interface{}, units interface{}) string {
	var unit int
	switch t := units.(type) {
	case string:
		var err error
		unit, err = strconv.Atoi(t)
		if err != nil {
			return fmt.Sprintf("cannot use %v of type %T as units", unit, unit)
		}
	case int:
		unit = t
	}
	switch v := wei.(type) {
	case *big.Int:
		return scaleBy(v.String(), fmt.Sprintf("%f", math.Pow10(unit)))
	case string:
		return scaleBy(v, fmt.Sprintf("%f", math.Pow10(unit)))
	case int:
		return scaleBy(strconv.Itoa(v), fmt.Sprintf("%f", math.Pow10(unit)))
	default:
		return fmt.Sprintf("cannot use %v of type %T as wei input", wei, wei)
	}
}

func scaleBy(text, scaleBy string) string {
	v := new(big.Float)
	v, ok := v.SetString(text)
	if !ok {
		return text
	}
	scale, _ := new(big.Float).SetString(scaleBy)
	res, _ := new(big.Float).Quo(v, scale).Float64()

	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.4f", math.Ceil(res*10000)/10000), "0"), ".")
}

func (t TokenAPI) GetExchangeRate(tokenAddress, fiatCurrency string) (float32, error) {
	tokenAddress = strings.ToLower(tokenAddress)
	fiatCurrency = strings.ToLower(fiatCurrency)

	var coinGeckoUrl string
	if isEthereumAddress(tokenAddress) {
		coinGeckoUrl = fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=%s", fiatCurrency)
		tokenAddress = "ethereum"
	} else {
		coinGeckoUrl = fmt.Sprintf("https://api.coingecko.com/api/v3/simple/token_price/ethereum?contract_addresses=%s&vs_currencies=%s", tokenAddress, fiatCurrency)
	}

	key := tokenAddress + fiatCurrency

	// try cache first
	price, found := t.tokenAddToPrice.Get(key)
	if found {
		return price.(float32), nil
	}

	// call CoinGecko
	price, err := t.callPriceAPIs(coinGeckoUrl, tokenAddress, fiatCurrency)
	if err == nil {
		t.tokenAddToPrice.Set(key, price, 5*time.Minute)
		return price.(float32), nil
	}

	// fallback to our own endpoint
	customEndpoint := fmt.Sprintf("https://xyxoolw445.execute-api.us-east-1.amazonaws.com/dev/%s", tokenAddress)
	price, err = t.callPriceAPIs(customEndpoint, tokenAddress, fiatCurrency)
	if err == nil {
		t.tokenAddToPrice.Set(key, price, 5*time.Minute)
		return price.(float32), nil
	}
	// sorry :(
	return 0, fmt.Errorf("cannot find %s value for token %s\n", fiatCurrency, tokenAddress)
}

func (t TokenAPI) callPriceAPIs(url, tokenAddress, fiatCurrency string) (float32, error) {
	// all APIs return data in this format
	// {
	// 	 "0xb1cd6e4153b2a390cf00a6556b0fc1458c4a5533": {
	//	   "usd": 1.58
	//	 }
	// }

	resp, err := t.httpCli.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var currencyMap map[string]map[string]float32
	err = json.Unmarshal(body, &currencyMap)
	if err != nil {
		return 0, err
	}

	val, ok := currencyMap[tokenAddress][fiatCurrency]
	if ok {
		return val, nil
	} else {
		return 0, fmt.Errorf("not found")
	}
}

func isEthereumAddress(s string) bool {
	return strings.ToLower(s) == "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" || strings.ToLower(s) == "0x0000000000000000000000000000000000000000"
}
