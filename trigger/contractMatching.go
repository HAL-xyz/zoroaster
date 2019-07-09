package trigger

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"strconv"
)

// returns the string value of the contract call outcome in case of match, "" otherwise
func MatchContract(client *ethrpc.EthRPC, tg *Trigger, blockNo int) string {

	methodId, err := EncodeMethod(tg.MethodName, tg.ContractABI, tg.Inputs)
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

	returnType := tg.Outputs[0].ReturnType

	if returnType == "Address" {
		ctVal := common.HexToAddress(result)
		tgVal := common.HexToAddress(cond.Attribute)
		if ctVal == tgVal {
			return ctVal.String()
		}
	}
	if returnType == "uint256" {
		ctVal := makeBigInt16(result)
		tgVal := makeBigInt(cond.Attribute)
		if validatePredBigInt(cond.Predicate, ctVal, tgVal) {
			return fmt.Sprintf("%v", ctVal)
		}
	}
	if returnType == "bool" {
		no, err := strconv.ParseInt(result[2:], 16, 32)
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
	}
	log.Debug("return type not supported:", returnType)
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
