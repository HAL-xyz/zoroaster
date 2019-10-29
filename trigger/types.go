package trigger

import "github.com/onrik/ethrpc"

// A match as represented internally by Zoroaster
type IMatch interface {
	ToPersistent() IPersistableMatch
	ToPostPayload() IPostablePaylaod
	GetTriggerUUID() string
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

func (m TxMatch) GetTriggerUUID() string {
	return m.Tg.TriggerUUID
}

type TxPostPayload struct {
	DecodedData struct {
		FunctionArguments *string
		FunctionName      *string
	}
	Tx          *ethrpc.Transaction
	TriggerName string
	TriggerType string
	TriggerUUID string
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
		TriggerUUID: m.Tg.TriggerUUID,
	}
}

func (PersistentTxMatch) isPersistable() {}

// CONTRACT MATCH

type CnMatch struct {
	Trigger        *Trigger
	BlockNo        int
	BlockTimestamp int
	BlockHash      string
	MatchUUID      string
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
	TriggerUUID string
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
		TriggerUUID: m.Trigger.TriggerUUID,
	}
}

func (m CnMatch) GetTriggerUUID() string {
	return m.Trigger.TriggerUUID
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
		BlockNo        int
		BlockTimestamp int
		TxHash         string
	}
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
			BlockNo        int
			BlockTimestamp int
			TxHash         string
		}{
			BlockHash:      m.Log.BlockHash,
			BlockNo:        m.Log.BlockNumber,
			BlockTimestamp: m.BlockTimestamp,
			TxHash:         m.Log.TransactionHash,
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
		BlockNo        int
		BlockTimestamp int
		TxHash         string
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
			BlockNo        int
			BlockTimestamp int
			TxHash         string
		}{
			BlockHash:      m.Log.BlockHash,
			BlockNo:        m.Log.BlockNumber,
			BlockTimestamp: m.BlockTimestamp,
			TxHash:         m.Log.TransactionHash,
		},
		TriggerName: m.Tg.TriggerName,
		TriggerType: m.Tg.TriggerType,
		TriggerUUID: m.Tg.TriggerUUID,
	}
}

func (m EventMatch) GetTriggerUUID() string {
	return m.Tg.TriggerUUID
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
