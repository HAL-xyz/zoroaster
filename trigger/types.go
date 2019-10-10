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
	ToPersistent() IPersistableMatch
	GetTriggerId() int
}

type IPersistableMatch interface {
	isPersistable()
}

type TxMatch struct {
	MatchId int
	Tg      *Trigger
	ZTx     *ZTransaction
}

// The TxMatch data we save in the jsonb data field on DB
type PersistentTxMatch struct {
	DecodedData struct {
		FunctionArguments *string
		FunctionName      *string
	}
	Tx *ethrpc.Transaction
}

func (m TxMatch) ToPersistent() IPersistableMatch {
	return &PersistentTxMatch{
		Tx: m.ZTx.Tx,
		DecodedData: struct {
			FunctionArguments *string
			FunctionName      *string
		}{
			m.ZTx.DecodedFnArgs,
			m.ZTx.DecodedFnName,
		},
	}
}

func (m TxMatch) GetTriggerId() int {
	return m.Tg.TriggerId
}

// Implements IPersistable interface
func (PersistentTxMatch) isPersistable() {}

// Implements IMatch interface
func (TxMatch) isMatch() {}

type CnMatch struct {
	Trigger        *Trigger
	BlockNo        int
	BlockTimestamp int
	BlockHash      string
	MatchId        int
	MatchedValues  string
	AllValues      string
}

// Implements IMatch interface
func (CnMatch) isMatch() {}

// The CnMatch data we save in the jsonb data field on DB
type PersistentCnMatch struct {
	BlockNo        int
	BlockTimestamp int
	BlockHash      string
	ContractAdd    string
	FunctionName   string
	ReturnedData   struct {
		MatchedValues string
		AllValues     string
	}
}

func (m CnMatch) ToPersistent() IPersistableMatch {
	return &PersistentCnMatch{
		BlockNo:        m.BlockNo,
		BlockTimestamp: m.BlockTimestamp,
		BlockHash:      m.BlockHash,
		ContractAdd:    m.Trigger.ContractAdd,
		FunctionName:   m.Trigger.MethodName,
		ReturnedData: struct {
			MatchedValues string
			AllValues     string
		}{
			MatchedValues: m.MatchedValues,
			AllValues:     m.AllValues},
	}
}

// Implements IPersistable interface
func (PersistentCnMatch) isPersistable() {}

// A json version of this is what we send via web hook;
// it is also what gets stored in the db under `outcomes.payload`.
type ContractPostData struct {
	BlockNo        int
	BlockTimestamp int
	ReturnedValue  string
	AllValues      string
}

func (m CnMatch) ToCnPostData() *ContractPostData {
	return &ContractPostData{
		BlockNo:        m.BlockNo,
		BlockTimestamp: m.BlockTimestamp,
		ReturnedValue:  m.MatchedValues,
		AllValues:      m.AllValues,
	}
}

func (m CnMatch) GetTriggerId() int {
	return m.Trigger.TriggerId
}

// Outcome is the result of executing an Action;
// it includes a payload (the body of the action request)
// and the actual outcome of that request.
// Both fields are represented as a json struct.
type Outcome struct {
	Payload string
	Outcome string
}

type WebhookResponse struct {
	StatusCode int
}

type EmailPayload struct {
	Recipients []string
	Body       string
}
