package trigger

import (
	"encoding/hex"
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
)

func decodeInputData(data string, cntABI string) (map[string]interface{}, error) {

	// load contract ABI
	xabi, err := abi.JSON(strings.NewReader(cntABI))
	if err != nil {
		return nil, err
	}

	// decode method signature
	// strip hex prefix (0x)
	// signature is the first 32 bits of the hash of the function
	// in HEX 1 char = 4 bits, so 32 bits = 8 hex chars
	decodedSig, err := hex.DecodeString(data[2:10])
	if err != nil {
		return nil, err
	}

	// recover Method from signature and ABI
	method, err := xabi.MethodById(decodedSig)
	if err != nil {
		return nil, err
	}

	// decode function arguments
	decodedData, err := hex.DecodeString(data[10:])
	if err != nil {
		return nil, err
	}

	// unpack method inputs
	getMap := map[string]interface{}{}
	err = method.Inputs.UnpackIntoMap(getMap, decodedData)
	if err != nil {
		return nil, err
	}

	return getMap, nil
}

func decodeInputDataToJsonMap(data string, cntABI string) (map[string]json.RawMessage, error) {
	ifData, err := decodeInputData(data, cntABI)
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

func decodeInputMethod(data *string, cntABI *string) (*string, error) {

	xabi, err := abi.JSON(strings.NewReader(*cntABI))
	if err != nil {
		return nil, err
	}

	decodedSig, err := hex.DecodeString((*data)[2:10])
	if err != nil {
		return nil, err
	}

	method, err := xabi.MethodById(decodedSig)
	if err != nil {
		return nil, err
	}
	return &method.Name, nil
}
