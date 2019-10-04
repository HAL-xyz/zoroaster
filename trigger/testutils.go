package trigger

import (
	"encoding/json"
	"fmt"
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"strconv"
)

func jsonToTransaction(jsonSrc []byte) (*ethrpc.Transaction, error) {
	var tx ethrpc.Transaction
	err := json.Unmarshal(jsonSrc, &tx)
	if err != nil {
		return nil, err
	}
	fixHexTransaction(&tx)
	return &tx, nil
}

func GetTransactionFromFile(path string) *ethrpc.Transaction {
	txSrc, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err)
	}

	tx, err := jsonToTransaction(txSrc)
	if err != nil {
		log.Error(err)
	}
	return tx
}

func jsonToBlock(jsonBlock []byte) (*ethrpc.Block, error) {
	var block ethrpc.Block
	err := json.Unmarshal(jsonBlock, &block)
	if err != nil {
		return nil, err
	}
	for i := range block.Transactions {
		fixHexTransaction(&block.Transactions[i])
	}
	return &block, nil
}

func GetBlockFromFile(path string) *ethrpc.Block {
	blockSrc, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err)
	}
	block, err := jsonToBlock(blockSrc)
	if err != nil {
		log.Error(err)
	}
	return block
}

func NewTriggerFromFile(path string) (*Trigger, error) {
	triggerSrc, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err)
	}
	return NewTriggerFromJson(string(triggerSrc))
}

// ethrpc.Transaction expects some fields to be hex values,
// which is a pain for testing because we want to use integers
// in our json, not hex values. E.g. if we have the int
// 21000 in our json test, it will be interpreted as an hex value,
// and converted to the int 135168.
// This function is an hack to 135168 back to 21000.
func fixHexIntCasting(v int) int {
	s := fmt.Sprintf("%x", v)
	ret, err := strconv.Atoi(s)
	if err != nil {
		log.Error("int fix casting failed:", err)
	}
	return ret
}

func fixHexTransaction(tx *ethrpc.Transaction) {
	tx.Nonce = fixHexIntCasting(tx.Nonce)
	*tx.BlockNumber = fixHexIntCasting(*tx.BlockNumber)
	*tx.TransactionIndex = fixHexIntCasting(*tx.TransactionIndex)
	tx.Gas = fixHexIntCasting(tx.Gas)
}
