package trigger

import (
	"encoding/json"
	"fmt"
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"strconv"
)

func JsonToTransaction(jsonSrc []byte) (*ethrpc.Transaction, error) {
	var tx ethrpc.Transaction
	err := json.Unmarshal(jsonSrc, &tx)
	if err != nil {
		return nil, err
	}
	fixHexTransaction(&tx)
	return &tx, nil
}

func GetTransactionFromFile(path string) (*ethrpc.Transaction, error) {
	txSrc, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	tx, err := JsonToTransaction(txSrc)
	if err != nil {
		return nil, err
	}
	return tx, nil
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

func GetBlockFromFile(path string) (*ethrpc.Block, error) {
	blockSrc, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, err := jsonToBlock(blockSrc)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func GetLogsFromFile(path string) ([]ethrpc.Log, error) {
	logsSrc, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err)
	}
	var logs []ethrpc.Log
	err = json.Unmarshal(logsSrc, &logs)
	if err != nil {
		return nil, err
	}
	for i := range logs {
		fixHexLog(&logs[i])
	}
	return logs, nil
}

func GetTriggerFromFile(path string) (*Trigger, error) {
	triggerSrc, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err)
	}
	return NewTriggerFromJson(string(triggerSrc))
}

// ethrpc.Transaction and ethrpc.Log expects some fields to be hex values,
// which is a pain for testing because we want to use integers
// in our json, not hex values. E.g. if we have the int
// 21000 in our json test, it will be interpreted as an hex value,
// and converted to the int 135168.
// This function is an hack to convert 135168 back to 21000.
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

func fixHexLog(log *ethrpc.Log) {
	log.LogIndex = fixHexIntCasting(log.LogIndex)
	log.TransactionIndex = fixHexIntCasting(log.TransactionIndex)
	log.BlockNumber = fixHexIntCasting(log.BlockNumber)
}
