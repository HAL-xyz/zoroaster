package rpc

import (
	"github.com/onrik/ethrpc"
	"log"
	"time"
)

const maderoNode = "http://35.246.166.209:8545"
const matteoNode = "https://nodether.com"
const K = 8 // next block to process is (last block mined - K)

func PollForLastBlock(c chan *ethrpc.Block) {

	var lastBlockProcessed int
	client := ethrpc.New(matteoNode)

	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		n, err := client.EthBlockNumber()
		if err != nil {
			log.Println("WARN: failed to poll ETH node -> ", err)
			time.Sleep(5 * time.Second)
			continue
		}
		// program startup
		if lastBlockProcessed == 0 {
			lastBlockProcessed = n - K
		}
		if n - K > lastBlockProcessed {
			block, err := client.EthGetBlockByNumber(lastBlockProcessed + 1, true)
			if err != nil {
				log.Printf("WARN: failed to get block %d -> %s", n, err)
				time.Sleep(5 * time.Second)
				continue
			}
			lastBlockProcessed += 1
			log.Printf("\t(%d blocks behind)", n - lastBlockProcessed)
			c <- block
		}
	}
}
