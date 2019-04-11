package trigger

import (
	"github.com/INFURA/go-libs/jsonrpc_client"
	"github.com/satori/go.uuid"
)

/*
	Triggers implement the Trigger interface. Each Action will reference a Trigger.uuid.
 */

type Trigger interface {
	GetUUID() uuid.UUID
	checkCondition(transaction *jsonrpc_client.Transaction) (uuid.UUID, bool)
}

// Transaction FROM
type TriggerTransactionFrom struct {
	uuid uuid.UUID
	wallet string
}

func (tg TriggerTransactionFrom) GetUUID() uuid.UUID {
	return tg.uuid
}

func (tg TriggerTransactionFrom) checkCondition(ts *jsonrpc_client.Transaction) (uuid.UUID, bool) {

	return tg.uuid, tg.wallet == ts.From
}

// Transaction NONCE
type TriggerTransactionNonce struct {
	uuid uuid.UUID
	filter Filter
}

func (tg TriggerTransactionNonce) checkCondition(ts *jsonrpc_client.Transaction) (uuid.UUID, bool) {

	switch v := tg.filter.(type) {
	case GreaterThan:
		return tg.GetUUID(), ts.Nonce > v.value
	case SmallerThan:
		return tg.GetUUID(), ts.Nonce < v.value
	default:
		// TODO: this should never happen. Return an error perhaps?
		return tg.GetUUID(), false
	}
}

func (tg TriggerTransactionNonce) GetUUID() uuid.UUID {
	return tg.uuid
}



// TODO: this will read an array of transactions I guess
func TriggerAction(trigger Trigger, transaction jsonrpc_client.Transaction) (uuid.UUID, bool) {

	return trigger.checkCondition(&transaction)

}
