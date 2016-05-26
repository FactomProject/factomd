// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/web"
)

const (
	httpBad = 400
)

var Servers map[int]*web.Server

func Start(state interfaces.IState) {
	var server *web.Server

	if Servers == nil {
		Servers = make(map[int]*web.Server)
	}

	if Servers[state.GetPort()] == nil {
		server = web.NewServer()
		Servers[state.GetPort()] = server
		server.Env["state"] = state

		server.Post("/v1/factoid-submit/?", HandleFactoidSubmit)
		server.Post("/v1/commit-chain/?", HandleCommitChain)
		server.Post("/v1/reveal-chain/?", HandleRevealChain)
		server.Post("/v1/commit-entry/?", HandleCommitEntry)
		server.Post("/v1/reveal-entry/?", HandleRevealEntry)
		server.Get("/v1/directory-block-head/?", HandleDirectoryBlockHead)
		server.Get("/v1/get-raw-data/([^/]+)", HandleGetRaw)
		server.Get("/v1/get-receipt/([^/]+)", HandleGetReceipt)
		server.Get("/v1/directory-block-by-keymr/([^/]+)", HandleDirectoryBlock)
		server.Get("/v1/directory-block-height/?", HandleDirectoryBlockHeight)
		server.Get("/v1/entry-block-by-keymr/([^/]+)", HandleEntryBlock)
		server.Get("/v1/entry-by-hash/([^/]+)", HandleEntry)
		server.Get("/v1/chain-head/([^/]+)", HandleChainHead)
		server.Get("/v1/entry-credit-balance/([^/]+)", HandleEntryCreditBalance)
		server.Get("/v1/factoid-balance/([^/]+)", HandleFactoidBalance)
		server.Get("/v1/factoid-get-fee/", HandleGetFee)
		server.Get("/v1/properties/", HandleProperties)

		server.Post("/v2", HandleV2)
		server.Get("/v2", HandleV2)

		log.Print("Starting server")
		go server.Run(fmt.Sprintf(":%d", state.GetPort()))
	}
}

func SetState(state interfaces.IState) {
	wait := func() {
		for Servers == nil && Servers[state.GetPort()] != nil {
			time.Sleep(10 * time.Millisecond)
		}
		Servers[state.GetPort()].Env["state"] = state
		os.Stderr.WriteString("API now directed to " + state.GetFactomNodeName() + "\n")
	}
	go wait()
}

func Stop(state interfaces.IState) {
	Servers[state.GetPort()].Close()
}

func handleV1Error(ctx *web.Context, err *primitives.JSONError) {
	if err.Data != nil {
		data, ok := err.Data.(string)
		if ok == true {
			returnMsg(ctx, err.Message+": "+data, false)
			return
		}
	}
	returnMsg(ctx, err.Message, false)
	return
}

func returnV1(ctx *web.Context, jsonResp *primitives.JSON2Response, jsonError *primitives.JSONError) {
	if jsonError != nil {
		handleV1Error(ctx, jsonError)
		return
	}
	returnMsg(ctx, jsonResp.Result, true)
}

func HandleCommitChain(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	type commitchain struct {
		CommitChainMsg string
	}
	c := new(commitchain)
	if p, err := ioutil.ReadAll(ctx.Request.Body); err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	} else {
		if err := json.Unmarshal(p, c); err != nil {
			handleV1Error(ctx, NewInvalidParamsError())
			return
		}
	}
	param := MessageRequest{Message: c.CommitChainMsg}
	req := primitives.NewJSON2Request("commit-chain", 1, param)
	_, jsonError := HandleV2Request(state, req)

	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	//log.Print(jsonResp.Result.(*RevealEntryResponse).Message)

	// this is the blank '200 ok' that is returned for V1
	returnV1Msg(ctx, "", false)

}

func HandleRevealChain(ctx *web.Context) {
	fmt.Println("RevealChain")
	HandleRevealEntry(ctx)
}

func HandleCommitEntry(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	type commitentry struct {
		CommitEntryMsg string
	}

	c := new(commitentry)
	if p, err := ioutil.ReadAll(ctx.Request.Body); err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	} else {
		if err := json.Unmarshal(p, c); err != nil {
			handleV1Error(ctx, NewInvalidParamsError())
			return
		}
	}

	param := EntryRequest{Entry: c.CommitEntryMsg}
	req := primitives.NewJSON2Request("commit-entry", 1, param)

	_, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	//log.Print( jsonResp.Result.(*RevealEntryResponse).Message)

	// this is the blank '200 ok' that is returned for V1
	returnV1Msg(ctx, "", true)

}

func HandleRevealEntry(ctx *web.Context) {
	fmt.Println("RevealEntry")
	state := ctx.Server.Env["state"].(interfaces.IState)
	type revealentry struct {
		Entry string
	}

	e := new(revealentry)
	if p, err := ioutil.ReadAll(ctx.Request.Body); err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	} else {
		if err := json.Unmarshal(p, e); err != nil {
			handleV1Error(ctx, NewInvalidParamsError())
			return
		}
	}

	param := EntryRequest{Entry: e.Entry}
	req := primitives.NewJSON2Request("reveal-entry", 1, param)

	_, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	//log.Print(jsonResp.Result.(*RevealEntryResponse).Message)
	returnV1Msg(ctx, "", true)

}

func HandleDirectoryBlockHead(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request("directory-block-head", 1, nil)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	tmp, err := json.Marshal(jsonResp.Result)
	resp := string(tmp)
	if err != nil {
		resp = "{\"KeyMR\",0}"
		returnV1Msg(ctx, resp, true)
	}

	resp = strings.Replace(resp, "keymr", "KeyMR", -1)

	returnV1Msg(ctx, resp, true)
}

func HandleGetRaw(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	param := HashRequest{Hash: hashkey}
	req := primitives.NewJSON2Request("raw-data", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleGetReceipt(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	param := HashRequest{Hash: hashkey}
	req := primitives.NewJSON2Request("get-receipt", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleDirectoryBlock(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)
	param := KeyMRRequest{KeyMR: hashkey}
	req := primitives.NewJSON2Request("directory-block", 1, param)
	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}

	type DBlock struct {
		Header struct {
			PrevBlockKeyMR string
			SequenceNumber int64
			Timestamp      int64
		}
		EntryBlockList []EBlockAddr
	}
	d := new(DBlock)

	d.Header.PrevBlockKeyMR = jsonResp.Result.(*DirectoryBlockResponse).Header.PrevBlockKeyMR
	d.Header.SequenceNumber = jsonResp.Result.(*DirectoryBlockResponse).Header.SequenceNumber
	d.Header.Timestamp = jsonResp.Result.(*DirectoryBlockResponse).Header.Timestamp
	d.EntryBlockList = jsonResp.Result.(*DirectoryBlockResponse).EntryBlockList

	// conflict if I use local structs.  using a string replace on the structs that would be pointer handled (*DirectoryBlockResponse)
	bResp, err := json.Marshal(d)
	if err != nil {

		returnMsg(ctx, d, true)
	}
	resp := string(bResp)
	resp = strings.Replace(resp, "{\"chainid\"", "{\"ChainID\"", -1)
	resp = strings.Replace(resp, ",\"keymr\":", ",\"KeyMR\":", -1)

	returnV1Msg(ctx, resp, true)
}

func HandleDirectoryBlockHeight(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request("directory-block-height", 1, nil)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	tmp, err := json.Marshal(jsonResp.Result)
	resp := string(tmp)
	if err != nil {
		resp = "{\"Height\",0}"
		returnV1Msg(ctx, resp, true)
	}

	type DirectoryBlockHeightResponse struct {
		Height int64 /*`json:"height"` V1 doesn't use the json tye def */
	}

	resp = strings.Replace(resp, "height", "Height", -1)

	returnV1Msg(ctx, resp, true)
}

func HandleEntryBlock(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	param := KeyMRRequest{KeyMR: hashkey}
	req := primitives.NewJSON2Request("entry-block", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}

	type EBlock struct {
		Header struct {
			BlockSequenceNumber int64
			ChainID             string
			PrevKeyMR           string
			Timestamp           int64
			DBHeight            int64
		}
		EntryList []EntryAddr
	}
	d := new(EBlock)

	d.Header.BlockSequenceNumber = jsonResp.Result.(*EntryBlockResponse).Header.BlockSequenceNumber
	d.Header.ChainID = jsonResp.Result.(*EntryBlockResponse).Header.ChainID
	d.Header.PrevKeyMR = jsonResp.Result.(*EntryBlockResponse).Header.PrevKeyMR
	d.Header.Timestamp = jsonResp.Result.(*EntryBlockResponse).Header.Timestamp
	d.Header.DBHeight = jsonResp.Result.(*EntryBlockResponse).Header.DBHeight
	d.EntryList = jsonResp.Result.(*EntryBlockResponse).EntryList

	returnMsg(ctx, d, true)
}

func HandleEntry(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	param := HashRequest{Hash: hashkey}
	req := primitives.NewJSON2Request("entry-by-hash", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	d := new(EntryStruct)

	d.ChainID = jsonResp.Result.(*EntryResponse).ChainID
	d.Content = jsonResp.Result.(*EntryResponse).Content
	d.ExtIDs = jsonResp.Result.(*EntryResponse).ExtIDs

	returnMsg(ctx, d, true)
}

func HandleChainHead(ctx *web.Context, chainid string) {

	state := ctx.Server.Env["state"].(interfaces.IState)
	param := ChainIDRequest{ChainID: chainid}
	req := primitives.NewJSON2Request("chain-head", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return

	}
	// restatement of chead from structs file
	// v1 doesn't like the lcase in v2'
	type CHead struct {
		ChainHead string
	}

	d := new(CHead)
	d.ChainHead = jsonResp.Result.(*ChainHeadResponse).ChainHead
	returnMsg(ctx, d, true)
}

func HandleEntryCreditBalance(ctx *web.Context, address string) {
	type x struct {
		Response string
		Success  bool
	}

	state := ctx.Server.Env["state"].(interfaces.IState)

	param := AddressRequest{Address: address}
	req := primitives.NewJSON2Request("entry-credit-balance", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}

	t := new(x)
	t.Response = fmt.Sprint(jsonResp.Result.(*EntryCreditBalanceResponse).Balance)
	t.Success = true
	returnMsg(ctx, t, true)
}

func HandleGetFee(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request("factoid-fee", 1, nil)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	type x struct{ Fee int64 }
	d := new(x)

	d.Fee = int64(jsonResp.Result.(*FactoidFeeResponse).Fee)

	returnMsg(ctx, d, true)
}

func HandleFactoidSubmit(ctx *web.Context) {
	type x struct {
		Response string
		Success  bool
	}

	type transaction struct{ Transaction string }
	t := new(transaction)

	state := ctx.Server.Env["state"].(interfaces.IState)

	var p []byte
	var err error
	if p, err = ioutil.ReadAll(ctx.Request.Body); err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	} else {
		if err := json.Unmarshal(p, t); err != nil {
			handleV1Error(ctx, NewInvalidParamsError())
			return
		}
	}

	param := TransactionRequest{Transaction: t.Transaction}
	req := primitives.NewJSON2Request("factoid-submit", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	r := new(x)
	r.Response = jsonResp.Result.(*FactoidSubmitResponse).Message
	r.Success = true
	returnMsg(ctx, r, true)
}

func HandleFactoidBalance(ctx *web.Context, address string) {

	type x struct {
		Response string
		Success  bool
	}
	state := ctx.Server.Env["state"].(interfaces.IState)
	param := AddressRequest{Address: address}
	req := primitives.NewJSON2Request("factoid-balance", 1, param)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}

	r := new(x)
	r.Response = fmt.Sprint(jsonResp.Result.(*FactoidBalanceResponse).Balance)
	r.Success = true
	returnMsg(ctx, r, true)

}

func HandleProperties(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)
	fmt.Println("Connected to:", state.GetFactomNodeName())
	req := primitives.NewJSON2Request("properties", 1, nil)

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	type x struct {
		Protocol_Version string
		Factomd_Version  string
	}
	d := new(x)
	d.Factomd_Version = jsonResp.Result.(*PropertiesResponse).FactomdVersion
	d.Protocol_Version = "0.0.0.0" // meaningless after v1
	returnMsg(ctx, d, true)
}

/*********************************************************
 * Support Functions
 *********************************************************/

func returnMsg(ctx *web.Context, msg interface{}, success bool) {
	type rtn struct {
		Response interface{}
		Success  bool
	}
	/*str, ok:=msg.(string)
	if ok == false {
		var err error
		str, err = primitives.EncodeJSONString(msg)
		if err != nil {
			wsLog.Error(err)
			return
		}
	}*/
	r := msg

	if p, err := json.Marshal(r); err != nil {
		wsLog.Error(err)
		return
	} else {
		ctx.Write(p)
	}
}

func returnV1Msg(ctx *web.Context, msg string, success bool) {

	/* V1 requires call specific case changes that can't be handled with
	interfaces for example.  Block Height needs to return  height as the json item name
	in golang, lower case names are private so won't be returned.
	Deal with the responses in the call specific v1 handlers until they are depricated.
	*/
	bMsg := []byte(msg)
	ctx.Write(bMsg)

}
