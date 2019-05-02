package trigger

import (
	"encoding/json"
	"github.com/INFURA/go-libs/jsonrpc_client"
	"io/ioutil"
	"log"
)

func JsonToTransaction(jsonTx string) (*jsonrpc_client.Transaction, error) {

	var tx jsonrpc_client.Transaction
	err := json.Unmarshal([]byte(jsonTx), &tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func getTransactionFromFile(path string) *jsonrpc_client.Transaction {
	txSrc, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	tx, err := JsonToTransaction(string(txSrc))
	if err != nil {
		log.Fatal(err)
	}
	return tx
}


func JsonToBlock(jsonBlock string) (*jsonrpc_client.Block, error) {

	var block jsonrpc_client.Block
	err := json.Unmarshal([]byte(jsonBlock), &block)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func getBlockFromFile(path string) *jsonrpc_client.Block {
	blockSrc, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	block, err := JsonToBlock(string(blockSrc))
	if err != nil {
		log.Fatal(err)
	}
	return block
}

func getTriggerFromFile(path string) *Trigger {
	triggerSrc, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	return getTriggerFromJson(string(triggerSrc))
}

func getTriggerFromJson(json string) *Trigger {
	tjs, err := NewTriggerJson(json)
	if err != nil {
		log.Fatal("Cannot parse json trigger:", err)
	}

	tg, err := tjs.ToTrigger()
	if err != nil {
		log.Fatal("Cannot convert json trigger to type trigger:", err)
	}
	return tg
}
