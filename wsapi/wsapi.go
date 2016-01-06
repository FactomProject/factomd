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
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
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

		server.Post("/v1/factoid-submit/?", HandleFactoidSubmit)
		server.Post("/v1/commit-chain/?", HandleCommitChain)
		server.Post("/v1/reveal-chain/?", HandleRevealChain)
		server.Post("/v1/commit-entry/?", HandleCommitEntry)
		server.Post("/v1/reveal-entry/?", HandleRevealEntry)
		server.Get("/v1/directory-block-head/?", HandleDirectoryBlockHead)
		server.Get("/v1/get-raw-data/([^/]+)", HandleGetRaw)
		server.Get("/v1/directory-block-by-keymr/([^/]+)", HandleDirectoryBlock)
		server.Get("/v1/entry-block-by-keymr/([^/]+)", HandleEntryBlock)
		server.Get("/v1/entry-by-hash/([^/]+)", HandleEntry)
		server.Get("/v1/chain-head/([^/]+)", HandleChainHead)
		server.Get("/v1/entry-credit-balance/([^/]+)", HandleEntryCreditBalance)
		server.Get("/v1/factoid-balance/([^/]+)", HandleFactoidBalance)
		server.Get("/v1/factoid-get-fee/", HandleGetFee)

		log.Print("Starting server")
		go server.Run(fmt.Sprintf("localhost:%d", state.GetPort()))
	}

}

func Stop(state interfaces.IState) {
	Servers[state.GetPort()].Close()
}

func HandleCommitChain(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	type commitchain struct {
		CommitChainMsg string
	}

	c := new(commitchain)
	if p, err := ioutil.ReadAll(ctx.Request.Body); err != nil {
		returnMsg(ctx, "Bad commit message", false)
		return
	} else {
		if err := json.Unmarshal(p, c); err != nil {
			returnMsg(ctx, "Bad commit message", false)
			return
		}
	}

	commit := entryCreditBlock.NewCommitChain()
	if p, err := hex.DecodeString(c.CommitChainMsg); err != nil {
		returnMsg(ctx, "Bad commit message", false)
		return
	} else {
		_, err := commit.UnmarshalBinaryData(p)
		if err != nil {
			returnMsg(ctx, "Bad commit message", false)
			return
		}
	}

	msg := new(messages.CommitChainMsg)
	msg.CommitChain = commit
	msg.Timestamp = state.GetTimestamp()
	state.InMsgQueue() <- msg

	returnMsg(ctx, "Chain Commit Success", true)
}

func HandleRevealChain(ctx *web.Context) {
	HandleRevealEntry(ctx)
}

func HandleCommitEntry(ctx *web.Context) {

}

func HandleRevealEntry(ctx *web.Context) {
	type revealentry struct {
		Entry string
	}

	e := new(revealentry)
	if p, err := ioutil.ReadAll(ctx.Request.Body); err != nil {
		returnMsg(ctx, "Error Reveal Entry: "+err.Error(), false)
		return
	} else {
		if err := json.Unmarshal(p, e); err != nil {
			returnMsg(ctx, "Error Reveal Entry: "+err.Error(), false)
			return
		}
	}

	entry := entryBlock.NewEntry()
	if p, err := hex.DecodeString(e.Entry); err != nil {
		returnMsg(ctx, "Error Reveal Entry: "+err.Error(), false)
		return
	} else {
		_, err := entry.UnmarshalBinaryData(p)
		if err != nil {
			returnMsg(ctx, "Error Reveal Entry: "+err.Error(), false)
			return
		}
	}

	state := ctx.Server.Env["state"].(interfaces.IState)

	msg := new(messages.RevealEntryMsg)
	msg.Entry = entry
	msg.Timestamp = state.GetTimestamp()
	state.InMsgQueue() <- msg

	returnMsg(ctx, "Entry Reveal Success", true)
}

func HandleDirectoryBlockHead(ctx *web.Context) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	h := new(DBHead)

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

func HandleGetRaw(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	//TODO: var block interfaces.BinaryMarshallable
	d := new(RawData)

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

func HandleDirectoryBlock(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	d := new(DBlock)

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
			//TODO: Handle block not found
			return
		}
	}

	d.Header.PrevBlockKeyMR = block.GetHeader().GetPrevKeyMR().String()
	d.Header.SequenceNumber = block.GetHeader().GetDBHeight()
	d.Header.Timestamp = block.GetHeader().GetTimestamp() * 60
	for _, v := range block.GetDBEntries() {
		l := new(EBlockAddr)
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

func HandleEntryBlock(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	e := new(EBlock)

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
			//TODO: Handle block not found
			return
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

	estack := make([]EntryAddr, 0)
	for _, v := range block.GetBody().GetEBEntries() {
		if n, exist := mins[v.String()]; exist {
			// the entry is a minute marker. add time to all of the
			// previous entries for the minute
			t := e.Header.Timestamp + 60*uint32(n)
			for _, w := range estack {
				w.Timestamp = t
				e.EntryList = append(e.EntryList, w)
			}
			estack = make([]EntryAddr, 0)
		} else {
			l := new(EntryAddr)
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

func HandleEntry(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	e := new(EntryStruct)

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
		//TODO: Handle block not found
		return
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

func HandleChainHead(ctx *web.Context, hashkey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	c := new(CHead)

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
		err := fmt.Errorf("Missing Chain Head")
		wsLog.Error(err)
		ctx.WriteHeader(httpBad)
		ctx.Write([]byte(err.Error()))
		return
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

func HandleEntryCreditBalance(ctx *web.Context, eckey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	var b FactoidBalance
	adr, err := primitives.HexToHash(eckey)
	if err == nil {
		b = FactoidBalance{Response: "Invalid Address", Success: false}
	}
	if err == nil {
		v := int64(state.GetFactoidState().GetECBalance(adr.Fixed()))
		str := fmt.Sprintf("%d", v)
		b = FactoidBalance{Response: str, Success: true}
	} else {
		b = FactoidBalance{Response: err.Error(), Success: false}
	}

	if p, err := json.Marshal(b); err != nil {
		wsLog.Error(err)
		return
	} else {
		ctx.Write(p)
	}

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

func HandleFactoidBalance(ctx *web.Context, eckey string) {
	state := ctx.Server.Env["state"].(interfaces.IState)

	var b FactoidBalance
	adr, err := hex.DecodeString(eckey)
	if err == nil && len(adr) != constants.HASH_LENGTH {
		b = FactoidBalance{Response: "Invalid Address", Success: false}
	}
	if err == nil {
		v := int64(state.GetFactoidState().GetFactoidBalance(factoid.NewAddress(adr).Fixed()))
		str := fmt.Sprintf("%d", v)
		b = FactoidBalance{Response: str, Success: true}
	} else {
		b = FactoidBalance{Response: err.Error(), Success: false}
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
