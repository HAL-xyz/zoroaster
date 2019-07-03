package trigger

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"strconv"
)

func MatchContract(client *ethrpc.EthRPC, cntAddress string, tg *Trigger, blockNo int) bool {

	methodId, err := EncodeMethod(tg.MethodName, tg.ContractABI, tg.Inputs)
	if err != nil {
		log.Debug("cannot encode method: ", err)
		return false
	}
	result, err := makeEthRpcCall(client, cntAddress, methodId, blockNo)
	if err != nil {
		log.Debug("rpc call failed: ", err)
		return false
	}
	log.Debug("result from call is -> ", result)

	cond, ok := tg.Outputs[0].Condition.(ConditionOutput)
	if ok != true {
		log.Error("wrong wrong wrong")
		return false
	}

	returnType := tg.Outputs[0].ReturnType

	if returnType == "Address" {
		ctVal := common.HexToAddress(result)
		tgVal := common.HexToAddress(cond.Attribute)
		return ctVal == tgVal
	}
	if returnType == "uint256" {
		ctVal := makeBigInt16(result)
		tgVal := makeBigInt(cond.Attribute)
		return validatePredBigInt(cond.Predicate, ctVal, tgVal)
	}
	if returnType == "bool" {
		no, err := strconv.ParseInt(result[2:], 16, 32)
		if err != nil {
			log.Debug(err)
			return false
		}
		ctVal := false
		if no == 1 {
			ctVal = true
		}
		return validatePredBool(cond.Predicate, ctVal, cond.Attribute)
	}
	log.Debug("return type not supported:", returnType)
	return false
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
