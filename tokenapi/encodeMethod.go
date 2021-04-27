package tokenapi

import (
	"encoding/hex"
	"fmt"
	"github.com/HAL-xyz/zoroaster/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"regexp"
	"strconv"
	"strings"
)

type Input struct {
	ParameterType  string
	ParameterValue string
}

func encodeMethod(methodName, cntABI string, inputs []Input) (string, error) {

	xabi, err := abi.JSON(strings.NewReader(cntABI))
	if err != nil {
		return "", err
	}

	args := make([]interface{}, len(inputs))
	for i, in := range inputs {
		in.ParameterValue = strings.TrimPrefix(in.ParameterValue, "0x")

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
		case "int32":
			v, _ := strconv.ParseInt(in.ParameterValue, 10, 32)
			args[i] = int32(v)
			continue
		case "int16":
			v, _ := strconv.ParseInt(in.ParameterValue, 10, 16)
			args[i] = int16(v)
			continue
		case "int8":
			v, _ := strconv.ParseInt(in.ParameterValue, 10, 8)
			args[i] = int8(v)
			continue
		case "bytes":
			args[i] = common.Hex2Bytes(in.ParameterValue)
			continue
		case "bytes1", "byte":
			v := [1]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes2":
			v := [2]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes3":
			v := [3]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes4":
			v := [4]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes5":
			v := [5]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes6":
			v := [6]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes7":
			v := [7]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes8":
			v := [8]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes9":
			v := [9]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes10":
			v := [10]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes11":
			v := [11]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes12":
			v := [12]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes13":
			v := [13]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes14":
			v := [14]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes15":
			v := [15]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes16":
			v := [16]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes17":
			v := [17]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes18":
			v := [18]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes19":
			v := [19]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes20":
			v := [20]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes21":
			v := [21]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes22":
			v := [22]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes23":
			v := [16]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes24":
			v := [24]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes25":
			v := [25]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes26":
			v := [26]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes27":
			v := [27]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes28":
			v := [28]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes29":
			v := [29]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes30":
			v := [30]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes31":
			v := [31]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		case "bytes32":
			v := [32]byte{}
			copy(v[:], common.Hex2Bytes(in.ParameterValue))
			args[i] = v
			continue
		}
		intRgx := regexp.MustCompile(`u?int\d*$`) // all big int > 32 bits
		if intRgx.MatchString(in.ParameterType) {
			args[i] = utils.MakeBigInt(in.ParameterValue)
			continue
		}
		arrayRgx8 := regexp.MustCompile(`int8\[\d*]$`) // int8[], int8[N]
		if arrayRgx8.MatchString(in.ParameterType) {
			ss := utils.GetValsFromStringifiedArray(in.ParameterValue)
			params := make([]int8, len(ss))
			for i, v := range ss {
				numericVal, _ := strconv.Atoi(v)
				params[i] = int8(numericVal)
			}
			args[i] = params
			continue
		}
		arrayRgx16 := regexp.MustCompile(`int16\[\d*]$`) // int16[], int16[N]
		if arrayRgx16.MatchString(in.ParameterType) {
			ss := utils.GetValsFromStringifiedArray(in.ParameterValue)
			params := make([]int16, len(ss))
			for i, v := range ss {
				numericVal, _ := strconv.Atoi(v)
				params[i] = int16(numericVal)
			}
			args[i] = params
			continue
		}
		arrayRgx32 := regexp.MustCompile(`int32\[\d*]$`) // int32[], int32[N]
		if arrayRgx32.MatchString(in.ParameterType) {
			ss := utils.GetValsFromStringifiedArray(in.ParameterValue)
			params := make([]int32, len(ss))
			for i, v := range ss {
				numericVal, _ := strconv.Atoi(v)
				params[i] = int32(numericVal)
			}
			args[i] = params
			continue
		}
		addressRgx := regexp.MustCompile(`address\[\d*]$`) // address[], address[N]
		if addressRgx.MatchString(in.ParameterType) {
			ss := utils.GetValsFromStringifiedArray(in.ParameterValue)
			params := make([]common.Address, len(ss))
			for i, v := range ss {
				params[i] = common.HexToAddress(v)
			}
			args[i] = params
			continue
		}
		return "", fmt.Errorf("Unsupported param type: %s\n", in.ParameterType)
	}

	result, err := xabi.Pack(methodName, args...)
	if err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(result), nil
}
