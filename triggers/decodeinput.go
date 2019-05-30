package trigger

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
)

func DecodeInputData(data string, cntABI string) (map[string]interface{}, error) {

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

	getMap := map[string]interface{}{}

	// unpack method inputs
	err = method.Inputs.UnpackIntoMap(getMap, decodedData)
	if err != nil {
		return nil, err
	}

	return getMap, nil
}

func DecodeInputMethod(data *string, cntABI *string) (*string, error) {

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
