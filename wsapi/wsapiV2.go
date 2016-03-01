// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/receipts"
	"github.com/FactomProject/web"
	"io/ioutil"
)

var _ = fmt.Print

func HandleV2Get(ctx *web.Context) {
	HandleV2(ctx, false)
}

func HandleV2Post(ctx *web.Context) {
	HandleV2(ctx, true)
}

func HandleV2(ctx *web.Context, post bool) {
	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		HandleV2Error(ctx, nil, NewInvalidRequestError())
		return
	}

	j, err := primitives.ParseJSON2Request(string(body))
	if err != nil {
		HandleV2Error(ctx, nil, NewInvalidRequestError())
		return
	}

	state := ctx.Server.Env["state"].(interfaces.IState)

	var jsonResp *primitives.JSON2Response
	var jsonError *primitives.JSONError
	if post == true {
		jsonResp, jsonError = HandleV2PostRequest(state, j)
	} else {
		jsonResp, jsonError = HandleV2GetRequest(state, j)
	}

	if jsonError != nil {
		HandleV2Error(ctx, j, jsonError)
		return
	}

	ctx.Write([]byte(jsonResp.String()))

}

func HandleV2PostRequest(state interfaces.IState, j *primitives.JSON2Request) (*primitives.JSON2Response, *primitives.JSONError) {
	var resp interface{}
	var jsonError *primitives.JSONError
	params := j.Params
	switch j.Method {
	case "factoid-submit":
		resp, jsonError = HandleV2FactoidSubmit(state, params)
		break
	case "commit-chain":
		resp, jsonError = HandleV2CommitChain(state, params)
		break
	case "reveal-chain":
		resp, jsonError = HandleV2RevealChain(state, params)
		break
	case "commit-entry":
		resp, jsonError = HandleV2CommitEntry(state, params)
		break
	case "reveal-entry":
		resp, jsonError = HandleV2RevealEntry(state, params)
		break
	default:
		jsonError = NewMethodNotFoundError()
		break
	}
	if jsonError != nil {
		return nil, jsonError
	}

	jsonResp := primitives.NewJSON2Response()
	jsonResp.ID = j.ID
	jsonResp.Result = resp

	return jsonResp, nil
}

func HandleV2GetRequest(state interfaces.IState, j *primitives.JSON2Request) (*primitives.JSON2Response, *primitives.JSONError) {
	var resp interface{}
	var jsonError *primitives.JSONError
	params := j.Params
	switch j.Method {
	case "directory-block-head":
		resp, jsonError = HandleV2DirectoryBlockHead(state, params)
		break
	case "get-raw-data":
		resp, jsonError = HandleV2GetRaw(state, params)
		break
	case "get-receipt":
		resp, jsonError = HandleV2GetReceipt(state, params)
		break
	case "directory-block-by-keymr":
		resp, jsonError = HandleV2DirectoryBlock(state, params)
		break
	case "entry-block-by-keymr":
		resp, jsonError = HandleV2EntryBlock(state, params)
		break
	case "entry-by-hash":
		resp, jsonError = HandleV2Entry(state, params)
		break
	case "chain-head":
		resp, jsonError = HandleV2ChainHead(state, params)
		break
	case "entry-credit-balance":
		resp, jsonError = HandleV2EntryCreditBalance(state, params)
		break
	case "factoid-balance":
		resp, jsonError = HandleV2FactoidBalance(state, params)
		break
	case "factoid-get-fee":
		resp, jsonError = HandleV2GetFee(state, params)
		break
	case "directory-block-height":
		resp, jsonError = HandleV2DirectoryBlockHeight(state, params)
		break
	default:
		jsonError = NewMethodNotFoundError()
		break
	}
	if jsonError != nil {
		return nil, jsonError
	}

	jsonResp := primitives.NewJSON2Response()
	jsonResp.ID = j.ID
	jsonResp.Result = resp

	return jsonResp, nil
}

func HandleV2Error(ctx *web.Context, j *primitives.JSON2Request, err *primitives.JSONError) {
	resp := primitives.NewJSON2Response()
	if j != nil {
		resp.ID = j.ID
	} else {
		resp.ID = nil
	}
	resp.Error = err

	ctx.WriteHeader(httpBad)
	ctx.Write([]byte(resp.String()))
}

func HandleV2CommitChain(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	commitChainMsg, ok := params.(string)
	if ok == false {
		return nil, NewInvalidParamsError()
	}

	commit := entryCreditBlock.NewCommitChain()
	if p, err := hex.DecodeString(commitChainMsg); err != nil {
		return nil, NewInvalidCommitChainError()
	} else {
		_, err := commit.UnmarshalBinaryData(p)
		if err != nil {
			return nil, NewInvalidCommitChainError()
		}
	}

	msg := new(messages.CommitChainMsg)
	msg.CommitChain = commit
	msg.Timestamp = state.GetTimestamp()
	state.InMsgQueue() <- msg

	resp := new(CommitChainResponse)
	resp.Message = "Chain Commit Success"

	return resp, nil
}

func HandleV2RevealChain(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	return HandleV2RevealEntry(state, params)
}

func HandleV2CommitEntry(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	commitEntryMsg, ok := params.(string)
	if ok == false {
		return nil, NewInvalidParamsError()
	}

	commit := entryCreditBlock.NewCommitEntry()
	if p, err := hex.DecodeString(commitEntryMsg); err != nil {
		return nil, NewInvalidCommitEntryError()
	} else {
		_, err := commit.UnmarshalBinaryData(p)
		if err != nil {
			return nil, NewInvalidCommitEntryError()
		}
	}

	msg := new(messages.CommitEntryMsg)
	msg.CommitEntry = commit
	msg.Timestamp = state.GetTimestamp()
	state.InMsgQueue() <- msg

	resp := new(CommitEntryResponse)
	resp.Message = "Entry Commit Success"

	return resp, nil
}

func HandleV2RevealEntry(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	e, ok := params.(string)
	if ok == false {
		return nil, NewInvalidParamsError()
	}

	entry := entryBlock.NewEntry()
	if p, err := hex.DecodeString(e); err != nil {
		return nil, NewInvalidEntryError()
	} else {
		_, err := entry.UnmarshalBinaryData(p)
		if err != nil {
			return nil, NewInvalidEntryError()
		}
	}

	msg := new(messages.RevealEntryMsg)
	msg.Entry = entry
	msg.Timestamp = state.GetTimestamp()
	state.InMsgQueue() <- msg

	resp := new(RevealEntryResponse)
	resp.Message = "Entry Reveal Success"

	return resp, nil
}

func HandleV2DirectoryBlockHead(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	h := new(DirectoryBlockHeadResponse)
	h.KeyMR = state.GetDirectoryBlock().GetKeyMR().String()
	return h, nil
}

func HandleV2GetRaw(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	//TODO: var block interfaces.BinaryMarshallable
	hashkey, ok := params.(string)
	if ok == false {
		return nil, NewInvalidParamsError()
	}

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		return nil, NewInvalidHashError()
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
	} else if block, _ := dbase.FetchECBlockByHeaderHash(h); block != nil {
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
	} else {
		return nil, NewEntryNotFoundError()
	}

	d := new(GetRawDataResponse)
	d.Data = hex.EncodeToString(b)
	return d, nil
}

func HandleV2GetReceipt(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	hashkey, ok := params.(string)
	if ok == false {
		return nil, NewInvalidParamsError()
	}

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		return nil, NewInvalidHashError()
	}

	dbase := state.GetDB()

	receipt, err := receipts.CreateFullReceipt(dbase, h)
	if err != nil {
		return nil, NewMiscError()
	}
	resp := new(GetReceiptResponse)
	resp.Receipt = receipt
	return resp, nil
}

func HandleV2DirectoryBlock(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	hashkey, ok := params.(string)
	if ok == false {
		return nil, NewInvalidParamsError()
	}
	d := new(DirectoryBlockResponse)

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		return nil, NewInvalidHashError()
	}

	dbase := state.GetDB()

	block, err := dbase.FetchDBlockByKeyMR(h)
	if err != nil {
		return nil, NewInvalidHashError()
	}
	if block == nil {
		block, err = dbase.FetchDBlockByHash(h)
		if err != nil {
			return nil, NewInvalidHashError()
		}
		if block == nil {
			return nil, NewBlockNotFoundError()
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

	return d, nil
}

func HandleV2EntryBlock(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	hashkey, ok := params.(string)
	if ok == false {
		return nil, NewInvalidParamsError()
	}
	e := new(EntryBlockResponse)

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		return nil, NewInvalidHashError()
	}

	dbase := state.GetDB()

	block, err := dbase.FetchEBlockByKeyMR(h)
	if err != nil {
		return nil, NewInvalidHashError()
	}
	if block == nil {
		block, err = dbase.FetchEBlockByHash(h)
		if err != nil {
			return nil, NewInvalidHashError()
		}
		if block == nil {
			return nil, NewBlockNotFoundError()
		}
	}

	e.Header.BlockSequenceNumber = block.GetHeader().GetEBSequence()
	e.Header.ChainID = block.GetHeader().GetChainID().String()
	e.Header.PrevKeyMR = block.GetHeader().GetPrevKeyMR().String()
	e.Header.DBHeight = block.GetHeader().GetDBHeight()

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

	return e, nil
}

func HandleV2Entry(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	hashkey, ok := params.(string)
	if ok == false {
		return nil, NewInvalidParamsError()
	}
	e := new(EntryResponse)

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		return nil, NewInvalidHashError()
	}

	dbase := state.GetDB()

	entry, err := dbase.FetchEntryByHash(h)
	if err != nil {
		return nil, NewInvalidHashError()
	}
	if entry == nil {
		return nil, NewEntryNotFoundError()
	}

	e.ChainID = entry.GetChainIDHash().String()
	e.Content = hex.EncodeToString(entry.GetContent())
	for _, v := range entry.ExternalIDs() {
		e.ExtIDs = append(e.ExtIDs, hex.EncodeToString(v))
	}

	return e, nil
}

func HandleV2ChainHead(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	hashkey, ok := params.(string)
	if ok == false {
		return nil, NewInvalidParamsError()
	}

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		return nil, NewInvalidHashError()
	}

	dbase := state.GetDB()

	mr, err := dbase.FetchHeadIndexByChainID(h)
	if err != nil {
		return nil, NewInvalidHashError()
	}
	if mr == nil {
		return nil, NewMissingChainHeadError()
	}
	c := new(ChainHeadResponse)
	c.ChainHead = mr.String()
	return c, nil
}

func HandleV2EntryCreditBalance(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	eckey, ok := params.(string)
	if ok == false {
		return nil, NewInvalidParamsError()
	}

	var adr []byte
	var err error

	if primitives.ValidateECUserStr(eckey) {
		adr = primitives.ConvertUserStrToAddress(eckey)
	} else {
		adr, err = hex.DecodeString(eckey)
		if err == nil && len(adr) != constants.HASH_LENGTH {
			return nil, NewInvalidAddressError()
		}
		if err != nil {
			return nil, NewInvalidAddressError()
		}
	}

	if len(adr) != constants.HASH_LENGTH {
		return nil, NewInvalidAddressError()
	}

	address, err := primitives.NewShaHash(adr)
	if err != nil {
		return nil, NewInvalidAddressError()
	}
	resp := new(EntryCreditBalanceResponse)
	resp.Balance = state.GetFactoidState().GetECBalance(address.Fixed())
	return resp, nil
}

func HandleV2GetFee(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	resp := new(FactoidGetFeeResponse)
	resp.Fee = state.GetFactoshisPerEC()

	fmt.Println(";;;;;;;;;;;;;;;;;;;;;;;;;;;;;; ", resp.Fee)

	return resp, nil
}

func HandleV2FactoidSubmit(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	t, ok := params.(string)
	if ok == false {
		return nil, NewInvalidParamsError()
	}

	msg := new(messages.FactoidTransaction)
	p, err := hex.DecodeString(t)
	if err != nil {
		return nil, NewUnableToDecodeTransactionError()
	}

	_, err = msg.UnmarshalTransData(p)
	if err != nil {
		return nil, NewUnableToDecodeTransactionError()
	}

	err = state.GetFactoidState().Validate(1, msg.Transaction)
	if err != nil {
		return nil, NewInvalidTransactionError()
	}

	state.InMsgQueue() <- msg

	resp := new(FactoidSubmitResponse)
	resp.Message = "Successfully submitted the transaction"

	return resp, nil
}

func HandleV2FactoidBalance(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	fkey, ok := params.(string)
	if ok == false {
		return nil, NewInvalidParamsError()
	}

	var adr []byte
	var err error

	if primitives.ValidateFUserStr(fkey) {
		adr = primitives.ConvertUserStrToAddress(fkey)
	} else {
		adr, err = hex.DecodeString(fkey)
		if err == nil && len(adr) != constants.HASH_LENGTH {
			return nil, NewInvalidAddressError()
		}
		if err != nil {
			return nil, NewInvalidAddressError()
		}
	}

	if len(adr) != constants.HASH_LENGTH {
		return nil, NewInvalidAddressError()
	}

	resp := new(FactoidBalanceResponse)
	resp.Balance = state.GetFactoidState().GetFactoidBalance(factoid.NewAddress(adr).Fixed())
	return resp, nil
}

func HandleV2DirectoryBlockHeight(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	h := new(DirectoryBlockHeightResponse)
	h.Height = int64(state.GetDirectoryBlock().GetHeader().GetDBHeight())
	return h, nil
}
