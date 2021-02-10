package rpc

import (
	"github.com/HAL-xyz/ethrpc"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

// An interface for eth rpc clients, as used within Zoroaster
type IEthRpc interface {
	EthGetLogs(params ethrpc.FilterParams) ([]ethrpc.Log, error)
	EthGetBlockByNumber(number int, withTransactions bool) (*ethrpc.Block, error)
	EthGetTransactionByHash(hash string) (*ethrpc.Transaction, error)
	EthBlockNumber() (int, error)
	URL() string
	EthCall(transaction ethrpc.T, tag string) (string, error)
	ResetCounterAndLogStats(blockNo int)
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
func New(ethCli *ethrpc.EthRPC, label string) *ZoroRPC {
	return &ZoroRPC{
		cli:   ethCli,
		label: label,
		calls: 0,
		cache: cache.New(5*time.Minute, 5*time.Minute),
	}
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

func (z *ZoroRPC) EthGetTransactionByHash(hash string) (*ethrpc.Transaction, error) {
	z.increaseCounterByOne()
	return z.cli.EthGetTransactionByHash(hash)
}

func (z *ZoroRPC) EthGetBlockByNumber(number int, withTransactions bool) (*ethrpc.Block, error) {
	z.increaseCounterByOne()
	return z.cli.EthGetBlockByNumber(number, withTransactions)
}

func (z *ZoroRPC) EthGetLogs(params ethrpc.FilterParams) ([]ethrpc.Log, error) {
	z.increaseCounterByOne()
	return z.cli.EthGetLogs(params)
}

func (z *ZoroRPC) EthBlockNumber() (int, error) {
	z.increaseCounterByOne()
	return z.cli.EthBlockNumber()
}

func (z *ZoroRPC) URL() string {
	return z.cli.URL()
}

func (z *ZoroRPC) EthCall(transaction ethrpc.T, tag string) (string, error) {
	key := transaction.To + transaction.Data + tag
	val, found := z.cache.Get(key)
	if found {
		return val.(string), nil
	}

	z.increaseCounterByOne()
	res, err := z.cli.EthCall(transaction, tag)
	if err == nil {
		z.cache.Set(key, res, cache.DefaultExpiration)
	}
	return res, err
}
