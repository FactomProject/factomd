// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/web"
)

const (
	httpOK  = 200
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

		server.Post("/v2", HandleV2Post)
		server.Get("/v2", HandleV2Get)

		log.Print("Starting server")
		go server.Run(fmt.Sprintf("localhost:%d", state.GetPort()))
	}
}

func SetState(state interfaces.IState) {
	Servers[state.GetPort()].Env["state"] = state
	fmt.Println("API now directed to", state.GetFactomNodeName())
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

	req := primitives.NewJSON2Request(1, c.CommitChainMsg, "commit-chain")

	jsonResp, jsonError := HandleV2PostRequest(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	returnMsg(ctx, jsonResp.Result.(*CommitChainResponse).Message, true)
}

func HandleRevealChain(ctx *web.Context) {
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

	req := primitives.NewJSON2Request(1, c.CommitEntryMsg, "commit-entry")

	jsonResp, jsonError := HandleV2PostRequest(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	returnMsg(ctx, jsonResp.Result.(*CommitChainResponse).Message, true)
}

func HandleRevealEntry(ctx *web.Context) {
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

	req := primitives.NewJSON2Request(1, e.Entry, "reveal-entry")

	jsonResp, jsonError := HandleV2PostRequest(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	returnMsg(ctx, jsonResp.Result.(*RevealEntryResponse).Message, true)
}

func HandleDirectoryBlockHead(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, nil, "directory-block-head")

	jsonResp, jsonError := HandleV2GetRequest(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	dhead := new(DBHead)
	dhead.KeyMR = jsonResp.Result.(*DirectoryBlockHeadResponse).KeyMR
	returnMsg(ctx, dhead, true)
}

func HandleGetRaw(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, hashkey, "get-raw-data")

	jsonResp, jsonError := HandleV2GetRequest(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleGetReceipt(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, hashkey, "get-receipt")

	jsonResp, jsonError := HandleV2GetRequest(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleDirectoryBlock(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, hashkey, "directory-block-by-keymr")

	jsonResp, jsonError := HandleV2GetRequest(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	d := new(DBlock)

	d.Header.PrevBlockKeyMR = jsonResp.Result.(*DirectoryBlockResponse).Header.PrevBlockKeyMR
	d.Header.SequenceNumber = jsonResp.Result.(*DirectoryBlockResponse).Header.SequenceNumber
	d.Header.Timestamp = jsonResp.Result.(*DirectoryBlockResponse).Header.Timestamp
	d.EntryBlockList = jsonResp.Result.(*DirectoryBlockResponse).EntryBlockList

	returnMsg(ctx, d, true)
}

func HandleDirectoryBlockHeight(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, nil, "directory-block-height")

	jsonResp, jsonError := HandleV2GetRequest(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}

	returnMsg(ctx, jsonResp.Result.(*DirectoryBlockHeightResponse), true)
}

func HandleEntryBlock(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, hashkey, "entry-block-by-keymr")

	jsonResp, jsonError := HandleV2GetRequest(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
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

	req := primitives.NewJSON2Request(1, hashkey, "entry-by-hash")

	jsonResp, jsonError := HandleV2GetRequest(state, req)
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

func HandleChainHead(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, hashkey, "chain-head")

	jsonResp, jsonError := HandleV2GetRequest(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	d := new(CHead)

	d.ChainHead = jsonResp.Result.(*ChainHeadResponse).ChainHead

	returnMsg(ctx, d, true)
}

func HandleEntryCreditBalance(ctx *web.Context, eckey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, eckey, "entry-credit-balance")

	jsonResp, jsonError := HandleV2GetRequest(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	bal := fmt.Sprintf("%v", jsonResp.Result.(*EntryCreditBalanceResponse).Balance)
	fmt.Println(bal)
	returnMsg(ctx, bal, true)
}

func HandleGetFee(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, nil, "factoid-get-fee")

	jsonResp, jsonError := HandleV2GetRequest(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	type x struct{ Fee int64 }
	d := new(x)

	d.Fee = int64(jsonResp.Result.(*FactoidGetFeeResponse).Fee)

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

	req := primitives.NewJSON2Request(1, t.Transaction, "factoid-submit")

	jsonResp, jsonError := HandleV2PostRequest(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	r := new(x)
	r.Response = jsonResp.Result.(*FactoidSubmitResponse).Message
	r.Success = true
	returnMsg(ctx, r, true)
}

func HandleFactoidBalance(ctx *web.Context, eckey string) {
	type x struct {
		Response string
		Success  bool
	}
	t := new(x)

	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, eckey, "factoid-balance")

	jsonResp, jsonError := HandleV2GetRequest(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}

	t.Response = fmt.Sprint(jsonResp.Result.(*FactoidBalanceResponse).Balance)
	t.Success = true
	returnMsg(ctx, t, true)
}

func HandleProperties(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)
	fmt.Println("Connected to:",state.GetFactomNodeName())
	req := primitives.NewJSON2Request(1, nil, "properties")

	jsonResp, jsonError := HandleV2GetRequest(state, req)
	if jsonError != nil {
		returnV1(ctx, nil, jsonError)
		return
	}
	type x struct {
		Protocol_Version string
		Factomd_Version  string
	}
	d := new(x)
	d.Factomd_Version = jsonResp.Result.(*PropertiesResponse).FactomdVersion+" "+state.GetFactomNodeName()
	d.Protocol_Version = jsonResp.Result.(*PropertiesResponse).ProtocolVersion

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
