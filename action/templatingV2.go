package action

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/HAL-xyz/zoroaster/abidec"
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/HAL-xyz/zoroaster/trigger"
	"html/template"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
)

func renderTemplateWithData(templateText string, data interface{}) (string, error) {

	funcMap := template.FuncMap{
		"upperCase":            strings.ToUpper,
		"hexToASCII":           hexToASCII,
		"hexToInt":             hexToInt,
		"etherscanTxLink":      etherscanTxLink,
		"etherscanAddressLink": etherscanAddressLink,
		"etherscanTokenLink":   etherscanTokenLink,
		"fromWei":              fromWei,
		"humanTime":            timestampToHumanTime,
		"symbol":               symbol,
		"decimals":             decimals,
	}

	tmpl := template.New("").Funcs(funcMap)
	t, err := tmpl.Parse(templateText)

	if err != nil {
		return fmt.Sprintf("Could not parse template: %s\n", err), err
	}

	var output bytes.Buffer
	err = t.Execute(&output, data)

	return output.String(), err
}

func hexToASCII(s string) string {
	s = strings.TrimPrefix(s, "0x")
	bs, err := hex.DecodeString(s)
	if err != nil {
		return s
	}
	return string(bs)
}

func hexToInt(s string) string {
	s = strings.TrimPrefix(s, "0x")
	i := new(big.Int)
	_, ok := i.SetString(s, 16)
	if !ok {
		return s
	}
	return i.String()
}

func etherscanTxLink(hash string) string {
	return fmt.Sprintf("https://etherscan.io/tx/%s", hash)
}

func etherscanAddressLink(address string) string {
	return fmt.Sprintf("https://etherscan.io/address/%s", address)
}

func etherscanTokenLink(token string) string {
	return fmt.Sprintf("https://etherscan.io/token/%s", token)
}

func fromWei(wei interface{}, units int) string {
	switch v := wei.(type) {
	case *big.Int:
		return scaleBy(v.String(), fmt.Sprintf("%f", math.Pow10(units)))
	case string:
		return scaleBy(v, fmt.Sprintf("%f", math.Pow10(units)))
	case int:
		return scaleBy(strconv.Itoa(v), fmt.Sprintf("%f", math.Pow10(units)))
	default:
		return fmt.Sprintf("cannot use %s of type %T as wei input", wei, wei)
	}
}

func timestampToHumanTime(timestamp interface{}) string {
	switch v := timestamp.(type) {
	case int:
		unixTimeUTC := time.Unix(int64(v), 0)
		return unixTimeUTC.Format(time.RFC822)
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return v
		}
		unixTimeUTC := time.Unix(i, 0)
		return unixTimeUTC.Format(time.RFC822)
	default:
		return fmt.Sprintf("cannot use %s of type %T as timestamp input", timestamp, timestamp)
	}
}

func symbol(address string) string {
	return callERC20(address, "0x95d89b41", "symbol")
}

func decimals(address string) string {
	return callERC20(address, "0x313ce567", "decimals")
}

func callERC20(address, methodHash, methodName string) string {

	const erc20abi = `[ { "constant": true, "inputs": [], "name": "name", "outputs": [ { "name": "", "type": "string" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": false, "inputs": [ { "name": "_spender", "type": "address" }, { "name": "_value", "type": "uint256" } ], "name": "approve", "outputs": [ { "name": "", "type": "bool" } ], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [], "name": "totalSupply", "outputs": [ { "name": "", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": false, "inputs": [ { "name": "_from", "type": "address" }, { "name": "_to", "type": "address" }, { "name": "_value", "type": "uint256" } ], "name": "transferFrom", "outputs": [ { "name": "", "type": "bool" } ], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [], "name": "decimals", "outputs": [ { "name": "", "type": "uint8" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": true, "inputs": [ { "name": "_owner", "type": "address" } ], "name": "balanceOf", "outputs": [ { "name": "balance", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": true, "inputs": [], "name": "symbol", "outputs": [ { "name": "", "type": "string" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": false, "inputs": [ { "name": "_to", "type": "address" }, { "name": "_value", "type": "uint256" } ], "name": "transfer", "outputs": [ { "name": "", "type": "bool" } ], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [ { "name": "_owner", "type": "address" }, { "name": "_spender", "type": "address" } ], "name": "allowance", "outputs": [ { "name": "", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "payable": true, "stateMutability": "payable", "type": "fallback" }, { "anonymous": false, "inputs": [ { "indexed": true, "name": "owner", "type": "address" }, { "indexed": true, "name": "spender", "type": "address" }, { "indexed": false, "name": "value", "type": "uint256" } ], "name": "Approval", "type": "event" }, { "anonymous": false, "inputs": [ { "indexed": true, "name": "from", "type": "address" }, { "indexed": true, "name": "to", "type": "address" }, { "indexed": false, "name": "value", "type": "uint256" } ], "name": "Transfer", "type": "event" } ]`

	// TODO: if we switch to a managed node we don't need to fetch last block every time
	lastBlock, err := config.CliMain.EthBlockNumber()
	if err != nil {
		return err.Error()
	}
	rawData, err := trigger.MakeEthRpcCall(config.CliMain, address, methodHash, lastBlock)
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
