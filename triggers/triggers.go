package trigger

import (
	"github.com/INFURA/go-libs/jsonrpc_client"
)

// Triggers

type Trigger interface {
	checkCondition(transaction *jsonrpc_client.Transaction) bool
}

// Transaction FROM
type TriggerTransactionFrom struct {
	wallet string
}

func (ttf TriggerTransactionFrom) checkCondition(transaction *jsonrpc_client.Transaction) bool {

	return ttf.wallet == transaction.From
}

// Transaction NONCE
type TriggerTransactionNonce struct {
	filter Filter
}

func (ttf TriggerTransactionNonce) checkCondition(transaction *jsonrpc_client.Transaction) bool {

	switch v := ttf.filter.(type) {
	case GreaterThan:
		return transaction.Nonce > v.value
	case SmallerThan:
		return transaction.Nonce < v.value
	default:
		// TODO: this should never happen. Return an error perhaps?
		return false
	}
}

// TODO: this will read an array of transactions I guess
func TriggerAction(trigger Trigger, transaction jsonrpc_client.Transaction) bool {

	return trigger.checkCondition(&transaction)

}
