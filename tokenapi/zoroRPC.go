package tokenapi

import (
	"fmt"
	"github.com/HAL-xyz/ethrpc"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

// An interface for eth rpc clients, as used within Zoroaster
type IEthRpc interface {
	EthGetLogsByHash(blockHash string) ([]ethrpc.Log, error)
	EthGetLogsByNumber(blockNo int, address string) ([]ethrpc.Log, error)
	EthGetBlockByNumber(number int, withTransactions bool) (*ethrpc.Block, error)
	EthGetTransactionByHash(hash string) (*ethrpc.Transaction, error)
	EthBlockNumber() (int, error)
	URL() string
	EthCall(transaction ethrpc.T, tag string) (string, error)
	ResetCounterAndLogStats(blockNo int)
	GetLabel() string
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
	ethClient := ethrpc.New(node)

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

func (z *ZoroRPC) EthGetTransactionByHash(hash string) (*ethrpc.Transaction, error) {
	z.increaseCounterByOne()
	return z.cli.EthGetTransactionByHash(hash)
}

func (z *ZoroRPC) EthGetBlockByNumber(number int, withTransactions bool) (*ethrpc.Block, error) {
	z.increaseCounterByOne()
	return z.cli.EthGetBlockByNumber(number, withTransactions)
}

// Lookups using only block hash are much faster than using block numbers and/or addresses
func (z *ZoroRPC) EthGetLogsByHash(blockHash string) ([]ethrpc.Log, error) {
	filter := ethrpc.FilterParams{
		BlockHash: blockHash,
	}
	z.increaseCounterByOne()
	return z.cli.EthGetLogs(filter)
}

// // Lookups using block numbers and/or addresses are slower, but useful for testing
func (z *ZoroRPC) EthGetLogsByNumber(blockNo int, address string) ([]ethrpc.Log, error) {
	filter := ethrpc.FilterParams{
		FromBlock: fmt.Sprintf("0x%x", blockNo),
		ToBlock:   fmt.Sprintf("0x%x", blockNo),
		Address:   []string{address},
	}
	z.increaseCounterByOne()
	return z.cli.EthGetLogs(filter)
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
