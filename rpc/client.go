package main

import (
	json2 "encoding/json"
	"fmt"
	"github.com/INFURA/go-libs/jsonrpc_client"
)

// TODO: require private secret
const (
	MAINNET_URL = "https://mainnet.infura.io/v3/"
	RINKEBY_URL = "https://rinkeby.infura.io/v3/"
	PROJECT_ID = "448136c4f7b5486995b34fb9e13f2a32"
	ENDPOINT   = RINKEBY_URL + PROJECT_ID
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

	trans, _ := client.Eth_getTransactionByHash("0x83c283364249687c4d2d41fd1822c06fba9ed4fadeefb88e1ececf6e22598c82")
	js2, _ := json2.Marshal(trans)
	fmt.Println(string(js2))




}
