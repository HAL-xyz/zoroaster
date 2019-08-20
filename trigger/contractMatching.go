package trigger

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
	"strings"
)

// returns:
// - the string value of the contract call in case of match
// - "" otherwise
func MatchContract(client *ethrpc.EthRPC, tg *Trigger, blockNo int) string {

	methodId, err := encodeMethod(tg.MethodName, tg.ContractABI, tg.Inputs)
	if err != nil {
		log.Debug("cannot encode method: ", err)
		return ""
	}
	result, err := makeEthRpcCall(client, tg.ContractAdd, methodId, blockNo)
	if err != nil {
		log.Debug("rpc call failed: ", err)
		return ""
	}
	log.Debug("result from call is -> ", result)

	cond, ok := tg.Outputs[0].Condition.(ConditionOutput)
	if ok != true {
		log.Error("wrong wrong wrong")
		return ""
	}
	return validateContractReturnValue(tg.Outputs[0].ReturnType, result, cond)

}

func validateContractReturnValue(cnReturnType string, contractValue string, cond ConditionOutput) string {
	// all single u/integers
	intRgx := regexp.MustCompile(`u?int\d*$`)
	if intRgx.MatchString(cnReturnType) {
		ctVal := makeBigIntFromHex(contractValue)
		tgVal := makeBigInt(cond.Attribute)
		if validatePredBigInt(cond.Predicate, ctVal, tgVal) {
			return fmt.Sprintf("%v", ctVal)
		}
	}
	switch cnReturnType {
	case "Address":
		ctVal := common.HexToAddress(contractValue)
		tgVal := common.HexToAddress(cond.Attribute)
		if ctVal == tgVal {
			return strings.ToLower(ctVal.String())
		}
	case "bool":
		no, err := strconv.ParseInt(contractValue[2:], 16, 32)
		if err != nil {
			log.Debug(err)
			return ""
		}
		ctVal := false
		if no == 1 {
			ctVal = true
		}
		if validatePredBool(cond.Predicate, ctVal, cond.Attribute) {
			return fmt.Sprintf("%v", ctVal)
		}
	case "string":
		s, err := hex.DecodeString(strings.Replace(contractValue, "0x", "", 1))
		if err != nil {
			log.Debug(err)
			return ""
		}
		s = bytes.Replace(s, []byte("\x00"), []byte{}, -1)
		ss := stripCtlAndExtFromUTF8(string(s))[1:] // remove some this and a space (??)
		if ss == cond.Attribute {
			return ss
		}
	default:
		log.Debug("return type not supported: ", cnReturnType)
	}
	return ""
}

func makeEthRpcCall(client *ethrpc.EthRPC, cntAddress, data string, blockNumber int) (string, error) {

	params := ethrpc.T{
		To:   cntAddress,
		From: cntAddress,
		Data: data,
	}

	hexBlockNo := fmt.Sprintf("0x%x", blockNumber)

	ret, err := client.EthCall(params, hexBlockNo)
	if err != nil {
		return "", err
	}
	return ret, nil
}
