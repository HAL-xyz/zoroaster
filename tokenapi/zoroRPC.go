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
	MakeEthRpcCall(cntAddress, data string, blockNumber int) (string, error)
}

// A wrapper for the ethrpc.EthRPC client.
type ZoroRPC struct {
	cli   *ethrpc.EthRPC
	label string
	calls int
	cache *cache.Cache
	sync.Mutex
	retries int
}

// Returns a new ZoroRPC client
func NewZRPC(node, label string, options ...func(rpc *ZoroRPC)) *ZoroRPC {
	zoroCli := &ZoroRPC{
		cli:     ethrpc.New(node, ethrpc.WithHttpClient(&http.Client{Timeout: 10 * time.Second})),
		label:   label,
		calls:   0,
		cache:   cache.New(5*time.Minute, 5*time.Minute),
		retries: 1,
	}

	for _, opt := range options {
		opt(zoroCli)
	}

	return zoroCli
}

func WithRetries(n int) func(rpc *ZoroRPC) {
	return func(rpc *ZoroRPC) {
		rpc.retries = n
	}
}

func (z *ZoroRPC) cacheGet(key string) (interface{}, bool) {
	z.Lock()
	val, found := z.cache.Get(key)
	z.Unlock()
	return val, found
}

func (z *ZoroRPC) cacheSet(key string, obj interface{}) {
	z.Lock()
	z.cache.Set(key, obj, cache.DefaultExpiration)
	z.Unlock()
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
	var res *ethrpc.Block
	var err error

	key := "get_block" + fmt.Sprintf("%d", number)
	val, found := z.cacheGet(key)
	if found {
		return val.(*ethrpc.Block), nil
	}

	for i := 0; i < z.retries; i++ {
		res, err := z.cli.EthGetBlockByNumber(number, withTransactions)
		z.increaseCounterByOne()
		if err == nil {
			z.cacheSet(key, res)
			return res, nil
		} else {
			log.Warnf("call GetBlockByNumber failed; attempt #%d", i+1)
			time.Sleep(time.Duration(i*i+1) * time.Second)
		}
	}
	return res, err
}

// Lookups using only block hash are much faster than using block numbers and/or addresses
func (z *ZoroRPC) EthGetLogsByHash(blockHash string) ([]ethrpc.Log, error) {
	var res []ethrpc.Log
	var err error

	key := "get_logs" + blockHash
	val, found := z.cacheGet(key)
	if found {
		return val.([]ethrpc.Log), nil
	}

	filter := ethrpc.FilterParams{
		BlockHash: blockHash,
	}

	for i := 0; i < z.retries; i++ {
		res, err = z.cli.EthGetLogs(filter)
		z.increaseCounterByOne()
		if err == nil {
			z.cacheSet(key, res)
			return res, nil
		} else {
			log.Warnf("call GetLogsByHash failed; attempt #%d", i+1)
			time.Sleep(time.Duration(i*i+1) * time.Second)
		}
	}
	return res, err
}

// Lookups using block numbers and/or addresses are slower, but useful for testing
func (z *ZoroRPC) EthGetLogsByNumber(blockNo int, address string) ([]ethrpc.Log, error) {
	var res []ethrpc.Log
	var err error

	key := "get_logs" + fmt.Sprintf("%d", blockNo)
	val, found := z.cacheGet(key)
	if found {
		return val.([]ethrpc.Log), nil
	}

	filter := ethrpc.FilterParams{
		FromBlock: fmt.Sprintf("0x%x", blockNo),
		ToBlock:   fmt.Sprintf("0x%x", blockNo),
		Address:   []string{address},
	}

	for i := 0; i < z.retries; i++ {
		res, err = z.cli.EthGetLogs(filter)
		z.increaseCounterByOne()
		if err == nil {
			z.cacheSet(key, res)
			return res, nil
		} else {
			log.Warnf("call GetLogsByNumber failed; attempt #%d", i+1)
			time.Sleep(time.Duration(i*i+1) * time.Second)
		}
	}
	return res, err
}

func (z *ZoroRPC) EthBlockNumber() (int, error) {
	var res int
	var err error

	for i := 0; i < z.retries; i++ {
		res, err = z.cli.EthBlockNumber()
		z.increaseCounterByOne()
		if err == nil {
			return res, err
		} else {
			timeout := i*i + 1
			log.Warnf("call EthBlockNumber failed; attempt #%d; retrying in %d seconds", i+1, timeout)
			time.Sleep(time.Duration(timeout) * time.Second)
		}
	}
	return res, err
}

func (z *ZoroRPC) MakeEthRpcCall(cntAddress, data string, blockNumber int) (string, error) {
	params := ethrpc.T{
		To:   cntAddress,
		From: "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		Data: data,
	}

	hexBlockNo := fmt.Sprintf("0x%x", blockNumber)

	key := params.To + params.Data + hexBlockNo
	val, found := z.cacheGet(key)
	if found {
		return val.(string), nil
	}

	z.increaseCounterByOne()
	res, err := z.cli.EthCall(params, hexBlockNo)
	if err == nil {
		z.cacheSet(key, res)
	}
	return res, err
}
