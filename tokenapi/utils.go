package tokenapi

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"strings"
	"time"
)

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
	result, err := decodeParams(strings.TrimPrefix(rawData, "0x"), erc20abi, methodName)
	if err != nil {
		return err.Error()
	}
	if len(result) == 1 {
		return fmt.Sprintf("%v", result[0])
	}
	return address
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

func (t *TokenAPI) increaseFiatStats(item string) {
	t.Lock()
	t.fiatStats[item] += 1
	t.Unlock()
}

func isEthereumAddress(s string) bool {
	return strings.ToLower(s) == "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" || strings.ToLower(s) == "0x0000000000000000000000000000000000000000"
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

func decodeParams(data string, cntABI string, methodName string) ([]interface{}, error) {

	encb, err := hex.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("invalid hex: %s", data)
	}

	xabi, err := abi.JSON(strings.NewReader(cntABI))
	if err != nil {
		return nil, fmt.Errorf("cannot read abi: %s", err)
	}

	methodObj, ok := xabi.Methods[methodName]
	if !ok {
		return nil, fmt.Errorf("method %s not found", methodName)
	}

	ls, err := methodObj.Outputs.UnpackValues(encb)

	if err != nil {
		return nil, fmt.Errorf("cannot unpack outputs: %s", err)
	}

	return ls, nil
}

func fetchAbi(address string) (string, error) {
	var etherscanUrl = fmt.Sprintf("https://api.etherscan.io/api?module=contract&action=getabi&address=%s&apikey=%s", address, config.Zconf.EtherscanKey)

	resp, err := http.Get(etherscanUrl)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	m := map[string]string{}
	if err = json.Unmarshal(body, &m); err != nil {
		return "", err
	}
	if m["message"] != "OK" {
		return "", err
	}

	return m["result"], nil
}
