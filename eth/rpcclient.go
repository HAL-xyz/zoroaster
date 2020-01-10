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
	"zoroaster/trigger"
)

func BlocksPoller(
	txChan chan *ethrpc.Block,
	cnChan chan *ethrpc.Block,
	evChan chan *ethrpc.Block,
	client *ethrpc.EthRPC,
	idb aws.IDB,
	blocksDelay int) {

	txLastBlockProcessed, err1 := idb.ReadLastBlockProcessed(trigger.WaT)
	cnLastBlockProcessed, err2 := idb.ReadLastBlockProcessed(trigger.WaC)
	evLastBlockProcessed, err3 := idb.ReadLastBlockProcessed(trigger.WaE)

	if err1 != nil || err2 != nil || err3 != nil {
		log.Fatal(err1, err2, err3)
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	for range ticker.C {
		lastBlockSeen, err := client.EthBlockNumber()
		if err != nil {
			log.Warn("failed to poll ETH node -> ", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Watch a Transaction
		fetchLastBlock(lastBlockSeen, &txLastBlockProcessed, txChan, client, true, blocksDelay)

		// Watch a Contract
		fetchLastBlock(lastBlockSeen, &cnLastBlockProcessed, cnChan, client, false, blocksDelay)

		// Watch an Event
		fetchLastBlock(lastBlockSeen, &evLastBlockProcessed, evChan, client, false, blocksDelay)
	}
}

func fetchLastBlock(
	lastBlockSeen int,
	lastBlockProcessed *int,
	ch chan *ethrpc.Block,
	client *ethrpc.EthRPC,
	withTxs bool,
	blocksDelay int) {

	// this is used to reset the last block processed
	if *lastBlockProcessed == 0 {
		*lastBlockProcessed = lastBlockSeen - blocksDelay
	}

	if lastBlockSeen-blocksDelay > *lastBlockProcessed {
		block, err := client.EthGetBlockByNumber(*lastBlockProcessed+1, withTxs)
		if err != nil {
			log.Warnf("failed to get block %d -> %s", *lastBlockProcessed+1, err)
			time.Sleep(5 * time.Second)
		} else {
			*lastBlockProcessed += 1
			ch <- block
		}
	}
}

func GetModifiedAccounts(blockMinusOneNo, blockNo int, nodeURI string) []string {

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

	response, err := http.Post(nodeURI, "application/json", bytes.NewBuffer(body))

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
