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
	"github.com/hoisie/web"
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
		server.Get("/v1/entry-block-by-keymr/([^/]+)", HandleEntryBlock)
		server.Get("/v1/entry-by-hash/([^/]+)", HandleEntry)
		server.Get("/v1/chain-head/([^/]+)", HandleChainHead)
		server.Get("/v1/entry-credit-balance/([^/]+)", HandleEntryCreditBalance)
		server.Get("/v1/factoid-balance/([^/]+)", HandleFactoidBalance)
		server.Get("/v1/factoid-get-fee/", HandleGetFee)

		server.Post("/v2", HandleV2)

		log.Print("Starting server")
		go server.Run(fmt.Sprintf("localhost:%d", state.GetPort()))
	}

}

func Stop(state interfaces.IState) {
	Servers[state.GetPort()].Close()
}

func handleV1Error(ctx *web.Context, err *primitives.JSONError) {
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

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleRevealChain(ctx *web.Context) {
	HandleRevealEntry(ctx)
}

func HandleCommitEntry(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, nil, "commit-entry")

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
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

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleDirectoryBlockHead(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, nil, "directory-block-head")

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleGetRaw(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, hashkey, "get-raw-data")

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleGetReceipt(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, hashkey, "get-receipt")

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleDirectoryBlock(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, hashkey, "directory-block-by-keymr")

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleEntryBlock(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, hashkey, "entry-block-by-keymr")

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleEntry(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, hashkey, "entry-by-hash")

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleChainHead(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, hashkey, "chain-head")

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleEntryCreditBalance(ctx *web.Context, eckey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, eckey, "entry-credit-balance")

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleGetFee(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, nil, "factoid-get-fee")

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleFactoidSubmit(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	type x struct{ Transaction string }
	t := new(x)

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

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
}

func HandleFactoidBalance(ctx *web.Context, eckey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req := primitives.NewJSON2Request(1, eckey, "factoid-balance")

	jsonResp, jsonError := HandleV2Request(state, req)
	returnV1(ctx, jsonResp, jsonError)
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
	r := rtn{Response: msg, Success: success}

	if p, err := json.Marshal(r); err != nil {
		wsLog.Error(err)
		return
	} else {
		ctx.Write(p)
	}
}
