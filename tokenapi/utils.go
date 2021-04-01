package tokenapi

import (
	"fmt"
	"github.com/HAL-xyz/zoroaster/utils"
	"math"
	"math/big"
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
	result, err := utils.DecodeParamsIntoList(strings.TrimPrefix(rawData, "0x"), erc20abi, methodName)
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
