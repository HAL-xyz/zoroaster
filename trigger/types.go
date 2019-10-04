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
	MatchedValues  string
	AllValues      string
	BlockTimestamp int
}

// Implements IMatch interface
func (CnMatch) isMatch() {}

// The CnMatch data we save in the jsonb data field on DB
type PersistentCnMatch struct {
	BlockNo        int
	BlockTimestamp int
	MatchedValues  string
	AllValues      string
}

func (m CnMatch) ToPersistent() *PersistentCnMatch {
	return &PersistentCnMatch{
		BlockNo:        m.BlockNo,
		BlockTimestamp: m.BlockTimestamp,
		MatchedValues:  m.MatchedValues,
		AllValues:      m.AllValues,
	}
}

// Outcome is the result of executing an Action;
// the Payload field can be a json struct
type Outcome struct {
	Outcome string
	Payload string
}

type EmailPayload struct {
	Recipients []string
	Body       string
}
