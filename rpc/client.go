package rpc

import (
	"github.com/onrik/ethrpc"
	"log"
	"time"
)

const maderoNode = "http://35.246.166.209:8545"
const matteoNode = "https://nodether.com"

func PollForLastBlock(c chan *ethrpc.Block) {

	var mostRecentBlockNo int
	client := ethrpc.New(matteoNode)

	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		n, err := client.EthBlockNumber()
		if err != nil {
			log.Fatal(err)
		}
		if n != mostRecentBlockNo {
			block, err := client.EthGetBlockByNumber(n, true)
			if err != nil {
				log.Fatal(err)
			}
			mostRecentBlockNo = n
			c <- block
		}
	}
}
