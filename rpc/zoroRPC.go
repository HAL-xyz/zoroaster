package rpc

import (
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
)

// An interface for eth rpc clients, as used within Zoroaster
type IEthRpc interface {
	EthGetLogs(params ethrpc.FilterParams) ([]ethrpc.Log, error)
	EthGetBlockByNumber(number int, withTransactions bool) (*ethrpc.Block, error)
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
}

// Returns a new ZoroRPC client
func New(ethCli *ethrpc.EthRPC, label string) *ZoroRPC {
	return &ZoroRPC{
		cli:   ethCli,
		label: label,
		calls: 0,
	}
}

func (z *ZoroRPC) resetCounter() {
	z.calls = 0
}

func (z *ZoroRPC) increaseCounterByOne() {
	z.calls += 1
}

func (z *ZoroRPC) ResetCounterAndLogStats(blockNo int) {
	if z.label == "BlocksPoller" && z.calls <= 1 {
		return
	}
	log.Infof("RPCStats: %s made %d calls for block %d\n", z.label, z.calls, blockNo)
	z.resetCounter()
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
	z.increaseCounterByOne()
	return z.cli.EthCall(transaction, tag)
}
