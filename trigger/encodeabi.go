package trigger

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"regexp"
	"strconv"
	"strings"
	"zoroaster/utils"
)

func encodeMethod(methodName, cntABI string, inputs []Input) (string, error) {

	xabi, err := abi.JSON(strings.NewReader(cntABI))
	if err != nil {
		return "", err
	}

	args := make([]interface{}, len(inputs))
	for i, in := range inputs {
		switch in.ParameterType {
		case "address":
			args[i] = common.HexToAddress(in.ParameterValue)
			continue
		case "uint32":
			v, _ := strconv.ParseInt(in.ParameterValue, 10, 32)
			args[i] = uint32(v)
			continue
		case "uint16":
			v, _ := strconv.ParseInt(in.ParameterValue, 10, 16)
			args[i] = uint16(v)
			continue
		case "uint8":
			v, _ := strconv.ParseInt(in.ParameterValue, 10, 8)
			args[i] = uint8(v)
			continue
		}
		intRgx := regexp.MustCompile(`u?int\d*$`) // all big int > 32 bits
		if intRgx.MatchString(in.ParameterType) {
			args[i] = utils.MakeBigInt(in.ParameterValue)
			continue
		}
	}

	result, err := xabi.Pack(methodName, args...)
	if err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(result), nil
}
