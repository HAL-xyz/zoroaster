package main

import (
	json2 "encoding/json"
	"fmt"
	"github.com/INFURA/go-libs/jsonrpc_client"
)

// TODO: require private secret
const (
	INFURA_URL = "https://mainnet.infura.io/v3/"
	PROJECT_ID = "448136c4f7b5486995b34fb9e13f2a32"
	ENDPOINT   = INFURA_URL + PROJECT_ID
)

func main() {

	client := jsonrpc_client.EthereumClient{ENDPOINT}

	//lastBlockNumber, err := client.Eth_blockNumber()
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(lastBlockNumber)
	//
	//lastBlock, err := client.Eth_getBlockByNumber(lastBlockNumber, true)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//js, _ := json2.Marshal(lastBlock)
	//fmt.Println(string(js))

	trans, _ := client.Eth_getTransactionByHash("0xde18abb190b828ab0d515297b01a7860f589998b2b0caae187aaba687122096f")
	js2, _ := json2.Marshal(trans)
	fmt.Println(string(js2))




}
