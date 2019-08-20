package trigger

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"strconv"
	"strings"
)

func encodeMethod(methodName, cntABI string, inputs []Input) (string, error) {

	xabi, err := abi.JSON(strings.NewReader(cntABI))
	if err != nil {
		return "", err
	}

	args := make([]interface{}, len(inputs))
	for i, in := range inputs {
		switch in.ParameterType {
		case "Address":
			args[i] = common.HexToAddress(in.ParameterValue)
		case "uint256":
			args[i] = makeBigInt(in.ParameterValue)
		case "uint16":
			v, _ := strconv.ParseInt(in.ParameterValue, 10, 16)
			args[i] = uint16(v)
		}
	}

	result, err := xabi.Pack(methodName, args...)
	if err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(result), nil
}
