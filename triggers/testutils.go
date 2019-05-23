package trigger

import (
	"encoding/json"
	"fmt"
	"github.com/onrik/ethrpc"
	"io/ioutil"
	"log"
	"strconv"
)

func JsonToTransaction(jsonSrc []byte) (*ethrpc.Transaction, error) {
	var tx ethrpc.Transaction
	err := json.Unmarshal(jsonSrc, &tx)
	if err != nil {
		return nil, err
	}
	tx.Nonce = fixIntCasting(tx.Nonce)
	tx.Gas = fixIntCasting(tx.Gas)
	return &tx, nil
}

func getTransactionFromFile(path string) *ethrpc.Transaction {
	txSrc, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	tx, err := JsonToTransaction(txSrc)
	if err != nil {
		log.Fatal(err)
	}
	return tx
}

func JsonToBlock(jsonBlock []byte) (*ethrpc.Block, error) {
	var block ethrpc.Block
	err := json.Unmarshal(jsonBlock, &block)
	if err != nil {
		return nil, err
	}
	for i, t := range block.Transactions {
		block.Transactions[i].Gas = fixIntCasting(t.Gas)
		block.Transactions[i].Nonce = fixIntCasting(t.Nonce)
	}
	return &block, nil
}

func GetBlockFromFile(path string) *ethrpc.Block {
	blockSrc, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	block, err := JsonToBlock(blockSrc)
	if err != nil {
		log.Fatal(err)
	}
	return block
}

func NewTriggerFromFile(path string) (*Trigger, error) {
	triggerSrc, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return NewTriggerFromJson(string(triggerSrc))
}

// ethrpc.Transaction expects Gas and Nonce to be hex values,
// which is a pain for testing. E.g. if we have the int
// 21000 in the tests it will be interpreted as an hex value,
// and converted to the int 135168 when reading the Json file.
// This function converts 135168 back to 21000.
func fixIntCasting(v int) int {
	s := fmt.Sprintf("%x", v)
	ret, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal("int fix casting failed:", err)
	}
	return ret
}
