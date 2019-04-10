package trigger

import (
	"encoding/json"
	"github.com/INFURA/go-libs/jsonrpc_client"
)

func JsonToBlock(jsonBlock string) (*jsonrpc_client.Block, error) {

	var block jsonrpc_client.Block
	err := json.Unmarshal([]byte(jsonBlock), &block)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

