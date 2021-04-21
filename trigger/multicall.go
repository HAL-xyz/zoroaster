package trigger

import (
	"fmt"
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/alethio/web3-go/ethrpc"
	"github.com/alethio/web3-multicall-go/multicall"
	"github.com/ethereum/go-ethereum/accounts/abi"
	log "github.com/sirupsen/logrus"
	"strings"
)

func MatchTriggersMulti(tgs []*Trigger, api tokenapi.ITokenAPI, blockNo int) ([]*CnMatch, []string) {

	resMap, err := runMulticallForTriggers(tgs, blockNo)
	if err != nil {
		log.Info(err)
		return []*CnMatch{}, []string{}
	}
	var tgsWithErrorsUUIDs []string

	var cnMatches []*CnMatch
	for _, tg := range tgs {
		res, found := resMap.Calls[tg.getKey()]
		if found && res.Success {
			match := matchTriggerWithResult(tg, res.Decoded, api)
			if match != nil {
				cnMatches = append(cnMatches, match)
			}
		}
		if found && !res.Success {
			tgsWithErrorsUUIDs = append(tgsWithErrorsUUIDs, tg.TriggerUUID)
		}
	}
	return cnMatches, tgsWithErrorsUUIDs
}

func runMulticallForTriggers(tgs []*Trigger, blockNo int) (*multicall.Result, error) {

	var views multicall.ViewCalls
	for _, tg := range tgs {
		v, err := makeViewFromTrigger(tg)
		if err == nil {
			views = append(views, v)
		}
	}

	cli, _ := ethrpc.NewWithDefaults(config.Zconf.EthNode)
	mc, err := multicall.New(cli)
	if err != nil {
		return nil, err
	}

	res, err := mc.Call(views, fmt.Sprintf("0x%x", blockNo))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func makeViewFromTrigger(tg *Trigger) (multicall.ViewCall, error) {

	viewMethod, err := makeViewMethod(tg)
	if err != nil {
		return multicall.ViewCall{}, err
	}

	var args []interface{}
	for _, input := range tg.Inputs {
		args = append(args, input.ParameterValue)
	}

	vc := multicall.NewViewCall(tg.getKey(), tg.ContractAdd, viewMethod, args)
	return vc, nil
}

func makeViewMethod(tg *Trigger) (string, error) {

	abiObj, err := abi.JSON(strings.NewReader(tg.ContractABI))
	if err != nil {
		return "", err
	}

	method, ok := abiObj.Methods[tg.FunctionName]
	if !ok {
		return "", fmt.Errorf("function %s not found", tg.FunctionName)
	}

	var inputTypes, outputTypes string

	for _, in := range method.Inputs {
		if len(inputTypes) != 0 {
			inputTypes += ","
		}
		inputTypes += in.Type.String()
	}

	for _, out := range method.Outputs {
		if len(outputTypes) != 0 {
			outputTypes += ","
		}
		outputTypes += out.Type.String()
	}

	return fmt.Sprintf("%s(%s)(%s)", tg.FunctionName, inputTypes, outputTypes), nil
}
