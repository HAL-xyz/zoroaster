package main

import (
	"fmt"
	"github.com/onrik/ethrpc"
	"log"
)

func main() {
	client := ethrpc.New("http://35.246.166.209:8545")

	n, err := client.EthBlockNumber()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Most recent block no: ", n)

	block, err := client.EthGetBlockByNumber(7535077, true)
	//js2, _ := json.Marshal(block)
	//fmt.Println(string(js2))
	fmt.Println("gas ", block.Transactions[5].Gas)

	//tx, err := client.EthGetTransactionByHash("0x0641bb18e73d9e874252d3de6993473d176200dc02f4482a64c6540749aecaff")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//js2, _ := json.Marshal(tx)
	//fmt.Println(string(js2))
}
