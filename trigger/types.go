package trigger

import "github.com/onrik/ethrpc"

type ZTransaction struct {
	BlockTimestamp int
	DecodedFnArgs  *string `json:"DecodedFnArgs,omitempty"`
	DecodedFnName  *string `json:"DecodedFnName,omitempty"`
	Tx             *ethrpc.Transaction
}

// IMatch is an interface to fake IMatch as a sum type {TxMatch, CnMatch}
type IMatch interface {
	isMatch()
}

type TxMatch struct {
	MatchId int
	Tg      *Trigger
	ZTx     *ZTransaction
}

// Implements IMatch interface
func (TxMatch) isMatch() {}

type CnMatch struct {
	MatchId        int
	BlockNo        int
	TgId           int
	TgUserId       int
	Value          string
	AllValues      string
	BlockTimestamp int
}

// Implements IMatch interface
func (CnMatch) isMatch() {}

type Outcome struct {
	Outcome string
	Payload string
}
