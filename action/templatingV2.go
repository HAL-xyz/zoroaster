package action

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/HAL-xyz/zoroaster/abidec"
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/HAL-xyz/zoroaster/trigger"
	"github.com/HAL-xyz/zoroaster/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/leekchan/accounting"
	"github.com/patrickmn/go-cache"
	"html/template"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
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
		"balanceOf":            balanceOf,
		"call":                 call,
		"add":                  add,
		"sub":                  sub,
		"mul":                  mul,
		"div":                  div,
		"round":                utils.Round,
		"pow":                  pow,
		"formatNumber":         formatNumber,
		"toFiat":               GetTokenAPI().addressToFiat,
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
	bs = bytes.Trim(bs, "\x00")
	return string(bs)
}

func hexToInt(s string) string {
	if !strings.HasPrefix(s, "0x") {
		return s
	}
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

func fromWei(wei interface{}, units interface{}) string {
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

func timestampToHumanTime(timestamp interface{}, optionalFormatting ...string) string {
	var format string
	if len(optionalFormatting) == 1 {
		format = optionalFormatting[0]
	} else {
		format = time.RFC822
	}
	switch v := timestamp.(type) {
	case int:
		unixTimeUTC := time.Unix(int64(v), 0).UTC()
		return unixTimeUTC.Format(format)
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return v
		}
		unixTimeUTC := time.Unix(i, 0).UTC()
		return unixTimeUTC.Format(format)
	default:
		return fmt.Sprintf("cannot use %s of type %T as timestamp input", timestamp, timestamp)
	}
}

func symbol(address string) string {
	if address == "0x0000000000000000000000000000000000000000" || address == "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" {
		return "ETH"
	}
	return callERC20(address, "0x95d89b41", "symbol")
}

func decimals(address string) string {
	if address == "0x0000000000000000000000000000000000000000" || address == "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" {
		return "18"
	}
	return callERC20(address, "0x313ce567", "decimals")
}

func balanceOf(token string, user string) string {
	if token == "0x0000000000000000000000000000000000000000" || token == "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" {
		return "0"
	}

	const erc20abi = `[ { "constant": true, "inputs": [], "name": "name", "outputs": [ { "name": "", "type": "string" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": false, "inputs": [ { "name": "_spender", "type": "address" }, { "name": "_value", "type": "uint256" } ], "name": "approve", "outputs": [ { "name": "", "type": "bool" } ], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [], "name": "totalSupply", "outputs": [ { "name": "", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": false, "inputs": [ { "name": "_from", "type": "address" }, { "name": "_to", "type": "address" }, { "name": "_value", "type": "uint256" } ], "name": "transferFrom", "outputs": [ { "name": "", "type": "bool" } ], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [], "name": "decimals", "outputs": [ { "name": "", "type": "uint8" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": true, "inputs": [ { "name": "_owner", "type": "address" } ], "name": "balanceOf", "outputs": [ { "name": "balance", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": true, "inputs": [], "name": "symbol", "outputs": [ { "name": "", "type": "string" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": false, "inputs": [ { "name": "_to", "type": "address" }, { "name": "_value", "type": "uint256" } ], "name": "transfer", "outputs": [ { "name": "", "type": "bool" } ], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [ { "name": "_owner", "type": "address" }, { "name": "_spender", "type": "address" } ], "name": "allowance", "outputs": [ { "name": "", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "payable": true, "stateMutability": "payable", "type": "fallback" }, { "anonymous": false, "inputs": [ { "indexed": true, "name": "owner", "type": "address" }, { "indexed": true, "name": "spender", "type": "address" }, { "indexed": false, "name": "value", "type": "uint256" } ], "name": "Approval", "type": "event" }, { "anonymous": false, "inputs": [ { "indexed": true, "name": "from", "type": "address" }, { "indexed": true, "name": "to", "type": "address" }, { "indexed": false, "name": "value", "type": "uint256" } ], "name": "Transfer", "type": "event" } ]`
	const methodName = "balanceOf"

	paramInput := trigger.Input{
		ParameterType:  "address",
		ParameterValue: user,
	}

	methodHash, err := trigger.EncodeMethod(methodName, erc20abi, []trigger.Input{paramInput})

	if err != nil {
		return err.Error()
	}

	return callERC20(token, methodHash, methodName)
}

func add(nums ...interface{}) *big.Float {
	total := big.NewFloat(0)
	for _, i := range nums {
		total = total.Add(total, utils.MakeBigFloat(i))
	}
	return total
}

func sub(a, b interface{}) *big.Float {
	total := utils.MakeBigFloat(a)
	return total.Sub(total, utils.MakeBigFloat(b))
}

func mul(a, b interface{}) *big.Float {
	x := utils.MakeBigFloat(a)
	y := utils.MakeBigFloat(b)
	return x.Mul(x, y)
}

func div(a, b interface{}) *big.Float {
	x := utils.MakeBigFloat(a)
	y := utils.MakeBigFloat(b)
	return new(big.Float).Quo(x, y)
}

func pow(a, b interface{}) *big.Float {
	x := utils.MakeBigFloat(a)

	var y int
	if v, ok := b.(int); ok {
		y = v
	}
	if v, err := strconv.Atoi(fmt.Sprintf("%v", b)); err == nil {
		y = v
	}

	var result = big.NewFloat(0.0).Copy(x)
	for i := 0; i < y-1; i++ {
		result = big.NewFloat(0.0).SetPrec(256).Mul(result, x)
	}
	return result
}

func formatNumber(number interface{}, precision int) string {
	return accounting.FormatNumberBigFloat(utils.MakeBigFloat(number), precision, ",", ".")
}

func callERC20(address, methodHash, methodName string) string {
	const erc20abi = `[ { "constant": true, "inputs": [], "name": "name", "outputs": [ { "name": "", "type": "string" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": false, "inputs": [ { "name": "_spender", "type": "address" }, { "name": "_value", "type": "uint256" } ], "name": "approve", "outputs": [ { "name": "", "type": "bool" } ], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [], "name": "totalSupply", "outputs": [ { "name": "", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": false, "inputs": [ { "name": "_from", "type": "address" }, { "name": "_to", "type": "address" }, { "name": "_value", "type": "uint256" } ], "name": "transferFrom", "outputs": [ { "name": "", "type": "bool" } ], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [], "name": "decimals", "outputs": [ { "name": "", "type": "uint8" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": true, "inputs": [ { "name": "_owner", "type": "address" } ], "name": "balanceOf", "outputs": [ { "name": "balance", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": true, "inputs": [], "name": "symbol", "outputs": [ { "name": "", "type": "string" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": false, "inputs": [ { "name": "_to", "type": "address" }, { "name": "_value", "type": "uint256" } ], "name": "transfer", "outputs": [ { "name": "", "type": "bool" } ], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [ { "name": "_owner", "type": "address" }, { "name": "_spender", "type": "address" } ], "name": "allowance", "outputs": [ { "name": "", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "payable": true, "stateMutability": "payable", "type": "fallback" }, { "anonymous": false, "inputs": [ { "indexed": true, "name": "owner", "type": "address" }, { "indexed": true, "name": "spender", "type": "address" }, { "indexed": false, "name": "value", "type": "uint256" } ], "name": "Approval", "type": "event" }, { "anonymous": false, "inputs": [ { "indexed": true, "name": "from", "type": "address" }, { "indexed": true, "name": "to", "type": "address" }, { "indexed": false, "name": "value", "type": "uint256" } ], "name": "Transfer", "type": "event" } ]`

	// TODO: if we switch to a managed node we don't need to fetch last block every time
	lastBlock, err := config.TemplateCli.EthBlockNumber()
	if err != nil {
		return err.Error()
	}
	rawData, err := trigger.MakeEthRpcCall(config.TemplateCli, address, methodHash, lastBlock)
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

// TODO: this is a horrible hack to get LoopRing working ASAP.
// In the future we'll do this properly to make it generic and work with any contract.
func call(contractName, methodName string, params ...string) string {
	if contractName == "LoopringDex" {
		if len(params) != 1 {
			return "no no no"
		}
		abi := `[{"constant":true,"inputs":[],"name":"getLRCFeeForRegisteringOneMoreToken","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"blockIdx","type":"uint256"}],"name":"getBlock","outputs":[{"internalType":"bytes32","name":"merkleRoot","type":"bytes32"},{"internalType":"bytes32","name":"publicDataHash","type":"bytes32"},{"internalType":"uint8","name":"blockState","type":"uint8"},{"internalType":"uint8","name":"blockType","type":"uint8"},{"internalType":"uint16","name":"blockSize","type":"uint16"},{"internalType":"uint32","name":"timestamp","type":"uint32"},{"internalType":"uint32","name":"numDepositRequestsCommitted","type":"uint32"},{"internalType":"uint32","name":"numWithdrawalRequestsCommitted","type":"uint32"},{"internalType":"bool","name":"blockFeeWithdrawn","type":"bool"},{"internalType":"uint16","name":"numWithdrawalsDistributed","type":"uint16"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"tokenAddress","type":"address"}],"name":"registerToken","outputs":[{"internalType":"uint16","name":"","type":"uint16"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[],"name":"clone","outputs":[{"internalType":"address","name":"cloneAddress","type":"address"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"durationMinutes","type":"uint256"}],"name":"getDowntimeCostLRC","outputs":[{"internalType":"uint256","name":"costLRC","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"_addressWhitelist","type":"address"}],"name":"setAddressWhitelist","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256[]","name":"blockIndices","type":"uint256[]"},{"internalType":"uint256[]","name":"proofs","type":"uint256[]"}],"name":"verifyBlocks","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getNumBlocksFinalized","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"isInMaintenance","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"tokenAddress","type":"address"}],"name":"disableTokenDeposit","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"blockIdx","type":"uint256"}],"name":"revertBlock","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"genesisBlockHash","outputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"token","type":"address"}],"name":"withdrawProtocolFees","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"pubKeyX","type":"uint256"},{"internalType":"uint256","name":"pubKeyY","type":"uint256"},{"internalType":"uint32","name":"nonce","type":"uint32"},{"internalType":"uint96","name":"balance","type":"uint96"},{"internalType":"uint256","name":"tradeHistoryRoot","type":"uint256"},{"internalType":"uint256[30]","name":"accountPath","type":"uint256[30]"},{"internalType":"uint256[12]","name":"balancePath","type":"uint256[12]"}],"name":"withdrawFromMerkleTree","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint8","name":"blockType","type":"uint8"},{"internalType":"uint16","name":"blockSize","type":"uint16"},{"internalType":"uint8","name":"blockVersion","type":"uint8"},{"internalType":"bytes","name":"","type":"bytes"},{"internalType":"bytes","name":"offchainData","type":"bytes"}],"name":"commitBlock","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"name":"deposit","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"pubKeyX","type":"uint256"},{"internalType":"uint256","name":"pubKeyY","type":"uint256"},{"internalType":"uint32","name":"nonce","type":"uint32"},{"internalType":"uint96","name":"balance","type":"uint96"},{"internalType":"uint256","name":"tradeHistoryRoot","type":"uint256"},{"internalType":"uint256[30]","name":"accountPath","type":"uint256[30]"},{"internalType":"uint256[12]","name":"balancePath","type":"uint256[12]"}],"name":"withdrawFromMerkleTreeFor","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getProtocolFeeValues","outputs":[{"internalType":"uint32","name":"timestamp","type":"uint32"},{"internalType":"uint8","name":"takerFeeBips","type":"uint8"},{"internalType":"uint8","name":"makerFeeBips","type":"uint8"},{"internalType":"uint8","name":"previousTakerFeeBips","type":"uint8"},{"internalType":"uint8","name":"previousMakerFeeBips","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getTotalTimeInMaintenanceSeconds","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"claimOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"version","outputs":[{"internalType":"string","name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"tokenAddress","type":"address"}],"name":"getTokenID","outputs":[{"internalType":"uint16","name":"","type":"uint16"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"index","type":"uint256"}],"name":"getDepositRequest","outputs":[{"internalType":"bytes32","name":"accumulatedHash","type":"bytes32"},{"internalType":"uint256","name":"accumulatedFee","type":"uint256"},{"internalType":"uint32","name":"timestamp","type":"uint32"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"pubKeyX","type":"uint256"},{"internalType":"uint256","name":"pubKeyY","type":"uint256"},{"internalType":"address","name":"token","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"},{"internalType":"bytes","name":"permission","type":"bytes"}],"name":"updateAccountAndDeposit","outputs":[{"internalType":"uint24","name":"accountID","type":"uint24"},{"internalType":"bool","name":"isAccountNew","type":"bool"},{"internalType":"bool","name":"isAccountUpdated","type":"bool"}],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"_accountCreationFeeETH","type":"uint256"},{"internalType":"uint256","name":"_accountUpdateFeeETH","type":"uint256"},{"internalType":"uint256","name":"_depositFeeETH","type":"uint256"},{"internalType":"uint256","name":"_withdrawalFeeETH","type":"uint256"}],"name":"setFees","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[],"name":"renounceOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"tokenAddress","type":"address"}],"name":"enableTokenDeposit","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getBlockHeight","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getNumDepositRequestsProcessed","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getExchangeCreationTimestamp","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getRemainingDowntime","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"name":"withdraw","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"isInWithdrawalMode","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"durationMinutes","type":"uint256"}],"name":"startOrContinueMaintenanceMode","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[],"name":"burnExchangeStake","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getConstants","outputs":[{"internalType":"uint256[20]","name":"","type":"uint256[20]"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[],"name":"getNumAccounts","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"withdrawProtocolFeeStake","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"address","name":"tokenAddress","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"name":"depositTo","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"depositIdx","type":"uint256"}],"name":"withdrawFromDepositRequest","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getNumAvailableDepositSlots","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address payable","name":"_operator","type":"address"}],"name":"setOperator","outputs":[{"internalType":"address payable","name":"","type":"address"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"blockIdx","type":"uint256"},{"internalType":"address payable","name":"feeRecipient","type":"address"}],"name":"withdrawBlockFee","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"isShutdown","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getRequestStats","outputs":[{"internalType":"uint256","name":"numDepositRequestsProcessed","type":"uint256"},{"internalType":"uint256","name":"numAvailableDepositSlots","type":"uint256"},{"internalType":"uint256","name":"numWithdrawalRequestsProcessed","type":"uint256"},{"internalType":"uint256","name":"numAvailableWithdrawalSlots","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"_loopringAddress","type":"address"},{"internalType":"address","name":"_owner","type":"address"},{"internalType":"uint256","name":"_id","type":"uint256"},{"internalType":"address payable","name":"_operator","type":"address"},{"internalType":"bool","name":"_onchainDataAvailability","type":"bool"}],"name":"initialize","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"index","type":"uint256"}],"name":"getWithdrawRequest","outputs":[{"internalType":"bytes32","name":"accumulatedHash","type":"bytes32"},{"internalType":"uint256","name":"accumulatedFee","type":"uint256"},{"internalType":"uint32","name":"timestamp","type":"uint32"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"tokenAddress","type":"address"}],"name":"withdrawTokenNotOwnedByUsers","outputs":[{"internalType":"uint256","name":"amount","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"recipient","type":"address"}],"name":"withdrawExchangeStake","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[],"name":"stopMaintenanceMode","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getFees","outputs":[{"internalType":"uint256","name":"_accountCreationFeeETH","type":"uint256"},{"internalType":"uint256","name":"_accountUpdateFeeETH","type":"uint256"},{"internalType":"uint256","name":"_depositFeeETH","type":"uint256"},{"internalType":"uint256","name":"_withdrawalFeeETH","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getNumAvailableWithdrawalSlots","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getNumWithdrawalRequestsProcessed","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"pendingOwner","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"blockIdx","type":"uint256"},{"internalType":"uint256","name":"maxNumWithdrawals","type":"uint256"}],"name":"distributeWithdrawals","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"pubKeyX","type":"uint256"},{"internalType":"uint256","name":"pubKeyY","type":"uint256"},{"internalType":"bytes","name":"permission","type":"bytes"}],"name":"createOrUpdateAccount","outputs":[{"internalType":"uint24","name":"accountID","type":"uint24"},{"internalType":"bool","name":"isAccountNew","type":"bool"},{"internalType":"bool","name":"isAccountUpdated","type":"bool"}],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"internalType":"uint16","name":"tokenID","type":"uint16"}],"name":"getTokenAddress","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"blockIdx","type":"uint256"},{"internalType":"uint256","name":"slotIdx","type":"uint256"}],"name":"withdrawFromApprovedWithdrawal","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getExchangeStake","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"merkleRoot","type":"uint256"},{"internalType":"uint24","name":"accountID","type":"uint24"},{"internalType":"uint16","name":"tokenID","type":"uint16"},{"internalType":"uint256","name":"pubKeyX","type":"uint256"},{"internalType":"uint256","name":"pubKeyY","type":"uint256"},{"internalType":"uint32","name":"nonce","type":"uint32"},{"internalType":"uint96","name":"balance","type":"uint96"},{"internalType":"uint256","name":"tradeHistoryRoot","type":"uint256"},{"internalType":"uint256[30]","name":"accountPath","type":"uint256[30]"},{"internalType":"uint256[12]","name":"balancePath","type":"uint256[12]"}],"name":"isAccountBalanceCorrect","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"owner","type":"address"}],"name":"getAccount","outputs":[{"internalType":"uint24","name":"accountID","type":"uint24"},{"internalType":"uint256","name":"pubKeyX","type":"uint256"},{"internalType":"uint256","name":"pubKeyY","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"shutdown","outputs":[{"internalType":"bool","name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"uint24","name":"id","type":"uint24"},{"indexed":false,"internalType":"uint256","name":"pubKeyX","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"pubKeyY","type":"uint256"}],"name":"AccountCreated","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"uint24","name":"id","type":"uint24"},{"indexed":false,"internalType":"uint256","name":"pubKeyX","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"pubKeyY","type":"uint256"}],"name":"AccountUpdated","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"token","type":"address"},{"indexed":true,"internalType":"uint16","name":"tokenId","type":"uint16"}],"name":"TokenRegistered","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"exchangeId","type":"uint256"},{"indexed":false,"internalType":"address","name":"oldOperator","type":"address"},{"indexed":false,"internalType":"address","name":"newOperator","type":"address"}],"name":"OperatorChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"exchangeId","type":"uint256"},{"indexed":false,"internalType":"address","name":"oldAddressWhitelist","type":"address"},{"indexed":false,"internalType":"address","name":"newAddressWhitelist","type":"address"}],"name":"AddressWhitelistChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"exchangeId","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"accountCreationFeeETH","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"accountUpdateFeeETH","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"depositFeeETH","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"withdrawalFeeETH","type":"uint256"}],"name":"FeesUpdated","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"timestamp","type":"uint256"}],"name":"Shutdown","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"blockIdx","type":"uint256"},{"indexed":true,"internalType":"bytes32","name":"publicDataHash","type":"bytes32"}],"name":"BlockCommitted","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"blockIdx","type":"uint256"}],"name":"BlockVerified","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"blockIdx","type":"uint256"}],"name":"BlockFinalized","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"blockIdx","type":"uint256"}],"name":"Revert","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"depositIdx","type":"uint256"},{"indexed":true,"internalType":"uint24","name":"accountID","type":"uint24"},{"indexed":true,"internalType":"uint16","name":"tokenID","type":"uint16"},{"indexed":false,"internalType":"uint96","name":"amount","type":"uint96"},{"indexed":false,"internalType":"uint256","name":"pubKeyX","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"pubKeyY","type":"uint256"}],"name":"DepositRequested","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"blockIdx","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"BlockFeeWithdrawn","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"withdrawalIdx","type":"uint256"},{"indexed":true,"internalType":"uint24","name":"accountID","type":"uint24"},{"indexed":true,"internalType":"uint16","name":"tokenID","type":"uint16"},{"indexed":false,"internalType":"uint96","name":"amount","type":"uint96"}],"name":"WithdrawalRequested","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint24","name":"accountID","type":"uint24"},{"indexed":true,"internalType":"uint16","name":"tokenID","type":"uint16"},{"indexed":false,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint96","name":"amount","type":"uint96"}],"name":"WithdrawalCompleted","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint24","name":"accountID","type":"uint24"},{"indexed":true,"internalType":"uint16","name":"tokenID","type":"uint16"},{"indexed":false,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint96","name":"amount","type":"uint96"}],"name":"WithdrawalFailed","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint8","name":"takerFeeBips","type":"uint8"},{"indexed":false,"internalType":"uint8","name":"makerFeeBips","type":"uint8"},{"indexed":false,"internalType":"uint8","name":"previousTakerFeeBips","type":"uint8"},{"indexed":false,"internalType":"uint8","name":"previousMakerFeeBips","type":"uint8"}],"name":"ProtocolFeesUpdated","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"sender","type":"address"},{"indexed":false,"internalType":"address","name":"token","type":"address"},{"indexed":false,"internalType":"address","name":"feeVault","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"TokenNotOwnedByUsersWithdrawn","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"clone","type":"address"}],"name":"Cloned","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"}]`

		paramInput := trigger.Input{
			ParameterType:  "uint16",
			ParameterValue: params[0],
		}
		methodId, err := trigger.EncodeMethod(methodName, abi, []trigger.Input{paramInput})
		if err != nil {
			return err.Error()
		}
		lastBlock, err := config.CliMain.EthBlockNumber()
		if err != nil {
			return err.Error()
		}
		rawData, err := trigger.MakeEthRpcCall(config.CliMain, "0x944644Ea989Ec64c2Ab9eF341D383cEf586A5777", methodId, lastBlock)
		if err != nil {
			return err.Error()
		}

		result, err := abidec.DecodeParamsIntoList(strings.TrimPrefix(rawData, "0x"), abi, "getTokenAddress")
		if err != nil {
			return err.Error()
		}
		if len(result) == 1 {
			return common.HexToAddress(fmt.Sprintf("%x", result[0])).String()
		}
		return "aaaand it's broken"
	}
	return "only loopring supported now"
}

type TokenAPI struct {
	tokenAddToPrice *cache.Cache
	httpCli         *http.Client
}

// package-level singleton accessed through GetTokenAPI()
var tokenApi = &TokenAPI{
	tokenAddToPrice: cache.New(5*time.Minute, 5*time.Minute),
	httpCli:         &http.Client{},
}

func GetTokenAPI() *TokenAPI {
	return tokenApi
}

func (t TokenAPI) addressToFiat(tokenAddress, fiatCurrency string) (float32, error) {
	tokenAddress = strings.ToLower(tokenAddress)
	fiatCurrency = strings.ToLower(fiatCurrency)
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/token_price/ethereum?contract_addresses=%s&vs_currencies=%s", tokenAddress, fiatCurrency)

	key := tokenAddress + fiatCurrency

	// try cache first
	val, found := t.tokenAddToPrice.Get(key)
	if found {
		return val.(float32), nil
	}

	// no luck, call Coingecko API
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
		t.tokenAddToPrice.Set(key, val, 5*time.Minute)
		return val.(float32), nil
	}
	return 0, fmt.Errorf("unexpected json response")
}
