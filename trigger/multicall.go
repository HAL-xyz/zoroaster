package trigger

import (
	"fmt"
	"github.com/HAL-xyz/web3-multicall-go/multicall"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
)

func MatchTriggersMulti(tgs []*Trigger, api tokenapi.ITokenAPI, blockNo int) ([]*CnMatch, []string) {

	resMap, err := runMulticallForTriggers(tgs, blockNo, api)
	if err != nil {
		log.Warnf("MatchTriggersMulti failed: %s", err)
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

func runMulticallForTriggers(tgs []*Trigger, blockNo int, api tokenapi.ITokenAPI) (*multicall.Result, error) {

	mc, err := multicall.New(api.GetRPCCli())
	if err != nil {
		return nil, fmt.Errorf("create mu cli: %s", err)
	}

	views := makeDistinctViews(tgs)
	log.Info("MUL Distinct views: ", len(views))

	// Geth has a hardcoded 5seconds timeout, so we batch our call in chunks and exec them in async
	var finalRes multicall.Result
	finalRes.Calls = make(map[string]multicall.CallResult, len(views))

	chunks := chunkViews(views, 50)

	chunkResults := make(chan *multicall.Result, len(chunks))
	chunkErrors := make(chan error, len(chunks))
	var wg sync.WaitGroup

	for _, chunk := range chunks {
		wg.Add(1)
		go func(cn []multicall.ViewCall) {
			defer wg.Done()
			log.Debug("Executing batch of ", len(cn))
			chunkResult, err := mc.Call(cn, fmt.Sprintf("0x%x", blockNo))
			if err != nil {
				chunkErrors <- fmt.Errorf("mc call failed: %s", err)
			}
			chunkResults <- chunkResult
		}(chunk)
	}
	wg.Wait()
	close(chunkErrors)
	close(chunkResults)

	log.Infof("mul calls succeeded: %d; failed: %d ", len(chunkResults), len(chunkErrors))

	if len(chunkErrors) != 0 {
		return nil, <-chunkErrors
	}
	for r := range chunkResults {
		for k, v := range r.Calls {
			finalRes.Calls[k] = v
		}
	}
	log.Debug("Total no of calls: ", len(finalRes.Calls))

	return &finalRes, nil
}

func makeViewFromTrigger(tg *Trigger) (multicall.ViewCall, error) {

	viewMethod, err := makeViewMethod(tg)
	if err != nil {
		return multicall.ViewCall{}, err
	}

	inputs := make([]tokenapi.Input, len(tg.Inputs))
	for i, tgin := range tg.Inputs {
		inputs[i].ParameterValue = tgin.ParameterValue
		inputs[i].ParameterType = tgin.ParameterType
	}
	args, err := tokenapi.MakeObjectsFromInput(inputs)

	vc := multicall.NewViewCall(tg.getKey(), strings.ToLower(tg.ContractAdd), viewMethod, args)
	return vc, vc.Validate()
}

func makeViewMethod(tg *Trigger) (string, error) {

	abiObj, err := tg.getABIObj()
	if err != nil {
		return "", fmt.Errorf("invalid abi for tg %s", tg.TriggerUUID)
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

func chunkViews(slice []multicall.ViewCall, chunkSize int) [][]multicall.ViewCall {

	var chunks = make([][]multicall.ViewCall, 0)

	beg := 0
	for beg+chunkSize < len(slice) {
		chunks = append(chunks, slice[beg:beg+chunkSize])
		beg += chunkSize
	}
	chunks = append(chunks, slice[beg:])

	return chunks
}

// triggers with the same key have identical call arguments, so we treat them as such
func makeDistinctViews(tgs []*Trigger) multicall.ViewCalls {
	var distinctViews = make(map[string]struct{})
	var views multicall.ViewCalls
	for _, tg := range tgs {
		_, found := distinctViews[tg.getKey()]
		if !found {
			v, err := makeViewFromTrigger(tg)
			if err == nil {
				views = append(views, v)
				distinctViews[tg.getKey()] = struct{}{}
			}
		}
	}
	return views
}
