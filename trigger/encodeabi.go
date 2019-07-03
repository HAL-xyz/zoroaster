package trigger

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"strings"
)

func EncodeMethod(methodName, cntABI string, inputs []Input) (string, error) {

	xabi, err := abi.JSON(strings.NewReader(cntABI))
	if err != nil {
		return "", err
	}

	args := make([]interface{}, len(inputs))
	for i, in := range inputs {
		if in.ParameterType == "Address" {
			args[i] = common.HexToAddress(in.ParameterValue)
		}
	}

	result, err := xabi.Pack(methodName, args...)
	if err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(result), nil
}
