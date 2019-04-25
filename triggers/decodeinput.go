package trigger

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
)

func DecodeInputData(data string, tsABI string) (map[string]interface{}, error) {

	// load contract ABI
	abi, err := abi.JSON(strings.NewReader(tsABI))
	if err != nil {
		println("I fail here")
		return nil, err
	}

	/*
		decode method signature
		strip hex prefix (0x)
		signature is the first 32 bits of the has of the function
		in HEX 1 char = 4 bits, so 32 bits = 8 chars
	*/
	decodedSig, err := hex.DecodeString(data[2:10])
	if err != nil {
		return nil, err
	}

	// recover Method from signature and ABI
	method, err := abi.MethodById(decodedSig)
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

	// Unpack into struct:
	//
	//type Whatever struct {
	//	TradeValues [8]*big.Int
	//	TradeAddresses [4]common.Address
	//	V [2]uint8
	//	Rs [4][32]uint8
	//}
	//
	//var what Whatever
	//err = method.Inputs.Unpack(&what, decodedData)
	//if err != nil {
	//	log.Fatal(err)
	//}

	/*
		Decoding Solidity => Go

		uint256[8] => [8]*big.Int
		addresses[4] => [4]common.Address
		uint8[2] => [2]uint8
		bytes32[4] => [4][32]uint8 (array of 4 arrays of 32 uint8 each)
	*/

}
