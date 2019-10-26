package trigger

import "github.com/onrik/ethrpc"

type IEthRpc interface {
	EthGetLogs(params ethrpc.FilterParams) ([]ethrpc.Log, error)
}
