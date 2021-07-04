package tokenapi

import (
	"encoding/json"
	"fmt"
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/ethereum/go-ethereum/accounts/abi"
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
	EthCall(address, method, abiJsn string, blockNo int, args ...string) ([]interface{}, error)
	GetRPCCli() IEthRpc
}

type TokenAPI struct {
	fiatCache        *cache.Cache
	fiatCacheHistory *cache.Cache
	fiatStats        map[string]int
	httpCli          *http.Client
	rpcCli           IEthRpc
	tokenMap         map[string]ERC20Token
	TokenEndpoint    string
	coingeckoIdsMap  map[string]string
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
		fiatCache:        cache.New(15*time.Minute, 15*time.Minute),
		fiatCacheHistory: cache.New(24*time.Hour, 24*time.Hour),
		fiatStats:        map[string]int{},
		httpCli:          &http.Client{},
		rpcCli:           cli,
		TokenEndpoint:    "https://23m8idpr31.execute-api.eu-central-1.amazonaws.com/PROD/v1",
		coingeckoIdsMap:  map[string]string{},
	}
	return &tapi
}

// Initialize the ERC20 map of all tokens.
// Only the methods that actually need the map will call this, so we don't
// load it every time we create an instance of token api for whatever reason
func (t *TokenAPI) init() {
	t.Lock()
	tokenMapLength := len(t.tokenMap)
	t.Unlock()
	if tokenMapLength == 0 {
		resp, err := http.Get(fmt.Sprintf("%s/all_tokens", t.TokenEndpoint))
		defer resp.Body.Close()
		if err != nil {
			log.Fatalf("cannot init TokenAPI: %s", err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("cannot init TokenAPI: %s", err)
		}
		t.Lock()
		err = json.Unmarshal(body, &t.tokenMap)
		t.Unlock()
		if err != nil {
			log.Fatalf("cannot init TokenAPI: %s", err)
		}
	}
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

// The token list is initialized lazily, e.g. the first time this method is called, rather
// than when the client is created. This allows us to mock the token map for tests.
func (t *TokenAPI) GetAllERC20TokensMap() map[string]ERC20Token {
	t.init()
	return t.tokenMap
}

func (t *TokenAPI) Symbol(address string) string {
	t.init()
	_, ok := t.tokenMap[address]
	if ok {
		return t.tokenMap[address].Symbol
	}
	token, err := t.callERC20api(address)
	if err != nil {
		return ""
	}
	return token.Symbol
}

func (t *TokenAPI) Decimals(address string) string {
	t.init()
	_, ok := t.tokenMap[address]
	if ok {
		return fmt.Sprintf("%d", t.tokenMap[address].Decimals)
	}
	token, err := t.callERC20api(address)
	if err != nil {
		return "18"
	}
	return fmt.Sprintf("%d", token.Decimals)
}

func (t *TokenAPI) BalanceOf(token string, user string) string {
	if isEthereumAddress(token) {
		return "0"
	}

	paramInput := Input{
		ParameterType:  "address",
		ParameterValue: user,
	}

	methodHash, err := encodeMethod("balanceOf", erc20abi, []Input{paramInput})

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

func (t *TokenAPI) GetExchangeRate(tokenAddress, fiatCurrency string) (float32, error) {
	tokenAddress = strings.ToLower(tokenAddress)
	fiatCurrency = strings.ToLower(fiatCurrency)

	if len(tokenAddress) != 42 {
		return 0, fmt.Errorf("invalid token address: %s", tokenAddress)
	}

	var coinGeckoUrl string
	if isEthereumAddress(tokenAddress) {
		coinGeckoUrl = fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=%s", fiatCurrency)
		tokenAddress = "ethereum"
	} else {
		coinGeckoUrl = fmt.Sprintf("https://api.coingecko.com/api/v3/simple/token_price/%s?contract_addresses=%s&vs_currencies=%s", getCoingeckoNetwork(), tokenAddress, fiatCurrency)
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

func getCoingeckoNetwork() string {
	if config.Zconf.IsNetworkETHMainnet() {
		return "ethereum"
	}
	if config.Zconf.IsNetworkPolygon() {
		return "polygon-pos"
	}
	if config.Zconf.IsNetworkXDAI() {
		return "xdai"
	}
	if config.Zconf.IsNetworkBinance() {
		return "binance-smart-chain"
	}
	return ""
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

func (t *TokenAPI) callERC20api(address string) (ERC20Token, error) {
	resp, err := http.Get(fmt.Sprintf("%s/token?address=%s", t.TokenEndpoint, address))
	if err != nil {
		return ERC20Token{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ERC20Token{}, err
	}
	var tokenMap map[string]ERC20Token
	err = json.Unmarshal(body, &tokenMap)
	if err != nil {
		return ERC20Token{}, err
	}
	return tokenMap[address], nil
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

// EthCall is a high level helper that executes a view call against a contract
// If the abi is not provided, we rely on Etherscan to fetch it
func (t *TokenAPI) EthCall(address, method, abiJsn string, blockNo int, args ...string) ([]interface{}, error) {

	var err error
	if abiJsn == "" {
		abiJsn, err = fetchAbi(address)
		if err != nil {
			return []interface{}{}, fmt.Errorf("cannot fetch abi for contract: %s - %s", address, err)
		}
	}
	abiObj, err := abi.JSON(strings.NewReader(abiJsn))
	if err != nil {
		return []interface{}{}, err
	}
	inputMethod, ok := abiObj.Methods[method]
	if !ok {
		return []interface{}{}, fmt.Errorf("cannot find method %s", method)
	}

	if len(args) != len(inputMethod.Inputs) {
		return []interface{}{}, fmt.Errorf("invalid number of arguments for method %s - expected %d, got %d", method, len(inputMethod.Inputs), len(args))
	}
	inputs := make([]Input, len(args))
	for i, arg := range args {
		inputs[i] = Input{
			ParameterValue: arg,
			ParameterType:  inputMethod.Inputs[i].Type.String(),
		}
	}

	methodId, err := encodeMethod(method, abiJsn, inputs)
	if err != nil {
		return []interface{}{}, fmt.Errorf("cannot encode method: %s", err)
	}
	rawData, err := t.GetRPCCli().MakeEthRpcCall(address, methodId, blockNo)
	if err != nil {
		return []interface{}{}, fmt.Errorf("rpc call failed with error : %s", err)
	}
	if rawData == "0x" {
		return []interface{}{}, fmt.Errorf("rpc call failed: returned 0x")
	}

	result, err := decodeParams(strings.TrimPrefix(rawData, "0x"), abiJsn, method)
	if err != nil {
		return []interface{}{}, err
	}

	return result, nil
}
