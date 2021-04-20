package tokenapi

import (
	"fmt"
	"github.com/HAL-xyz/ethrpc"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

// An interface for eth rpc clients, as used within Zoroaster
type IEthRpc interface {
	EthGetLogsByHash(blockHash string) ([]ethrpc.Log, error)
	EthGetLogsByNumber(blockNo int, address string) ([]ethrpc.Log, error)
	EthGetBlockByNumber(number int, withTransactions bool) (*ethrpc.Block, error)
	EthBlockNumber() (int, error)
	ResetCounterAndLogStats(blockNo int)
	GetLabel() string
	EncodeMethod(methodName, cntABI string, inputs []Input) (string, error)
	MakeEthRpcCall(cntAddress, data string, blockNumber int) (string, error)
}

// A wrapper for the ethrpc.EthRPC client.
type ZoroRPC struct {
	cli   *ethrpc.EthRPC
	label string
	calls int
	cache *cache.Cache
	sync.Mutex
}

// Returns a new ZoroRPC client
func NewZRPC(node, label string) *ZoroRPC {
	ethClient := ethrpc.New(node, ethrpc.WithHttpClient(&http.Client{Timeout: 30 * time.Second}))

	return &ZoroRPC{
		cli:   ethClient,
		label: label,
		calls: 0,
		cache: cache.New(5*time.Minute, 5*time.Minute),
	}
}

func (z *ZoroRPC) GetLabel() string {
	return z.label
}

func (z *ZoroRPC) resetCounter() {
	z.Lock()
	z.calls = 0
	z.Unlock()
}

func (z *ZoroRPC) increaseCounterByOne() {
	z.Lock()
	z.calls += 1
	z.Unlock()
}

func (z *ZoroRPC) ResetCounterAndLogStats(blockNo int) {
	if z.label == "BlocksPoller" && z.calls <= 1 {
		return
	}
	log.Infof("RPCStats: %s made %d calls for block %d; cache size is %d", z.label, z.calls, blockNo, z.cache.ItemCount())
	z.resetCounter()
}

func (z *ZoroRPC) EthGetBlockByNumber(number int, withTransactions bool) (*ethrpc.Block, error) {
	key := "get_block" + fmt.Sprintf("%d", number)
	val, found := z.cache.Get(key)
	if found {
		return val.(*ethrpc.Block), nil
	}

	z.increaseCounterByOne()
	res, err := z.cli.EthGetBlockByNumber(number, withTransactions)
	if err == nil {
		z.cache.Set(key, res, cache.DefaultExpiration)
	}
	return res, err
}

// Lookups using only block hash are much faster than using block numbers and/or addresses
func (z *ZoroRPC) EthGetLogsByHash(blockHash string) ([]ethrpc.Log, error) {
	key := "get_logs" + blockHash
	val, found := z.cache.Get(key)
	if found {
		return val.([]ethrpc.Log), nil
	}

	filter := ethrpc.FilterParams{
		BlockHash: blockHash,
	}
	z.increaseCounterByOne()

	res, err := z.cli.EthGetLogs(filter)
	if err == nil {
		z.cache.Set(key, res, cache.DefaultExpiration)
	}
	return res, err
}

// Lookups using block numbers and/or addresses are slower, but useful for testing
func (z *ZoroRPC) EthGetLogsByNumber(blockNo int, address string) ([]ethrpc.Log, error) {
	key := "get_logs" + fmt.Sprintf("%d", blockNo)
	val, found := z.cache.Get(key)
	if found {
		return val.([]ethrpc.Log), nil
	}

	filter := ethrpc.FilterParams{
		FromBlock: fmt.Sprintf("0x%x", blockNo),
		ToBlock:   fmt.Sprintf("0x%x", blockNo),
		Address:   []string{address},
	}
	z.increaseCounterByOne()
	res, err := z.cli.EthGetLogs(filter)
	if err == nil {
		z.cache.Set(key, res, cache.DefaultExpiration)
	}
	return res, err
}

func (z *ZoroRPC) EthBlockNumber() (int, error) {
	z.increaseCounterByOne()
	return z.cli.EthBlockNumber()
}

func (z *ZoroRPC) MakeEthRpcCall(cntAddress, data string, blockNumber int) (string, error) {
	params := ethrpc.T{
		To:   cntAddress,
		From: "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		Data: data,
	}

	hexBlockNo := fmt.Sprintf("0x%x", blockNumber)

	key := params.To + params.Data + hexBlockNo
	val, found := z.cache.Get(key)
	if found {
		return val.(string), nil
	}

	z.increaseCounterByOne()
	res, err := z.cli.EthCall(params, hexBlockNo)
	if err == nil {
		z.cache.Set(key, res, cache.DefaultExpiration)
	}
	return res, err
}
