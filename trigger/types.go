package trigger

import "github.com/onrik/ethrpc"

type ZTransaction struct {
	BlockTimestamp int
	DecodedFnArgs  *string `json:"DecodedFnArgs,omitempty"`
	DecodedFnName  *string `json:"DecodedFnName,omitempty"`
	Tx             *ethrpc.Transaction
}

type Match struct {
	Tg      *Trigger
	ZTx     *ZTransaction
	MatchId int
}

type CnMatch struct {
	BlockNo int
	TgId    int
	Value   string
}

type Outcome struct {
	Outcome string
	Payload string
}
