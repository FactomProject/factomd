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
	if block, _ := dbase.FetchFBlockByKeyMR(h); block != nil {
		b, _ = block.MarshalBinary()
	} else if block, _ := dbase.FetchDBlockByKeyMR(h); block != nil {
		b, _ = block.MarshalBinary()
	} else if block, _ := dbase.FetchABlockByKeyMR(h); block != nil {
		b, _ = block.MarshalBinary()
	} else if block, _ := dbase.FetchEBlockByKeyMR(h); block != nil {
		b, _ = block.MarshalBinary()
	} else if block, _ := dbase.FetchECBlockByKeyMR(h); block != nil {
		b, _ = block.MarshalBinary()
	} else if block, _ := dbase.FetchEntryByHash(h); block != nil {
		b, _ = block.MarshalBinary()
	} else if block, _ := dbase.FetchFBlockByHash(h); block != nil {
		b, _ = block.MarshalBinary()
	} else if block, _ := dbase.FetchDBlockByHash(h); block != nil {
		b, _ = block.MarshalBinary()
	} else if block, _ := dbase.FetchABlockByHash(h); block != nil {
		b, _ = block.MarshalBinary()
	} else if block, _ := dbase.FetchEBlockByHash(h); block != nil {
		b, _ = block.MarshalBinary()
	} else if block, _ := dbase.FetchECBlockByHash(h); block != nil {
		b, _ = block.MarshalBinary()
	}

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

	type eblockaddr struct {
		ChainID string
		KeyMR   string
	}

	type dblock struct {
		Header struct {
			PrevBlockKeyMR string
			SequenceNumber uint32
			Timestamp      uint32
		}
		EntryBlockList []eblockaddr
	}
	d := new(dblock)

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
			return
		}
	}

	d.Header.PrevBlockKeyMR = block.GetHeader().GetPrevKeyMR().String()
	d.Header.SequenceNumber = block.GetHeader().GetDBHeight()
	d.Header.Timestamp = block.GetHeader().GetTimestamp() * 60
	for _, v := range block.GetDBEntries() {
		l := new(eblockaddr)
		l.ChainID = v.GetChainID().String()
		l.KeyMR = v.GetKeyMR().String()
		d.EntryBlockList = append(d.EntryBlockList, *l)
	}

	if p, err := json.Marshal(d); err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	} else {
		ctx.Write(p)
	}
}

func handleEntryBlock(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	type entryaddr struct {
		EntryHash string
		Timestamp uint32
	}

	type eblock struct {
		Header struct {
			BlockSequenceNumber uint32
			ChainID             string
			PrevKeyMR           string
			Timestamp           uint32
		}
		EntryList []entryaddr
	}
	e := new(eblock)

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

	e.Header.BlockSequenceNumber = block.GetHeader().GetEBSequence()
	e.Header.ChainID = block.GetHeader().GetChainID().String()
	e.Header.PrevKeyMR = block.GetHeader().GetPrevKeyMR().String()

	if dblock, err := dbase.FetchDBlockByHeight(block.GetHeader().GetDBHeight()); err == nil {
		e.Header.Timestamp = dblock.GetHeader().GetTimestamp() * 60
	}

	// create a map of possible minute markers that may be found in the
	// EBlock Body
	mins := make(map[string]uint8)
	for i := byte(1); i <= 10; i++ {
		h := make([]byte, 32)
		h[len(h)-1] = i
		mins[hex.EncodeToString(h)] = i
	}

	estack := make([]entryaddr, 0)
	for _, v := range block.GetBody().GetEBEntries() {
		if n, exist := mins[v.String()]; exist {
			// the entry is a minute marker. add time to all of the
			// previous entries for the minute
			t := e.Header.Timestamp + 60*uint32(n)
			for _, w := range estack {
				w.Timestamp = t
				e.EntryList = append(e.EntryList, w)
			}
			estack = make([]entryaddr, 0)
		} else {
			l := new(entryaddr)
			l.EntryHash = v.String()
			estack = append(estack, *l)
		}
	}

	if p, err := json.Marshal(e); err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	} else {
		ctx.Write(p)
	}

}

func handleEntry(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	type entryStruct struct {
		ChainID string
		Content string
		ExtIDs  []string
	}

	e := new(entryStruct)

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	}

	dbase := state.GetDB()

	entry, err := dbase.FetchEntryByHash(h)
	if err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	}
	if entry == nil {
		//TODO: handle block not found
	}

	e.ChainID = entry.GetChainIDHash().String()
	e.Content = hex.EncodeToString(entry.GetContent())
	for _, v := range entry.ExternalIDs() {
		e.ExtIDs = append(e.ExtIDs, hex.EncodeToString(v))
	}

	if p, err := json.Marshal(e); err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	} else {
		ctx.Write(p)
	}

}

func handleChainHead(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	type chead struct {
		ChainHead string
	}

	c := new(chead)

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	}

	dbase := state.GetDB()

	mr, err := dbase.FetchHeadIndexByChainID(h)
	if err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	}
	if mr == nil {
		//TODO: handle block not found
	}
	c.ChainHead = mr.String()
	if p, err := json.Marshal(c); err != nil {
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
	} else {
		ctx.Write(p)
	}

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

	_, err = msg.UnmarshalTransData(p)
	
	if err != nil {
		returnMsg(ctx, err.Error(), false)
		return
	}

	err = state.GetFactoidState().Validate(1, msg.Transaction)

	if err != nil {
		returnMsg(ctx, err.Error(), false)
		return
	}

	state.InMsgQueue() <- msg

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
