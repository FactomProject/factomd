// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
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

	req:=primitives.NewJSON2Request(1, c.CommitChainMsg, "commit-chain")

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		handleV1Error(ctx, jsonError)
		return
	}
	returnMsg(ctx, jsonResp.Result, true)
}

func HandleRevealChain(ctx *web.Context) {
	HandleRevealEntry(ctx)
}

func HandleCommitEntry(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req:=primitives.NewJSON2Request(1, nil, "commit-entry")

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		handleV1Error(ctx, jsonError)
		return
	}
	returnMsg(ctx, jsonResp.Result, true)
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

	req:=primitives.NewJSON2Request(1, e.Entry, "reveal-entry")

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		handleV1Error(ctx, jsonError)
		return
	}
	returnMsg(ctx, jsonResp.Result, true)
}

func HandleDirectoryBlockHead(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	req:=primitives.NewJSON2Request(1, nil, "directory-block-head")

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		handleV1Error(ctx, jsonError)
		return
	}
	returnMsg(ctx, jsonResp.Result, true)
}

func HandleGetRaw(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	}

	req:=primitives.NewJSON2Request(1, h, "get-raw-data")

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		handleV1Error(ctx, jsonError)
		return
	}
	returnMsg(ctx, jsonResp.Result, true)
}

func HandleGetReceipt(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	}

	req:=primitives.NewJSON2Request(1, h, "get-receipt")

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		handleV1Error(ctx, jsonError)
		return
	}
	returnMsg(ctx, jsonResp.Result, true)
}

func HandleDirectoryBlock(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	}

	req:=primitives.NewJSON2Request(1, h, "directory-block-by-keymr")

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		handleV1Error(ctx, jsonError)
		return
	}

	returnMsg(ctx, jsonResp.Result, true)
}

func HandleEntryBlock(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	}

	req:=primitives.NewJSON2Request(1, h, "entry-block-by-keymr")

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		handleV1Error(ctx, jsonError)
		return
	}

	returnMsg(ctx, jsonResp.Result, true)
}

func HandleEntry(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	}

	req:=primitives.NewJSON2Request(1, h, "entry-by-hash")

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		handleV1Error(ctx, jsonError)
		return
	}

	returnMsg(ctx, jsonResp.Result, true)
}

func HandleChainHead(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	c := new(CHead)

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	}

	req:=primitives.NewJSON2Request(1, h, "chain-head")

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		handleV1Error(ctx, jsonError)
		return
	}

	c.ChainHead = jsonResp.Result.(string)

	returnMsg(ctx, c, true)
}

func HandleEntryCreditBalance(ctx *web.Context, eckey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	adr, err := primitives.HexToHash(eckey)
	if err != nil {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	}

	req:=primitives.NewJSON2Request(1, adr, "entry-credit-balance")

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		handleV1Error(ctx, jsonError)
		return
	}

	returnMsg(ctx, jsonResp.Result, true)
}

func HandleGetFee(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	type x struct{ Fee int64 }

	b := new(x)

	b.Fee = int64(state.GetFactoidState().GetFactoshisPerEC())

	if p, err := json.Marshal(b); err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		return
	} else {
		ctx.Write(p)
	}
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

	req:=primitives.NewJSON2Request(1, t.Transaction, "factoid-submit")

	jsonResp, jsonError := HandleV2Request(state, req)
	if jsonError != nil {
		handleV1Error(ctx, jsonError)
		return
	}
	returnMsg(ctx, jsonResp.Result, true)
}

func HandleFactoidBalance(ctx *web.Context, eckey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	var b FactoidBalance
	adr, err := hex.DecodeString(eckey)
	if err == nil && len(adr) != constants.HASH_LENGTH {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	}
	if err == nil {
		v := int64(state.GetFactoidState().GetFactoidBalance(factoid.NewAddress(adr).Fixed()))
		str := fmt.Sprintf("%d", v)
		b = FactoidBalance{Response: str, Success: true}
	} else {
		handleV1Error(ctx, NewInvalidParamsError())
		return
	}

	if p, err := json.Marshal(b); err != nil {
		wsLog.Error(err)
		return
	} else {
		ctx.Write(p)
	}
}

/*********************************************************
 * Support Functions
 *********************************************************/

func returnMsg(ctx *web.Context, msg interface{}, success bool) {
	type rtn struct {
		Response string
		Success  bool
	}
	str, ok:=msg.(string)
	if ok == false {
		var err error
		str, err = primitives.EncodeJSONString(msg)
		if err != nil {
			wsLog.Error(err)
			return
		}
	}
	r := rtn{Response: str, Success: success}

	if p, err := json.Marshal(r); err != nil {
		wsLog.Error(err)
		return
	} else {
		ctx.Write(p)
	}
}
