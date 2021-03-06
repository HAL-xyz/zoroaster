package trigger

import (
	"encoding/hex"
	"fmt"
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
	if len(data) <= 2 {
		return nil, fmt.Errorf("no input data")
	}

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

func decodeInputMethod(data *string, cntABI *string) (*string, error) {

	xabi, err := abi.JSON(strings.NewReader(*cntABI))
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, fmt.Errorf("no input data")
	}
	if len(*data) <= 2 {
		return nil, fmt.Errorf("no input data")
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
