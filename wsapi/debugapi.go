// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/PaulSnow/factom2d/common/globals"

	"regexp"

	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
)

type success struct {
	Status string `json:"status"`
}

func HandleDebug(writer http.ResponseWriter, request *http.Request) {
	_ = globals.Params

	state, err := GetState(request)
	if err != nil {
		wsDebugLog.Errorf("failed to extract port from request: %s", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := checkAuthHeader(state, request); err != nil {
		handleUnauthorized(request, writer)
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		HandleV2Error(writer, nil, NewInvalidRequestError())
		return
	}

	j, err := primitives.ParseJSON2Request(string(body))
	if err != nil {
		HandleV2Error(writer, nil, NewInvalidRequestError())
		return
	}

	jsonResp, jsonError := HandleDebugRequest(state, j)

	if jsonError != nil {
		HandleV2Error(writer, j, jsonError)
		return
	}

	writer.Write([]byte(jsonResp.String()))
}

func HandleDebugRequest(state interfaces.IState, j *primitives.JSON2Request) (*primitives.JSON2Response, *primitives.JSONError) {
	var resp interface{}
	var jsonError *primitives.JSONError
	params := j.Params
	wsDebugLog.Printf("request %v", j.String())

	switch j.Method {
	case "audit-servers":
		resp, jsonError = HandleAuditServers(state, params)
		break
	case "authorities":
		resp, jsonError = HandleAuthorities(state, params)
		break
	case "configuration":
		resp, jsonError = HandleConfig(state, params)
		break
	case "current-minute":
		resp, jsonError = HandleCurrentMinute(state, params)
		break
	case "delay":
		resp, jsonError = HandleDelay(state, params)
		break
	case "set-delay":
		resp, jsonError = HandleSetDelay(state, params)
		break
	case "drop-rate":
		resp, jsonError = HandleDropRate(state, params)
		break
	case "set-drop-rate":
		resp, jsonError = HandleSetDropRate(state, params)
		break
	case "federated-servers":
		resp, jsonError = HandleFedServers(state, params)
		break
	case "holding-queue":
		resp, jsonError = HandleHoldingQueue(state, params)
		break
	case "messages":
		resp, jsonError = HandleMessages(state, params)
		break
	case "network-info":
		resp, jsonError = HandleNetworkInfo(state, params)
		break
	case "summary":
		resp, jsonError = HandleSummary(state, params)
		break
	case "predictive-fer":
		resp, jsonError = HandlePredictiveFER(state, params)
		break
	case "process-list":
		resp, jsonError = HandleProcessList(state, params)
		break
	case "write-configuration":
		resp, jsonError = HandleWriteConfig(state, params)
		break
	case "reload-configuration":
		resp, jsonError = HandleReloadConfig(state, params)
		break
	case "sim-ctrl":
		resp, jsonError = HandleSimControl(state, params)
		break
	case "wait-blocks":
		resp, jsonError = HandleWaitBlocks(state, params)
		break
	case "wait-for-block":
		resp, jsonError = HandleWaitForBlock(state, params)
		break
	case "wait-minutes":
		resp, jsonError = HandleWaitMinutes(state, params)
		break
	case "wait-for-minute":
		resp, jsonError = HandleWaitForMinute(state, params)
		break
	case "message-filter":
		resp, jsonError = HandleMessageFilter(state, params)
		break
	default:
		jsonError = NewMethodNotFoundError()
		break
	}
	if jsonError != nil {
		wsDebugLog.Printf("error %v", jsonError)
		return nil, jsonError
	}

	jsonResp := primitives.NewJSON2Response()
	jsonResp.ID = j.ID
	jsonResp.Result = resp
	wsDebugLog.Printf("response %v", jsonResp.String())

	return jsonResp, nil
}

func HandleAuditServers(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	type ret struct {
		AuditServers []interfaces.IServer
	}
	r := new(ret)

	r.AuditServers = state.GetAuditServers(state.GetLeaderHeight())
	return r, nil
}

func HandleAuthorities(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	type ret struct {
		Authorities []interfaces.IAuthority `json: "authorities"`
	}
	r := new(ret)

	r.Authorities = state.GetAuthorities()
	return r, nil
}

func HandleConfig(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	return state.GetCfg(), nil
}

func HandleCurrentMinute(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	type ret struct {
		Minute int
	}
	r := new(ret)

	r.Minute = state.GetCurrentMinute()
	return r, nil
}

func HandleDelay(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	type ret struct {
		Delay int64
	}
	r := new(ret)

	r.Delay = state.GetDelay()
	return r, nil
}

func HandleSetDelay(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	type ret struct {
		Delay int64
	}
	r := new(ret)

	delay := new(SetDelayRequest)
	err := MapToObject(params, delay)
	if err != nil {
		return nil, NewInvalidParamsError()
	}

	state.SetDelay(delay.Delay)
	r.Delay = delay.Delay

	return r, nil
}

func HandleDropRate(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	type ret struct {
		DropRate int
	}
	r := new(ret)

	r.DropRate = state.GetDropRate()
	return r, nil
}

func HandleSetDropRate(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	type ret struct {
		DropRate int
	}
	r := new(ret)

	droprate := new(SetDropRateRequest)
	err := MapToObject(params, droprate)
	if err != nil {
		return nil, NewInvalidParamsError()
	}

	state.SetDropRate(droprate.DropRate)
	r.DropRate = droprate.DropRate
	return r, nil
}

func HandleFedServers(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	type ret struct {
		FederatedServers []interfaces.IServer
	}
	r := new(ret)

	r.FederatedServers = state.GetFedServers(state.GetLeaderHeight())
	return r, nil
}

func HandleHoldingQueue(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	type ret struct {
		Messages []interfaces.IMsg
	}
	r := new(ret)

	for _, v := range state.LoadHoldingMap() {
		r.Messages = append(r.Messages, v)
	}
	return r, nil
}

func HandleMessages(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	type ret struct {
		Messages []json.RawMessage
	}
	r := new(ret)
	for _, v := range state.GetJournalMessages() {
		r.Messages = append(r.Messages, v)
	}
	return r, nil
}

func HandleNetworkInfo(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	type ret struct {
		NodeName      string
		Role          string
		NetworkNumber int
		NetworkName   string
		NetworkID     uint32
	}
	r := new(ret)
	r.NodeName = state.GetFactomNodeName()
	r.Role = getRole(state)
	r.NetworkNumber = state.GetNetworkNumber()
	r.NetworkName = state.GetNetworkName()
	r.NetworkID = state.GetNetworkID()
	return r, nil
}

func HandleSummary(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	type ret struct {
		Summary string
	}
	r := new(ret)
	r.Summary = state.ShortString()

	return r, nil
}

func HandlePredictiveFER(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	type ret struct {
		PredictiveFER uint64
	}
	r := new(ret)
	r.PredictiveFER = state.GetPredictiveFER()
	return r, nil
}

func HandleProcessList(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	type ret struct {
		ProcessList string
	}
	r := new(ret)
	r.ProcessList = state.GetLeaderPL().String()
	return r, nil
}

func HandleReloadConfig(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	// LoacConfig with "" strings should load the default location
	state.LoadConfig(state.GetConfigPath(), state.GetNetworkName())

	return state.GetCfg(), nil
}

func HandleWriteConfig(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	fmt.Sprintf("WRITE_CONFIG: %v", params)

	type testCfg struct {
		Config string
	}

	newCfg := new(testCfg)
	MapToObject(params, newCfg)

	cfgPath := state.GetConfigPath()
	f, err := os.Create(cfgPath)
	if err == nil {
		f.WriteString(fmt.Sprintf("%s", newCfg.Config))
	}

	return new(success), nil
}

func runCmd(cmd string) {
	//os.Stderr.WriteString("Executing: " + cmd + "\n")
	globals.InputChan <- cmd

	return
}

func HandleSimControl(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	cmdLines := new(GetCommands)
	err := MapToObject(params, cmdLines)
	if err != nil {
		return nil, NewInvalidParamsError()
	}

	for _, cmdStr := range cmdLines.Commands {
		runCmd(cmdStr)
	}

	r := new(success)
	r.Status = "Success!"
	return r, nil
}

type SetDelayRequest struct {
	Delay int64 `json:"delay"`
}

type SetDropRateRequest struct {
	DropRate int `json:"droprate"`
}

type GetCommands struct {
	Commands []string `json:"commands"`
}

func HandleMessageFilter(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	wsDebugLog.Println("Factom Node Name: ", state.GetFactomNodeName())
	x, ok := params.(map[string]interface{})
	if !ok {
		return nil, NewCustomInvalidParamsError("ERROR! Invalid params passed in")
	}

	wsDebugLog.Println(`x["output-regex"]`, x["output-regex"])
	wsDebugLog.Println(`x["input-regex"]`, x["input-regex"])

	OutputString := fmt.Sprintf("%s", x["output-regex"])
	if OutputString != "" {
		OutputRegEx := regexp.MustCompile(OutputString)
		state.PassOutputRegEx(OutputRegEx, OutputString)

	} else if OutputString == "off" {
		state.PassOutputRegEx(nil, "")
	}

	InputString := fmt.Sprintf("%s", x["input-regex"])
	if InputString != "" {
		InputRegEx := regexp.MustCompile(InputString)
		state.PassInputRegEx(InputRegEx, InputString)

	} else if InputString == "off" {
		state.PassInputRegEx(nil, "")
	}

	h := new(MessageFilter)
	h.Params = "Success"

	return h, nil
}

func getParamMap(params interface{}) (x map[string]interface{}, ok bool) {
	x, ok = params.(map[string]interface{})
	return x, ok
}

func HandleWaitMinutes(s interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	x, _ := getParamMap(params)
	min := int(x["minutes"].(float64))
	newTime := int(s.GetLLeaderHeight())*10 + s.GetCurrentMinute() + min
	newBlock := newTime / 10
	newMinute := newTime % 10
	waitForQuiet(s, newBlock, newMinute)

	r := new(success)
	r.Status = "Success!"
	return r, nil
}

func HandleWaitBlocks(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {

	x, _ := getParamMap(params)
	blks := int(x["blocks"].(float64))
	waitForQuiet(state, blks+int(state.GetLLeaderHeight()), 0)

	r := new(success)
	r.Status = "Success!"
	return r, nil
}

func HandleWaitForBlock(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {

	x, _ := getParamMap(params)
	waitForQuiet(state, int(x["block"].(float64)), 0)

	r := new(success)
	r.Status = "Success!"
	return r, nil
}

func HandleWaitForMinute(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {

	x, _ := getParamMap(params)
	newMinute := int(x["minute"].(float64))
	if newMinute > 10 {
		panic("invalid minute")
	}
	newBlock := int(state.GetLLeaderHeight())
	if state.GetCurrentMinute() > newMinute {
		newBlock++
	}
	waitForQuiet(state, newBlock, newMinute)

	r := new(success)
	r.Status = "Success!"
	return r, nil
}

func waitForQuiet(s interfaces.IState, newBlock int, newMinute int) {
	//	fmt.Printf("%s: %d-:-%d WaitFor(%d-:-%d)\n", s.FactomNodeName, s.LLeaderHeight, s.CurrentMinute, newBlock, newMinute)
	sleepTime := time.Duration(globals.Params.BlkTime) * 1000 / 40 // Figure out how long to sleep in milliseconds
	if newBlock*10+newMinute < int(s.GetLLeaderHeight())*10+s.GetCurrentMinute() {
		panic("Wait for the past")
	}

	fmt.Printf("wait for quiet : %v", newBlock)

	for int(s.GetLLeaderHeight()) < newBlock {
		x := int(s.GetLLeaderHeight())
		// wait for the next block
		for int(s.GetLLeaderHeight()) == x {
			time.Sleep(sleepTime * time.Millisecond) // wake up and about 4 times per minute
		}
	}

	// wait for the right minute
	for s.GetCurrentMinute() != newMinute {
		time.Sleep(sleepTime * time.Millisecond) // wake up and about 4 times per minute
	}
}

func getRole(s interfaces.IState) string {
	feds := s.GetFedServers(s.GetLLeaderHeight())
	for _, fed := range feds {
		if s.GetIdentityChainID().IsSameAs(fed.GetChainID()) {
			return "Leader"
		}
	}

	audits := s.GetAuditServers(s.GetLLeaderHeight())
	for _, aud := range audits {
		if s.GetIdentityChainID().IsSameAs(aud.GetChainID()) {
			return "Audit"
		}
	}
	return "Follower"
}
