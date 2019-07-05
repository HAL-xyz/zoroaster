package eth

import (
	"bytes"
	"encoding/json"
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
	"zoroaster/aws"
	"zoroaster/config"
)

func BlocksPoller(c chan *ethrpc.Block, client *ethrpc.EthRPC, zconf *config.ZConfiguration) {

	const K = 8 // next block to process is (last block mined - K)

	lastBlockProcessed := aws.ReadLastBlockProcessed(zconf.TriggersDB.TableStats)

	ticker := time.NewTicker(2500 * time.Millisecond)
	for range ticker.C {
		n, err := client.EthBlockNumber()
		if err != nil {
			log.Warn("failed to poll ETH node -> ", err)
			time.Sleep(5 * time.Second)
			continue
		}
		// this should only happen during dev
		if lastBlockProcessed == 0 {
			lastBlockProcessed = n - K
		}
		if n-K > lastBlockProcessed {
			block, err := client.EthGetBlockByNumber(lastBlockProcessed+1, true)
			if err != nil {
				log.Warnf("failed to get block %d -> %s", n, err)
				time.Sleep(5 * time.Second)
				continue
			}
			lastBlockProcessed += 1
			log.Infof("(BlocksPoller is %d blocks behind)", n-lastBlockProcessed)
			c <- block
		}
	}
}

func GetModifiedAccounts(blockMinusOneNo, blockNo int) []string {

	type ethRequest struct {
		ID      int    `json:"id"`
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  []int  `json:"params"`
	}

	p := []int{blockMinusOneNo, blockNo}

	request := ethRequest{
		ID:      1,
		JSONRPC: "2.0",
		Method:  "debug_getModifiedAccountsByNumber",
		Params:  p,
	}

	body, err := json.Marshal(request)
	if err != nil {
		log.Error(err)
		return nil
	}

	cxtLog := log.WithFields(log.Fields{
		"request": string(body),
	})

	node := "https://reader:PVHCtb9AT4NzUY3ZpWs8nFTG2wJdKuju3Y3FPCf9YnULfsA4RTcfJBw2rfadhzeT@node-0.hal.xyz"

	response, err := http.Post(node, "application/json", bytes.NewBuffer(body))

	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		cxtLog.Error(err)
		return nil
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		cxtLog.Error(err)
		return nil
	}

	if response.StatusCode != 200 {
		cxtLog.Error(err)
		return nil
	}

	// result be like
	// {"jsonrpc":"2.0","id":1,"result":["0x31b93ca83b5ad17582e886c400667c6f698b8ccd",...]}

	type ethResponse struct {
		Result []string `json:"result"`
		Error  struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	var ethResp ethResponse

	err = json.Unmarshal(data, &ethResp)
	if err != nil {
		cxtLog.Error(err)
		return nil
	}

	if ethResp.Error.Message != "" {
		cxtLog.Error(ethResp.Error.Message)
	}

	return ethResp.Result
}
