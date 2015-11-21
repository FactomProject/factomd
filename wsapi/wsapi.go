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
	"github.com/FactomProject/factomd/common/messages"
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

		server.Post("/v1/factoid-submit/?", handleFactoidSubmit)
		server.Post("/v1/commit-chain/?", handleCommitChain)
		server.Post("/v1/reveal-chain/?", handleRevealChain)
		server.Post("/v1/commit-entry/?", handleCommitEntry)
		server.Post("/v1/reveal-entry/?", handleRevealEntry)
		server.Get("/v1/directory-block-head/?", handleDirectoryBlockHead)
		server.Get("/v1/get-raw-data/([^/]+)", handleGetRaw)
		server.Get("/v1/directory-block-by-keymr/([^/]+)", handleDirectoryBlock)
		server.Get("/v1/entry-block-by-keymr/([^/]+)", handleEntryBlock)
		server.Get("/v1/entry-by-hash/([^/]+)", handleEntry)
		server.Get("/v1/chain-head/([^/]+)", handleChainHead)
		server.Get("/v1/entry-credit-balance/([^/]+)", handleEntryCreditBalance)
		server.Get("/v1/factoid-balance/([^/]+)", handleFactoidBalance)
		server.Get("/v1/factoid-get-fee/", handleGetFee)

		log.Print("Starting server")
		go server.Run(fmt.Sprintf("localhost:%d", state.GetPort()))
	}

}

func Stop(state interfaces.IState) {
	Servers[state.GetPort()].Close()
}

func handleCommitChain(ctx *web.Context) {

}

func handleRevealChain(ctx *web.Context) {

}

func handleCommitEntry(ctx *web.Context) {

}

func handleRevealEntry(ctx *web.Context) {

}

func handleDirectoryBlockHead(ctx *web.Context) {

	state := ctx.Server.Env["state"].(interfaces.IState)

	type dbhead struct {
		KeyMR string
	}

	h := new(dbhead)

	h.KeyMR = state.GetPreviousDirectoryBlock().GetKeyMR().String()

	fmt.Println(h.KeyMR)

	if p, err := json.Marshal(h); err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	} else {
		ctx.Write(p)
	}

}

func handleGetRaw(ctx *web.Context, hashkey string) {

	state := ctx.Server.Env["state"].(interfaces.IState)

	type rawData struct {
		Data string
	}
	//TODO: var block interfaces.BinaryMarshallable
	d := new(rawData)

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	}

	dbase := state.GetDB()

	var b []byte
	// try to find the block data in db and return the first one found
	if block, _ := dbase.FetchFBlockByHash(h); block != nil {
		b, _ = block.MarshalBinary()
	} else if block, _ := dbase.FetchDBlockByHash(h); block != nil {
		b, _ = block.MarshalBinary()
	} else if block, _ := dbase.FetchABlockByHash(h); block != nil {
		b, _ = block.MarshalBinary()
	} else if block, _ := dbase.FetchEBlockByHash(h); block != nil {
		b, _ = block.MarshalBinary()
	} else if block, _ := dbase.FetchECBlockByHash(h); block != nil {
		b, _ = block.MarshalBinary()
	} else if block, _ := dbase.FetchEntryByHash(h); block != nil {
		b, _ = block.MarshalBinary()
	}
	//TODO: fetch by KeyMR first, then by Hashes

	d.Data = hex.EncodeToString(b)

	if p, err := json.Marshal(d); err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	} else {
		ctx.Write(p)
	}
}

func handleDirectoryBlock(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)
	/*
		type data struct {
		}
		d := new(data)
	*/

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	}

	dbase := state.GetDB()

	block, err := dbase.FetchDBlockByKeyMR(h)
	if err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	}
	if block == nil {
		block, err = dbase.FetchDBlockByHash(h)
		if err != nil {
			wsLog.Error(err)
			ctx.WriteHeader(httpBad)
			ctx.Write([]byte(err.Error()))
			return
		}
		if block == nil {
			//TODO: handle block not found
		}
	}
	//TODO: handle block found
}

func handleEntryBlock(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)
	/*
		type data struct {
		}
		d := new(data)
	*/

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	}

	dbase := state.GetDB()

	block, err := dbase.FetchEBlockByKeyMR(h)
	if err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	}
	if block == nil {
		block, err = dbase.FetchEBlockByHash(h)
		if err != nil {
			wsLog.Error(err)
			ctx.WriteHeader(httpBad)
			ctx.Write([]byte(err.Error()))
			return
		}
		if block == nil {
			//TODO: handle block not found
		}
	}
	//TODO: handle block found

}

func handleEntry(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)
	/*
		type data struct {
		}
		d := new(data)
	*/

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	}

	dbase := state.GetDB()

	block, err := dbase.FetchEntryByHash(h)
	if err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	}
	if block == nil {
		//TODO: handle block not found
	}
	//TODO: handle block found

}

func handleChainHead(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)
	/*
		type data struct {
		}
		d := new(data)
	*/

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	}

	dbase := state.GetDB()

	block, err := dbase.FetchDBlockByKeyMR(h)
	if err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	}
	if block == nil {
		block, err = dbase.FetchDBlockByHash(h)
		if err != nil {
			wsLog.Error(err)
			ctx.WriteHeader(httpBad)
			ctx.Write([]byte(err.Error()))
			return
		}
		if block == nil {
			//TODO: handle block not found
		}
	}
	//TODO: handle block found

}

func handleEntryCreditBalance(ctx *web.Context) {

}

func handleGetFee(ctx *web.Context) {
	
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

func handleFactoidSubmit(ctx *web.Context) {

	state := ctx.Server.Env["state"].(interfaces.IState)

	type x struct{ Transaction string }
	t := new(x)

	var p []byte
	var err error
	if p, err = ioutil.ReadAll(ctx.Request.Body); err != nil {
		wsLog.Error(err)
		returnMsg(ctx, "Unable to read the request", false)
		return
	} else {
		if err := json.Unmarshal(p, t); err != nil {
			returnMsg(ctx, "Unable to Unmarshal the request", false)
			return
		}
	}

	msg := new(messages.FactoidTransaction)

	if p, err = hex.DecodeString(t.Transaction); err != nil {
		returnMsg(ctx, "Unable to decode the transaction", false)
		return
	}

	fmt.Println(t,"\n",p)
	
	err = msg.UnmarshalBinary(p)

	if err != nil {
		returnMsg(ctx, err.Error(), false)
		return
	}

	err = state.GetFactoidState().Validate(1, msg.Transaction)

	if err != nil {
		returnMsg(ctx, err.Error(), false)
		return
	}

	state.NetworkInMsgQueue() <- msg

	returnMsg(ctx, "Successfully submitted the transaction", true)

}

func handleFactoidBalance(ctx *web.Context, eckey string) {

	state := ctx.Server.Env["state"].(interfaces.IState)

	type fbal struct {
		Response string
		Success  bool
	}
	var b fbal
	adr, err := hex.DecodeString(eckey)
	if err == nil && len(adr) != constants.HASH_LENGTH {
		b = fbal{Response: "Invalid Address", Success: false}
	}
	if err == nil {
		v := int64(state.GetFactoidState().GetBalance(factoid.NewAddress(adr).Fixed()))
		str := fmt.Sprintf("%d", v)
		b = fbal{Response: str, Success: true}
	} else {
		b = fbal{Response: err.Error(), Success: false}
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

func returnMsg(ctx *web.Context, msg string, success bool) {
	type rtn struct {
		Response string
		Success  bool
	}
	r := rtn{Response: msg, Success: success}

	if p, err := json.Marshal(r); err != nil {
		wsLog.Error(err)
		return
	} else {
		ctx.Write(p)
	}
}
