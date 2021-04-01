package tokenapi

import (
	"encoding/json"
	"fmt"
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/HAL-xyz/zoroaster/utils"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ITokenAPI interface {
	GetAllERC20TokensMap() map[string]ERC20Token
	Symbol(address string) string
	Decimals(address string) string
	BalanceOf(token string, user string) string
	FromWei(wei interface{}, units interface{}) string
	GetExchangeRate(tokenAddress, fiatCurrency string) (float32, error)
	GetExchangeRateAtDate(tokenAddress, fiatCurrency, when string) (float32, error)
	LogFiatStatsAndReset(blockNo int)
	GetRPCCli() IEthRpc
}

type TokenAPI struct {
	fiatCache *cache.Cache
	fiatCacheHistory *cache.Cache
	fiatStats map[string]int
	httpCli   *http.Client
	rpcCli    IEthRpc
	tokenMap  map[string]ERC20Token
	TokenEndpoint string
	coingeckoIdsMap map[string]string
	sync.Mutex
}

// package-level singleton accessed through GetTokenAPI()
// some day it would be nice to pass it explicitly as a dependency of the templating system
var tokenApi = New(NewZRPC(config.Zconf.EthNode, "templating client"))

func GetTokenAPI() *TokenAPI {
	return tokenApi
}

// returns a new TokenAPI
func New(cli IEthRpc) *TokenAPI {

	tapi := TokenAPI{
		fiatCache: cache.New(10*time.Minute, 10*time.Minute),
		fiatCacheHistory: cache.New(12*time.Hour, 12*time.Hour),
		fiatStats: map[string]int{},
		httpCli:   &http.Client{},
		rpcCli:    cli,
		TokenEndpoint: "https://23m8idpr31.execute-api.eu-central-1.amazonaws.com/PROD/v1",
		coingeckoIdsMap: map[string]string{},
	}

	return &tapi
}

func (t *TokenAPI) LogFiatStatsAndReset(blockNo int) {
	log.Infof("FiatStats: %s on block %d made %d calls to Coingecko, %d calls to Custom, had %d not-found errors, %d network errors. Cache size is %d",
		t.rpcCli.GetLabel(), blockNo, t.fiatStats["coingecko"], t.fiatStats["custom"], t.fiatStats["not_found"], t.fiatStats["network_error"], t.fiatCache.ItemCount())
	t.Lock()
	for k := range t.fiatStats {
		delete(t.fiatStats, k)
	}
	t.Unlock()
}

func (t *TokenAPI) GetRPCCli() IEthRpc {
	return t.rpcCli
}

func (t *TokenAPI) ResetETHRPCstats(blockNo int) {
	t.rpcCli.ResetCounterAndLogStats(blockNo)
}

type ERC20Token struct {
	ChainId  int    `json:"chainId"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
	LogoURI  string `json:"logoURI,omitempty"`
}

// The token list is initialized lazily, e.g. the first time this method is called, rather
// than when the client is created. This allows us to mock the token map for tests.
func (t *TokenAPI) GetAllERC20TokensMap() map[string]ERC20Token {
	if len(t.tokenMap) == 0 {
		resp, err := http.Get(fmt.Sprintf("%s/all_tokens", t.TokenEndpoint))
		defer resp.Body.Close()
		if err != nil {
			log.Fatalf("cannot init TokenAPI: %s", err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("cannot init TokenAPI: %s", err)
		}
		err = json.Unmarshal(body, &t.tokenMap)
		if err != nil {
			log.Fatalf("cannot init TokenAPI: %s", err)
		}
	}
	return t.tokenMap
}

const erc20abi = `[ { "constant": true, "inputs": [], "name": "name", "outputs": [ { "name": "", "type": "string" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": false, "inputs": [ { "name": "_spender", "type": "address" }, { "name": "_value", "type": "uint256" } ], "name": "approve", "outputs": [ { "name": "", "type": "bool" } ], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [], "name": "totalSupply", "outputs": [ { "name": "", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": false, "inputs": [ { "name": "_from", "type": "address" }, { "name": "_to", "type": "address" }, { "name": "_value", "type": "uint256" } ], "name": "transferFrom", "outputs": [ { "name": "", "type": "bool" } ], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [], "name": "decimals", "outputs": [ { "name": "", "type": "uint8" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": true, "inputs": [ { "name": "_owner", "type": "address" } ], "name": "balanceOf", "outputs": [ { "name": "balance", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": true, "inputs": [], "name": "symbol", "outputs": [ { "name": "", "type": "string" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": false, "inputs": [ { "name": "_to", "type": "address" }, { "name": "_value", "type": "uint256" } ], "name": "transfer", "outputs": [ { "name": "", "type": "bool" } ], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [ { "name": "_owner", "type": "address" }, { "name": "_spender", "type": "address" } ], "name": "allowance", "outputs": [ { "name": "", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "payable": true, "stateMutability": "payable", "type": "fallback" }, { "anonymous": false, "inputs": [ { "indexed": true, "name": "owner", "type": "address" }, { "indexed": true, "name": "spender", "type": "address" }, { "indexed": false, "name": "value", "type": "uint256" } ], "name": "Approval", "type": "event" }, { "anonymous": false, "inputs": [ { "indexed": true, "name": "from", "type": "address" }, { "indexed": true, "name": "to", "type": "address" }, { "indexed": false, "name": "value", "type": "uint256" } ], "name": "Transfer", "type": "event" } ]`

func (t *TokenAPI) callERC20(address, methodHash, methodName string) string {

	lastBlock, err := t.rpcCli.EthBlockNumber()
	if err != nil {
		return err.Error()
	}
	rawData, err := t.GetRPCCli().MakeEthRpcCall(address, methodHash, lastBlock)
	if err != nil {
		return err.Error()
	}
	result, err := utils.DecodeParamsIntoList(strings.TrimPrefix(rawData, "0x"), erc20abi, methodName)
	if err != nil {
		return err.Error()
	}
	if len(result) == 1 {
		return fmt.Sprintf("%v", result[0])
	}
	return address
}

func (t *TokenAPI) Symbol(address string) string {
	return t.tokenMap[address].Symbol
}

func (t *TokenAPI) Decimals(address string) string {
	return fmt.Sprintf("%d", t.tokenMap[address].Decimals)
}

func (t *TokenAPI) BalanceOf(token string, user string) string {
	if isEthereumAddress(token) {
		return "0"
	}

	paramInput := Input{
		ParameterType:  "address",
		ParameterValue: user,
	}

	methodHash, err := t.GetRPCCli().EncodeMethod("balanceOf", erc20abi, []Input{paramInput})

	if err != nil {
		return err.Error()
	}

	return t.callERC20(token, methodHash, "balanceOf")
}

func (t *TokenAPI) FromWei(wei interface{}, units interface{}) string {
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

type ApiNetworkErr struct {
	msg string
}

func (e ApiNetworkErr) Error() string {
	return fmt.Sprintf(e.msg)
}

type ApiNotFoundErr struct {
	msg string
}

func (e ApiNotFoundErr) Error() string {
	return fmt.Sprintf(e.msg)
}

func (t *TokenAPI) GetExchangeRate(tokenAddress, fiatCurrency string) (float32, error) {
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
	price, found := t.fiatCache.Get(key)
	if found {
		return price.(float32), nil
	}

	// call CoinGecko
	price, err := t.callPriceAPIs(coinGeckoUrl, tokenAddress, fiatCurrency)
	if err == nil {
		t.fiatCache.Set(key, price, cache.DefaultExpiration)
		t.increaseFiatStats("coingecko")
		return price.(float32), nil
	}

	if _, ok := err.(ApiNetworkErr); ok {
		t.increaseFiatStats("network_error")
		log.Error(err)
		return 0, err
	}

	// not found on Coingecko, fallback to our own endpoint
	customEndpoint := fmt.Sprintf("https://xyxoolw445.execute-api.us-east-1.amazonaws.com/dev/%s", tokenAddress)
	price, err = t.callPriceAPIs(customEndpoint, tokenAddress, fiatCurrency)
	if err == nil {
		t.fiatCache.Set(key, price, cache.DefaultExpiration)
		t.increaseFiatStats("custom")
		return price.(float32), nil
	}

	// sorry :(
	switch err.(type) {
	case ApiNetworkErr:
		t.increaseFiatStats("network_error")
		log.Error(err)
	case ApiNotFoundErr:
		log.Error(err)
		t.increaseFiatStats("not_found")
	default:
		t.increaseFiatStats("unknown_error")
		log.Errorf("unknown error for currency %s fiat %s", tokenAddress, fiatCurrency)
	}
	return 0, err
}

func (t *TokenAPI) callPriceAPIs(url, tokenAddress, fiatCurrency string) (float32, error) {
	// all APIs return data in this format
	// {
	// 	 "0xb1cd6e4153b2a390cf00a6556b0fc1458c4a5533": {
	//	   "usd": 1.58
	//	 }
	// }

	resp, err := t.httpCli.Get(url)
	if err != nil {
		return 0, ApiNetworkErr{fmt.Sprintf("network error for currency %s fiat %s, %s", tokenAddress, fiatCurrency, err.Error())}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, ApiNetworkErr{fmt.Sprintf("network error for currency %s fiat %s, %s", tokenAddress, fiatCurrency, err.Error())}
	}

	var currencyMap map[string]map[string]float32
	err = json.Unmarshal(body, &currencyMap)
	if err != nil {
		return 0, ApiNotFoundErr{fmt.Sprintf("not found error for currency %s fiat %s, unexpected json input", tokenAddress, fiatCurrency)}
	}

	val, ok := currencyMap[tokenAddress][fiatCurrency]
	if ok {
		return val, nil
	} else {
		return 0, ApiNotFoundErr{fmt.Sprintf("not found error for currency %s fiat %s", tokenAddress, fiatCurrency)}
	}
}

func (t *TokenAPI) increaseFiatStats(item string) {
	t.Lock()
	t.fiatStats[item] += 1
	t.Unlock()
}

func isEthereumAddress(s string) bool {
	return strings.ToLower(s) == "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" || strings.ToLower(s) == "0x0000000000000000000000000000000000000000"
}

type GeckoIDSJson []struct {
	ID        string `json:"id"`
	Platforms struct {
		Ethereum string `json:"ethereum"`
	} `json:"platforms"`
}

func (t *TokenAPI) GetExchangeRateAtDate(tokenAddress, fiatCurrency, when string) (float32, error) {
	tokenAddress = strings.ToLower(tokenAddress)
	fiatCurrency = strings.ToLower(fiatCurrency)

	// first time we run this we download the full list of token-ids for Coingecko
	if len(t.coingeckoIdsMap) == 0 {
		resp, err := http.Get("https://api.coingecko.com/api/v3/coins/list?include_platform=true")
		if err != nil {
			log.Fatal(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		ids := GeckoIDSJson{}
		err = json.Unmarshal(body, &ids)
		if err != nil {
			log.Fatal(err)
		}

		// create a map tokenAdd -> coinGecko-id
		for _, e := range ids {
			if e.Platforms.Ethereum != "" {
				t.coingeckoIdsMap[e.Platforms.Ethereum] = e.ID
			}
		}
		// add ETH entries
		t.coingeckoIdsMap["0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"] = "ethereum"
		t.coingeckoIdsMap["0x0000000000000000000000000000000000000000"] = "ethereum"
	}

	// try cache first
	key := tokenAddress + fiatCurrency + when

	price, found := t.fiatCacheHistory.Get(key)
	if found {
		return price.(float32), nil
	}

	// make request using date
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s/history?date=%s&localization=false", t.coingeckoIdsMap[tokenAddress], parseCurrencyDate(when))
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	m := map[string]json.RawMessage{}
	if err = json.Unmarshal(body, &m); err != nil {
		return 0, err
	}

	mm := map[string]map[string]float32{}
	if err = json.Unmarshal(m["market_data"], &mm); err != nil {
		return 0, err
	}

	historicalPrice := mm["current_price"][fiatCurrency]
	t.fiatCacheHistory.Set(key, historicalPrice, cache.DefaultExpiration)
	t.increaseFiatStats("coingecko")

	return historicalPrice, nil
}

func parseCurrencyDate(when string) string {
	switch when {
	case "yesterday":
		return time.Now().Add(-24 * time.Hour).Format("02-01-2006")
	case "last_week":
		return time.Now().Add(-168 * time.Hour).Format("02-01-2006")
	default:
		return ""
	}
}