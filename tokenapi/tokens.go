package tokenapi

import (
	"encoding/json"
	"fmt"
	"github.com/HAL-xyz/ethrpc"
	"github.com/HAL-xyz/zoroaster/abidec"
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/HAL-xyz/zoroaster/rpc"
	"github.com/ethereum/go-ethereum/common"
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
}

// package-level singleton accessed through GetTokenAPI()
// some day it would be nice to pass it explicitly as a dependency of the templating system
var tokenApi = &TokenAPI{
	tokenAddToPrice: cache.New(5*time.Minute, 5*time.Minute),
	httpCli:         &http.Client{},
	rpcCli:          rpc.New(ethrpc.New(config.Zconf.EthNode), "templating client"),
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
	return t.callERC20(address, "0x313ce567", "decimals")
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

// TODO: this is a horrible hack to get LoopRing working ASAP.
// In the future we'll do this properly to make it generic and work with any contract.
func (t TokenAPI) Call(contractName, methodName string, params ...string) string {
	if contractName == "LoopringDex" {
		if len(params) != 1 {
			return "no no no"
		}
		abi := `[{"constant":true,"inputs":[],"name":"getLRCFeeForRegisteringOneMoreToken","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"blockIdx","type":"uint256"}],"name":"getBlock","outputs":[{"internalType":"bytes32","name":"merkleRoot","type":"bytes32"},{"internalType":"bytes32","name":"publicDataHash","type":"bytes32"},{"internalType":"uint8","name":"blockState","type":"uint8"},{"internalType":"uint8","name":"blockType","type":"uint8"},{"internalType":"uint16","name":"blockSize","type":"uint16"},{"internalType":"uint32","name":"timestamp","type":"uint32"},{"internalType":"uint32","name":"numDepositRequestsCommitted","type":"uint32"},{"internalType":"uint32","name":"numWithdrawalRequestsCommitted","type":"uint32"},{"internalType":"bool","name":"blockFeeWithdrawn","type":"bool"},{"internalType":"uint16","name":"numWithdrawalsDistributed","type":"uint16"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"tokenAddress","type":"address"}],"name":"registerToken","outputs":[{"internalType":"uint16","name":"","type":"uint16"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[],"name":"clone","outputs":[{"internalType":"address","name":"cloneAddress","type":"address"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"durationMinutes","type":"uint256"}],"name":"getDowntimeCostLRC","outputs":[{"internalType":"uint256","name":"costLRC","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"_addressWhitelist","type":"address"}],"name":"setAddressWhitelist","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256[]","name":"blockIndices","type":"uint256[]"},{"internalType":"uint256[]","name":"proofs","type":"uint256[]"}],"name":"verifyBlocks","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getNumBlocksFinalized","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"isInMaintenance","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"tokenAddress","type":"address"}],"name":"disableTokenDeposit","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"blockIdx","type":"uint256"}],"name":"revertBlock","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"genesisBlockHash","outputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"token","type":"address"}],"name":"withdrawProtocolFees","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"pubKeyX","type":"uint256"},{"internalType":"uint256","name":"pubKeyY","type":"uint256"},{"internalType":"uint32","name":"nonce","type":"uint32"},{"internalType":"uint96","name":"balance","type":"uint96"},{"internalType":"uint256","name":"tradeHistoryRoot","type":"uint256"},{"internalType":"uint256[30]","name":"accountPath","type":"uint256[30]"},{"internalType":"uint256[12]","name":"balancePath","type":"uint256[12]"}],"name":"withdrawFromMerkleTree","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint8","name":"blockType","type":"uint8"},{"internalType":"uint16","name":"blockSize","type":"uint16"},{"internalType":"uint8","name":"blockVersion","type":"uint8"},{"internalType":"bytes","name":"","type":"bytes"},{"internalType":"bytes","name":"offchainData","type":"bytes"}],"name":"commitBlock","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"name":"deposit","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"pubKeyX","type":"uint256"},{"internalType":"uint256","name":"pubKeyY","type":"uint256"},{"internalType":"uint32","name":"nonce","type":"uint32"},{"internalType":"uint96","name":"balance","type":"uint96"},{"internalType":"uint256","name":"tradeHistoryRoot","type":"uint256"},{"internalType":"uint256[30]","name":"accountPath","type":"uint256[30]"},{"internalType":"uint256[12]","name":"balancePath","type":"uint256[12]"}],"name":"withdrawFromMerkleTreeFor","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getProtocolFeeValues","outputs":[{"internalType":"uint32","name":"timestamp","type":"uint32"},{"internalType":"uint8","name":"takerFeeBips","type":"uint8"},{"internalType":"uint8","name":"makerFeeBips","type":"uint8"},{"internalType":"uint8","name":"previousTakerFeeBips","type":"uint8"},{"internalType":"uint8","name":"previousMakerFeeBips","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getTotalTimeInMaintenanceSeconds","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"claimOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"version","outputs":[{"internalType":"string","name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"tokenAddress","type":"address"}],"name":"getTokenID","outputs":[{"internalType":"uint16","name":"","type":"uint16"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"index","type":"uint256"}],"name":"getDepositRequest","outputs":[{"internalType":"bytes32","name":"accumulatedHash","type":"bytes32"},{"internalType":"uint256","name":"accumulatedFee","type":"uint256"},{"internalType":"uint32","name":"timestamp","type":"uint32"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"pubKeyX","type":"uint256"},{"internalType":"uint256","name":"pubKeyY","type":"uint256"},{"internalType":"address","name":"token","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"},{"internalType":"bytes","name":"permission","type":"bytes"}],"name":"updateAccountAndDeposit","outputs":[{"internalType":"uint24","name":"accountID","type":"uint24"},{"internalType":"bool","name":"isAccountNew","type":"bool"},{"internalType":"bool","name":"isAccountUpdated","type":"bool"}],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"_accountCreationFeeETH","type":"uint256"},{"internalType":"uint256","name":"_accountUpdateFeeETH","type":"uint256"},{"internalType":"uint256","name":"_depositFeeETH","type":"uint256"},{"internalType":"uint256","name":"_withdrawalFeeETH","type":"uint256"}],"name":"setFees","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[],"name":"renounceOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"tokenAddress","type":"address"}],"name":"enableTokenDeposit","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getBlockHeight","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getNumDepositRequestsProcessed","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getExchangeCreationTimestamp","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getRemainingDowntime","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"name":"withdraw","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"isInWithdrawalMode","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"durationMinutes","type":"uint256"}],"name":"startOrContinueMaintenanceMode","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[],"name":"burnExchangeStake","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getConstants","outputs":[{"internalType":"uint256[20]","name":"","type":"uint256[20]"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[],"name":"getNumAccounts","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"withdrawProtocolFeeStake","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"address","name":"tokenAddress","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"name":"depositTo","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"depositIdx","type":"uint256"}],"name":"withdrawFromDepositRequest","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getNumAvailableDepositSlots","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address payable","name":"_operator","type":"address"}],"name":"setOperator","outputs":[{"internalType":"address payable","name":"","type":"address"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"blockIdx","type":"uint256"},{"internalType":"address payable","name":"feeRecipient","type":"address"}],"name":"withdrawBlockFee","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"isShutdown","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getRequestStats","outputs":[{"internalType":"uint256","name":"numDepositRequestsProcessed","type":"uint256"},{"internalType":"uint256","name":"numAvailableDepositSlots","type":"uint256"},{"internalType":"uint256","name":"numWithdrawalRequestsProcessed","type":"uint256"},{"internalType":"uint256","name":"numAvailableWithdrawalSlots","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"_loopringAddress","type":"address"},{"internalType":"address","name":"_owner","type":"address"},{"internalType":"uint256","name":"_id","type":"uint256"},{"internalType":"address payable","name":"_operator","type":"address"},{"internalType":"bool","name":"_onchainDataAvailability","type":"bool"}],"name":"initialize","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"index","type":"uint256"}],"name":"getWithdrawRequest","outputs":[{"internalType":"bytes32","name":"accumulatedHash","type":"bytes32"},{"internalType":"uint256","name":"accumulatedFee","type":"uint256"},{"internalType":"uint32","name":"timestamp","type":"uint32"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"tokenAddress","type":"address"}],"name":"withdrawTokenNotOwnedByUsers","outputs":[{"internalType":"uint256","name":"amount","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"recipient","type":"address"}],"name":"withdrawExchangeStake","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[],"name":"stopMaintenanceMode","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getFees","outputs":[{"internalType":"uint256","name":"_accountCreationFeeETH","type":"uint256"},{"internalType":"uint256","name":"_accountUpdateFeeETH","type":"uint256"},{"internalType":"uint256","name":"_depositFeeETH","type":"uint256"},{"internalType":"uint256","name":"_withdrawalFeeETH","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getNumAvailableWithdrawalSlots","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getNumWithdrawalRequestsProcessed","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"pendingOwner","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"blockIdx","type":"uint256"},{"internalType":"uint256","name":"maxNumWithdrawals","type":"uint256"}],"name":"distributeWithdrawals","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"pubKeyX","type":"uint256"},{"internalType":"uint256","name":"pubKeyY","type":"uint256"},{"internalType":"bytes","name":"permission","type":"bytes"}],"name":"createOrUpdateAccount","outputs":[{"internalType":"uint24","name":"accountID","type":"uint24"},{"internalType":"bool","name":"isAccountNew","type":"bool"},{"internalType":"bool","name":"isAccountUpdated","type":"bool"}],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"internalType":"uint16","name":"tokenID","type":"uint16"}],"name":"getTokenAddress","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"blockIdx","type":"uint256"},{"internalType":"uint256","name":"slotIdx","type":"uint256"}],"name":"withdrawFromApprovedWithdrawal","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getExchangeStake","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"merkleRoot","type":"uint256"},{"internalType":"uint24","name":"accountID","type":"uint24"},{"internalType":"uint16","name":"tokenID","type":"uint16"},{"internalType":"uint256","name":"pubKeyX","type":"uint256"},{"internalType":"uint256","name":"pubKeyY","type":"uint256"},{"internalType":"uint32","name":"nonce","type":"uint32"},{"internalType":"uint96","name":"balance","type":"uint96"},{"internalType":"uint256","name":"tradeHistoryRoot","type":"uint256"},{"internalType":"uint256[30]","name":"accountPath","type":"uint256[30]"},{"internalType":"uint256[12]","name":"balancePath","type":"uint256[12]"}],"name":"isAccountBalanceCorrect","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"owner","type":"address"}],"name":"getAccount","outputs":[{"internalType":"uint24","name":"accountID","type":"uint24"},{"internalType":"uint256","name":"pubKeyX","type":"uint256"},{"internalType":"uint256","name":"pubKeyY","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"shutdown","outputs":[{"internalType":"bool","name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"uint24","name":"id","type":"uint24"},{"indexed":false,"internalType":"uint256","name":"pubKeyX","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"pubKeyY","type":"uint256"}],"name":"AccountCreated","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"uint24","name":"id","type":"uint24"},{"indexed":false,"internalType":"uint256","name":"pubKeyX","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"pubKeyY","type":"uint256"}],"name":"AccountUpdated","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"token","type":"address"},{"indexed":true,"internalType":"uint16","name":"tokenId","type":"uint16"}],"name":"TokenRegistered","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"exchangeId","type":"uint256"},{"indexed":false,"internalType":"address","name":"oldOperator","type":"address"},{"indexed":false,"internalType":"address","name":"newOperator","type":"address"}],"name":"OperatorChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"exchangeId","type":"uint256"},{"indexed":false,"internalType":"address","name":"oldAddressWhitelist","type":"address"},{"indexed":false,"internalType":"address","name":"newAddressWhitelist","type":"address"}],"name":"AddressWhitelistChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"exchangeId","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"accountCreationFeeETH","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"accountUpdateFeeETH","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"depositFeeETH","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"withdrawalFeeETH","type":"uint256"}],"name":"FeesUpdated","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"timestamp","type":"uint256"}],"name":"Shutdown","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"blockIdx","type":"uint256"},{"indexed":true,"internalType":"bytes32","name":"publicDataHash","type":"bytes32"}],"name":"BlockCommitted","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"blockIdx","type":"uint256"}],"name":"BlockVerified","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"blockIdx","type":"uint256"}],"name":"BlockFinalized","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"blockIdx","type":"uint256"}],"name":"Revert","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"depositIdx","type":"uint256"},{"indexed":true,"internalType":"uint24","name":"accountID","type":"uint24"},{"indexed":true,"internalType":"uint16","name":"tokenID","type":"uint16"},{"indexed":false,"internalType":"uint96","name":"amount","type":"uint96"},{"indexed":false,"internalType":"uint256","name":"pubKeyX","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"pubKeyY","type":"uint256"}],"name":"DepositRequested","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"blockIdx","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"BlockFeeWithdrawn","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"withdrawalIdx","type":"uint256"},{"indexed":true,"internalType":"uint24","name":"accountID","type":"uint24"},{"indexed":true,"internalType":"uint16","name":"tokenID","type":"uint16"},{"indexed":false,"internalType":"uint96","name":"amount","type":"uint96"}],"name":"WithdrawalRequested","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint24","name":"accountID","type":"uint24"},{"indexed":true,"internalType":"uint16","name":"tokenID","type":"uint16"},{"indexed":false,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint96","name":"amount","type":"uint96"}],"name":"WithdrawalCompleted","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint24","name":"accountID","type":"uint24"},{"indexed":true,"internalType":"uint16","name":"tokenID","type":"uint16"},{"indexed":false,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint96","name":"amount","type":"uint96"}],"name":"WithdrawalFailed","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint8","name":"takerFeeBips","type":"uint8"},{"indexed":false,"internalType":"uint8","name":"makerFeeBips","type":"uint8"},{"indexed":false,"internalType":"uint8","name":"previousTakerFeeBips","type":"uint8"},{"indexed":false,"internalType":"uint8","name":"previousMakerFeeBips","type":"uint8"}],"name":"ProtocolFeesUpdated","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"sender","type":"address"},{"indexed":false,"internalType":"address","name":"token","type":"address"},{"indexed":false,"internalType":"address","name":"feeVault","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"TokenNotOwnedByUsersWithdrawn","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"clone","type":"address"}],"name":"Cloned","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"}]`

		paramInput := Input{
			ParameterType:  "uint16",
			ParameterValue: params[0],
		}
		methodId, err := t.EncodeMethod(methodName, abi, []Input{paramInput})
		if err != nil {
			return err.Error()
		}
		lastBlock, err := t.rpcCli.EthBlockNumber()
		if err != nil {
			return err.Error()
		}
		rawData, err := t.MakeEthRpcCall("0x944644Ea989Ec64c2Ab9eF341D383cEf586A5777", methodId, lastBlock)
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
