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
// it includes a payload (the body of the action request)
// and the actual outcome of that request.
// Both fields are represented as a json struct.
type Outcome struct {
	Payload string
	Outcome string
}

// A json version of this is what we send via web hook;
// it is also what gets stored in the db under `outcomes.payload`.
type ContractPostData struct {
	BlockNo        int
	BlockTimestamp int
	ReturnedValue  string
	AllValues      string
}

func (m *CnMatch) ToCnPostData() *ContractPostData {
	return &ContractPostData{
		BlockNo:        m.BlockNo,
		BlockTimestamp: m.BlockTimestamp,
		ReturnedValue:  m.MatchedValues,
		AllValues:      m.AllValues,
	}
}

type WebhookResponse struct {
	StatusCode int
}

type EmailPayload struct {
	Recipients []string
	Body       string
}
