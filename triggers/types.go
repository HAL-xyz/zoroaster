package trigger

import "github.com/onrik/ethrpc"

type ZTransaction struct {
	BlockTimestamp int
	DecodedInput   string
	Tx             *ethrpc.Transaction
}

type Match struct {
	Tg  *Trigger
	ZTx *ZTransaction
}

type ActionEvent struct {
	ZTx     *ZTransaction
	Actions []string
}
