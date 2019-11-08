package abidec

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
)

// Equivalent of https://web3js.readthedocs.io/en/v1.2.0/web3-eth-abi.html#decodeparameters
// I.e. takes the byte code returned by invoking a contract's function and returns
// a map with all the decoded arguments, e.g.
// given data = 0000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000067361696c6f72000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000046d6f6f6e00000000000000000000000000000000000000000000000000000000
// returns a map[r1:4 r2:sailor r3:moon]

func DecodeParameters(data string, cntABI string, methodName string) (map[string]interface{}, error) {

	encb, err := hex.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("invalid hex: %s", data)
	}

	xabi, err := abi.JSON(strings.NewReader(cntABI))
	if err != nil {
		return nil, fmt.Errorf("cannot read abi: %s", err)
	}

	methodObj, ok := xabi.Methods[methodName]
	if !ok {
		return nil, fmt.Errorf("method %s not found", methodName)
	}

	getMap := map[string]interface{}{}
	err = methodObj.Outputs.UnpackIntoMap(getMap, encb)

	if err != nil {
		return nil, fmt.Errorf("cannot unpack outputs: %s", err)
	}

	return getMap, nil
}

func DecodeParamsIntoList(data string, cntABI string, methodName string) ([]interface{}, error) {

	encb, err := hex.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("invalid hex: %s", data)
	}

	xabi, err := abi.JSON(strings.NewReader(cntABI))
	if err != nil {
		return nil, fmt.Errorf("cannot read abi: %s", err)
	}

	methodObj, ok := xabi.Methods[methodName]
	if !ok {
		return nil, fmt.Errorf("method %s not found", methodName)
	}

	ls, err := methodObj.Outputs.UnpackValues(encb)

	if err != nil {
		return nil, fmt.Errorf("cannot unpack outputs: %s", err)
	}

	return ls, nil
}

func DecodeParamsToJsonMap(data string, cntABI string, methodName string) (map[string]json.RawMessage, error) {
	ifData, err := DecodeParameters(data, cntABI, methodName)
	if err != nil {
		return nil, err
	}
	jsonData, err := json.Marshal(ifData)
	if err != nil {
		return nil, err
	}
	out := map[string]json.RawMessage{}
	err = json.Unmarshal(jsonData, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
