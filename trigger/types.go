package trigger

import "github.com/onrik/ethrpc"

type ZTransaction struct {
	BlockTimestamp int
	DecodedFnArgs  *string `json:"DecodedFnArgs,omitempty"`
	DecodedFnName  *string `json:"DecodedFnName,omitempty"`
	Tx             *ethrpc.Transaction
}

// A match as represented internally by Zoroaster
type IMatch interface {
	ToPersistent() IPersistableMatch
	ToPostPayload() IPostablePaylaod
	GetTriggerId() int
}

// A match persisted on the DB (in its json form)
type IPersistableMatch interface {
	isPersistable()
}

// A payload sent via web hook, and persisted under outcomes.payload
type IPostablePaylaod interface {
	isPostablePayload()
}

// TX MATCH

type TxMatch struct {
	MatchId int
	Tg      *Trigger
	ZTx     *ZTransaction
}

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

type TxPostPayload struct {
	DecodedData struct {
		FunctionArguments *string
		FunctionName      *string
	}
	Tx          *ethrpc.Transaction
	TriggerName string
	TriggerType string
	TriggerId   int
}

func (TxPostPayload) isPostablePayload() {}

func (m TxMatch) ToPostPayload() IPostablePaylaod {
	return TxPostPayload{
		Tx: m.ZTx.Tx,
		DecodedData: struct {
			FunctionArguments *string
			FunctionName      *string
		}{
			m.ZTx.DecodedFnArgs,
			m.ZTx.DecodedFnName,
		},
		TriggerName: m.Tg.TriggerName,
		TriggerType: m.Tg.TriggerType,
		TriggerId:   m.Tg.TriggerId,
	}
}

func (PersistentTxMatch) isPersistable() {}

// CONTRACT MATCH

type CnMatch struct {
	Trigger        *Trigger
	BlockNo        int
	BlockTimestamp int
	BlockHash      string
	MatchId        int
	MatchedValues  string
	AllValues      string
}

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
			AllValues:     m.AllValues,
		},
	}
}

func (PersistentCnMatch) isPersistable() {}

type CnPostPayload struct {
	BlockNo        int
	BlockTimestamp int
	BlockHash      string
	ContractAdd    string
	FunctionName   string
	ReturnedData   struct {
		MatchedValues string
		AllValues     string
	}
	TriggerName string
	TriggerType string
	TriggerId   int
}

func (CnPostPayload) isPostablePayload() {}

func (m CnMatch) ToPostPayload() IPostablePaylaod {
	return &CnPostPayload{
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
			AllValues:     m.AllValues,
		},
		TriggerName: m.Trigger.TriggerName,
		TriggerType: m.Trigger.TriggerType,
		TriggerId:   m.Trigger.TriggerId,
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
