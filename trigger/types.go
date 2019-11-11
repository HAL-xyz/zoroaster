package trigger

import (
	"encoding/json"
	"github.com/onrik/ethrpc"
	"math/big"
)

// A match as represented internally by Zoroaster
type IMatch interface {
	ToPersistent() IPersistableMatch
	ToPostPayload() IPostablePaylaod
	GetTriggerUUID() string
	GetUserUUID() string
	GetMatchUUID() string
}

// A match persisted on the DB (in its json form)
type IPersistableMatch interface {
	isPersistable()
}

// A payload sent via web hook, and persisted under outcomes.payload
type IPostablePaylaod interface {
	isPostablePayload()
}

type ZTransaction struct {
	BlockTimestamp int
	DecodedFnArgs  *string `json:"DecodedFnArgs,omitempty"`
	DecodedFnName  *string `json:"DecodedFnName,omitempty"`
	Tx             *ethrpc.Transaction
}

// TX MATCH

type TxMatch struct {
	MatchUUID string
	Tg        *Trigger
	ZTx       *ZTransaction
}

type PersistentTx struct {
	BlockHash      string
	BlockNumber    *int
	BlockTimestamp int
	From           string
	Gas            int
	GasPrice       *big.Int
	Nonce          int
	To             string
	Hash           string
}

type PersistentTxMatch struct {
	DecodedData struct {
		FunctionArguments *string
		FunctionName      *string
	}
	PTx PersistentTx `json:"Transaction"`
}

func (m TxMatch) ToPersistent() IPersistableMatch {
	return &PersistentTxMatch{
		PTx: PersistentTx{
			BlockHash:      m.ZTx.Tx.BlockHash,
			BlockNumber:    m.ZTx.Tx.BlockNumber,
			BlockTimestamp: m.ZTx.BlockTimestamp,
			From:           m.ZTx.Tx.From,
			Gas:            m.ZTx.Tx.Gas,
			GasPrice:       &m.ZTx.Tx.GasPrice,
			Nonce:          m.ZTx.Tx.Nonce,
			To:             m.ZTx.Tx.To,
			Hash:           m.ZTx.Tx.Hash,
		},
		DecodedData: struct {
			FunctionArguments *string
			FunctionName      *string
		}{
			m.ZTx.DecodedFnArgs,
			m.ZTx.DecodedFnName,
		},
	}
}

func (m TxMatch) GetTriggerUUID() string {
	return m.Tg.TriggerUUID
}

func (m TxMatch) GetMatchUUID() string {
	return m.MatchUUID
}

func (m TxMatch) GetUserUUID() string {
	return m.Tg.UserUUID
}

type TxPostPayload struct {
	DecodedData struct {
		FunctionArguments *string
		FunctionName      *string
	}
	Transaction PersistentTx
	TriggerName string
	TriggerType string
	TriggerUUID string
}

func (TxPostPayload) isPostablePayload() {}

func (m TxMatch) ToPostPayload() IPostablePaylaod {
	return TxPostPayload{
		Transaction: PersistentTx{
			BlockHash:      m.ZTx.Tx.BlockHash,
			BlockNumber:    m.ZTx.Tx.BlockNumber,
			BlockTimestamp: m.ZTx.BlockTimestamp,
			From:           m.ZTx.Tx.From,
			Gas:            m.ZTx.Tx.Gas,
			GasPrice:       &m.ZTx.Tx.GasPrice,
			Nonce:          m.ZTx.Tx.Nonce,
			To:             m.ZTx.Tx.To,
			Hash:           m.ZTx.Tx.Hash,
		},
		DecodedData: struct {
			FunctionArguments *string
			FunctionName      *string
		}{
			m.ZTx.DecodedFnArgs,
			m.ZTx.DecodedFnName,
		},
		TriggerName: m.Tg.TriggerName,
		TriggerType: m.Tg.TriggerType,
		TriggerUUID: m.Tg.TriggerUUID,
	}
}

func (PersistentTxMatch) isPersistable() {}

// CONTRACT MATCH

type CnMatch struct {
	Trigger        *Trigger
	BlockNumber    int
	BlockTimestamp int
	BlockHash      string
	MatchUUID      string
	MatchedValues  string
	AllValues      []interface{}
}

type PersistentCnMatch struct {
	BlockNumber    int
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
	stringValues, _ := json.Marshal(m.AllValues)

	return &PersistentCnMatch{
		BlockNumber:    m.BlockNumber,
		BlockTimestamp: m.BlockTimestamp,
		BlockHash:      m.BlockHash,
		ContractAdd:    m.Trigger.ContractAdd,
		FunctionName:   m.Trigger.MethodName,
		ReturnedData: struct {
			MatchedValues string
			AllValues     string
		}{
			MatchedValues: m.MatchedValues,
			AllValues:     string(stringValues),
		},
	}
}

func (PersistentCnMatch) isPersistable() {}

type CnPostPayload struct {
	BlockNumber    int
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
	TriggerUUID string
}

func (CnPostPayload) isPostablePayload() {}

func (m CnMatch) ToPostPayload() IPostablePaylaod {
	stringValues, _ := json.Marshal(m.AllValues)

	return &CnPostPayload{
		BlockNumber:    m.BlockNumber,
		BlockTimestamp: m.BlockTimestamp,
		BlockHash:      m.BlockHash,
		ContractAdd:    m.Trigger.ContractAdd,
		FunctionName:   m.Trigger.MethodName,
		ReturnedData: struct {
			MatchedValues string
			AllValues     string
		}{
			MatchedValues: m.MatchedValues,
			AllValues:     string(stringValues),
		},
		TriggerName: m.Trigger.TriggerName,
		TriggerType: m.Trigger.TriggerType,
		TriggerUUID: m.Trigger.TriggerUUID,
	}
}

func (m CnMatch) GetTriggerUUID() string {
	return m.Trigger.TriggerUUID
}

func (m CnMatch) GetMatchUUID() string {
	return m.MatchUUID
}

func (m CnMatch) GetUserUUID() string {
	return m.Trigger.UserUUID
}

// EVENT MATCH

type EventMatch struct {
	MatchUUID      string
	Tg             *Trigger
	Log            *ethrpc.Log
	EventParams    map[string]string
	BlockTimestamp int
}

type PersistentEventMatch struct {
	ContractAdd string
	EventName   string
	EventData   struct {
		EventParameters map[string]string // decoded data + topics
		Data            string
		Topics          []string
	}
	Transaction struct {
		BlockHash      string
		BlockNumber    int
		BlockTimestamp int
		Hash           string
	} `json:"Transaction"`
}

func (PersistentEventMatch) isPersistable() {}

func (m EventMatch) ToPersistent() IPersistableMatch {
	return &PersistentEventMatch{
		ContractAdd: m.Tg.ContractAdd,
		EventName:   m.Tg.Filters[0].EventName,
		EventData: struct {
			EventParameters map[string]string // decoded data + topics
			Data            string
			Topics          []string
		}{
			EventParameters: m.EventParams,
			Data:            m.Log.Data,
			Topics:          m.Log.Topics,
		},
		Transaction: struct {
			BlockHash      string
			BlockNumber    int
			BlockTimestamp int
			Hash           string
		}{
			BlockHash:      m.Log.BlockHash,
			BlockNumber:    m.Log.BlockNumber,
			BlockTimestamp: m.BlockTimestamp,
			Hash:           m.Log.TransactionHash,
		},
	}
}

type EventPostPayload struct {
	ContractAdd string
	EventName   string
	EventData   struct {
		EventParameters map[string]string // decoded data + topics
		Data            string
		Topics          []string
	}
	Transaction struct {
		BlockHash      string
		BlockNumber    int
		BlockTimestamp int
		Hash           string
	}
	TriggerName string
	TriggerType string
	TriggerUUID string
}

func (EventPostPayload) isPostablePayload() {}

func (m EventMatch) ToPostPayload() IPostablePaylaod {
	return &EventPostPayload{
		ContractAdd: m.Tg.ContractAdd,
		EventName:   m.Tg.Filters[0].EventName,
		EventData: struct {
			EventParameters map[string]string // decoded data + topics
			Data            string
			Topics          []string
		}{
			EventParameters: m.EventParams,
			Data:            m.Log.Data,
			Topics:          m.Log.Topics,
		},
		Transaction: struct {
			BlockHash      string
			BlockNumber    int
			BlockTimestamp int
			Hash           string
		}{
			BlockHash:      m.Log.BlockHash,
			BlockNumber:    m.Log.BlockNumber,
			BlockTimestamp: m.BlockTimestamp,
			Hash:           m.Log.TransactionHash,
		},
		TriggerName: m.Tg.TriggerName,
		TriggerType: m.Tg.TriggerType,
		TriggerUUID: m.Tg.TriggerUUID,
	}
}

func (m EventMatch) GetTriggerUUID() string {
	return m.Tg.TriggerUUID
}

func (m EventMatch) GetMatchUUID() string {
	return m.MatchUUID
}

func (m EventMatch) GetUserUUID() string {
	return m.Tg.UserUUID
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
	HttpCode int
	Response string
}

type EmailPayload struct {
	Recipients []string
	Body       string
	Subject    string
}

type TgType int

const (
	WaT TgType = iota
	WaC
	WaE
)

func TgTypeToString(tgType TgType) string {
	switch tgType {
	case WaT:
		return "WatchTransactions"
	case WaC:
		return "WatchContracts"
	case WaE:
		return "WatchEvents"
	default:
		return ""
	}
}

func TgTypeToPrefix(tgType TgType) string {
	switch tgType {
	case WaT:
		return "wat"
	case WaC:
		return "wac"
	case WaE:
		return "wae"
	default:
		return ""
	}
}
