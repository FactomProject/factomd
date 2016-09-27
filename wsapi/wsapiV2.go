// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/receipts"
	"github.com/FactomProject/web"
)

const API_VERSION string = "2.0"

func HandleV2(ctx *web.Context) {
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

	ServersMutex.Lock()
	state := ctx.Server.Env["state"].(interfaces.IState)
	ServersMutex.Unlock()

	jsonResp, jsonError := HandleV2Request(state, j)

	if jsonError != nil {
		HandleV2Error(ctx, j, jsonError)
		return
	}

	ctx.Write([]byte(jsonResp.String()))
}

func HandleV2Request(state interfaces.IState, j *primitives.JSON2Request) (*primitives.JSON2Response, *primitives.JSONError) {
	var resp interface{}
	var jsonError *primitives.JSONError
	params := j.Params
	switch j.Method {
	case "chain-head":
		resp, jsonError = HandleV2ChainHead(state, params)
		break
	case "commit-chain":
		resp, jsonError = HandleV2CommitChain(state, params)
		break
	case "commit-entry":
		resp, jsonError = HandleV2CommitEntry(state, params)
		break
	case "directory-block":
		resp, jsonError = HandleV2DirectoryBlock(state, params)
		break
	case "directory-block-head":
		resp, jsonError = HandleV2DirectoryBlockHead(state, params)
		break
	case "directory-block-height":
		resp, jsonError = HandleV2DirectoryBlockHeight(state, params)
		break
	case "entry-block":
		resp, jsonError = HandleV2EntryBlock(state, params)
		break
	case "entry":
		resp, jsonError = HandleV2Entry(state, params)
		break
	case "entry-credit-balance":
		resp, jsonError = HandleV2EntryCreditBalance(state, params)
		break
	case "entry-credit-rate":
		resp, jsonError = HandleV2EntryCreditRate(state, params)
		break
	case "factoid-balance":
		resp, jsonError = HandleV2FactoidBalance(state, params)
		break
	case "factoid-submit":
		resp, jsonError = HandleV2FactoidSubmit(state, params)
		break
	case "raw-data":
		resp, jsonError = HandleV2RawData(state, params)
		break
	case "receipt":
		resp, jsonError = HandleV2Receipt(state, params)
		break
	case "properties":
		resp, jsonError = HandleV2Properties(state, params)
		break
	case "reveal-chain":
		resp, jsonError = HandleV2RevealChain(state, params)
		break
	case "reveal-entry":
		resp, jsonError = HandleV2RevealEntry(state, params)
		break
	case "factoid-ack":
		resp, jsonError = HandleV2FactoidACK(state, params)
		break
	case "entry-ack":
		resp, jsonError = HandleV2EntryACK(state, params)
		break
	case "send-raw-message":
		resp, jsonError = HandleV2SendRawMessage(state, params)
		break
	case "get-transaction":
		resp, jsonError = HandleV2GetTranasction(state, params)
		break
	case "get-height":
		resp, jsonError = HandleV2GetHeight(state, params)
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

func MapToObject(source interface{}, dst interface{}) error {
	b, err := json.Marshal(source)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

func HandleV2CommitChain(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	commitChainMsg := new(MessageRequest)
	err := MapToObject(params, commitChainMsg)
	if err != nil {
		return nil, NewInvalidParamsError()
	}

	commit := entryCreditBlock.NewCommitChain()
	if p, err := hex.DecodeString(commitChainMsg.Message); err != nil {
		return nil, NewInvalidCommitChainError()
	} else {
		_, err := commit.UnmarshalBinaryData(p)
		if err != nil {
			return nil, NewInvalidCommitChainError()
		}
	}

	msg := new(messages.CommitChainMsg)
	msg.CommitChain = commit
	state.APIQueue() <- msg
	state.IncECCommits()

	resp := new(CommitChainResponse)
	resp.Message = "Chain Commit Success"
	resp.TxID = commit.GetSigHash().String()

	return resp, nil
}

func HandleV2RevealChain(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	return HandleV2RevealEntry(state, params)
}

func HandleV2CommitEntry(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	commitEntryMsg := new(EntryRequest)
	err := MapToObject(params, commitEntryMsg)
	if err != nil {
		return nil, NewInvalidParamsError()
	}

	commit := entryCreditBlock.NewCommitEntry()
	if p, err := hex.DecodeString(commitEntryMsg.Entry); err != nil {
		return nil, NewInvalidCommitEntryError()
	} else {
		_, err := commit.UnmarshalBinaryData(p)
		if err != nil {
			return nil, NewInvalidCommitEntryError()
		}
	}

	msg := new(messages.CommitEntryMsg)
	msg.CommitEntry = commit
	state.APIQueue() <- msg
	state.IncECommits()

	resp := new(CommitEntryResponse)
	resp.Message = "Entry Commit Success"
	resp.TxID = commit.GetSigHash().String()

	return resp, nil
}

func HandleV2RevealEntry(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	e := new(EntryRequest)
	err := MapToObject(params, e)
	if err != nil {
		return nil, NewInvalidParamsError()
	}

	entry := entryBlock.NewEntry()
	if p, err := hex.DecodeString(e.Entry); err != nil {
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
	state.APIQueue() <- msg

	resp := new(RevealEntryResponse)
	resp.Message = "Entry Reveal Success"
	resp.EntryHash = entry.GetHash().String()

	return resp, nil
}

func HandleV2DirectoryBlockHead(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	h := new(DirectoryBlockHeadResponse)
	d := state.GetDirectoryBlockByHeight(state.GetHighestCompletedBlock())
	h.KeyMR = d.GetKeyMR().String()
	return h, nil
}

func HandleV2RawData(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	hashkey := new(HashRequest)
	err := MapToObject(params, hashkey)
	if err != nil {
		panic(reflect.TypeOf(params))
		return nil, NewInvalidParamsError()
	}

	h, err := primitives.HexToHash(hashkey.Hash)
	if err != nil {
		return nil, NewInvalidHashError()
	}

	var block interfaces.BinaryMarshallable
	var b []byte

	if block, _ = state.FetchECTransactionByHash(h); block != nil {
		b, _ = block.MarshalBinary()
	} else if block, _ = state.FetchFactoidTransactionByHash(h); block != nil {
		b, _ = block.MarshalBinary()
	} else if block, _ = state.FetchEntryByHash(h); block != nil {
		b, _ = block.MarshalBinary()
	}

	if b == nil {
		dbase := state.GetAndLockDB()
		defer state.UnlockDB()

		// try to find the block data in db and return the first one found
		if block, _ = dbase.FetchFBlock(h); block != nil {
			b, _ = block.MarshalBinary()
		} else if block, _ = dbase.FetchDBlock(h); block != nil {
			b, _ = block.MarshalBinary()
		} else if block, _ = dbase.FetchABlock(h); block != nil {
			b, _ = block.MarshalBinary()
		} else if block, _ = dbase.FetchEBlock(h); block != nil {
			b, _ = block.MarshalBinary()
		} else if block, _ = dbase.FetchECBlock(h); block != nil {
			b, _ = block.MarshalBinary()
		} else if block, _ = dbase.FetchFBlock(h); block != nil {
			b, _ = block.MarshalBinary()
		} else if block, _ = dbase.FetchEntry(h); block != nil {
			b, _ = block.MarshalBinary()
		} else {
			return nil, NewEntryNotFoundError()
		}
	}

	d := new(RawDataResponse)
	d.Data = hex.EncodeToString(b)
	return d, nil
}

func HandleV2Receipt(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	hashkey := new(HashRequest)
	err := MapToObject(params, hashkey)
	if err != nil {
		return nil, NewInvalidParamsError()
	}

	h, err := primitives.HexToHash(hashkey.Hash)
	if err != nil {
		return nil, NewInvalidHashError()
	}

	dbase := state.GetAndLockDB()
	defer state.UnlockDB()

	receipt, err := receipts.CreateFullReceipt(dbase, h)
	if err != nil {
		return nil, NewReceiptError()
	}
	resp := new(ReceiptResponse)
	resp.Receipt = receipt

	return resp, nil
}

func HandleV2DirectoryBlock(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	keymr := new(KeyMRRequest)
	err := MapToObject(params, keymr)
	if err != nil {
		return nil, NewInvalidParamsError()
	}

	h, err := primitives.HexToHash(keymr.KeyMR)
	if err != nil {
		return nil, NewInvalidHashError()
	}

	dbase := state.GetAndLockDB()
	defer state.UnlockDB()

	block, err := dbase.FetchDBlock(h)
	if err != nil {
		return nil, NewInvalidHashError()
	}
	if block == nil {
		block, err = dbase.FetchDBlock(h)
		if err != nil {
			return nil, NewInvalidHashError()
		}
		if block == nil {
			return nil, NewBlockNotFoundError()
		}
	}

	d := new(DirectoryBlockResponse)
	d.Header.PrevBlockKeyMR = block.GetHeader().GetPrevKeyMR().String()
	d.Header.SequenceNumber = int64(block.GetHeader().GetDBHeight())
	d.Header.Timestamp = block.GetHeader().GetTimestamp().GetTimeSeconds()
	for _, v := range block.GetDBEntries() {
		l := new(EBlockAddr)
		l.ChainID = v.GetChainID().String()
		l.KeyMR = v.GetKeyMR().String()
		d.EntryBlockList = append(d.EntryBlockList, *l)
	}

	return d, nil
}

func HandleV2EntryBlock(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	keymr := new(KeyMRRequest)
	err := MapToObject(params, keymr)
	if err != nil {
		return nil, NewInvalidParamsError()
	}
	e := new(EntryBlockResponse)

	h, err := primitives.HexToHash(keymr.KeyMR)
	if err != nil {
		return nil, NewInvalidHashError()
	}

	dbase := state.GetAndLockDB()
	defer state.UnlockDB()

	block, err := dbase.FetchEBlock(h)
	if err != nil {
		return nil, NewInvalidHashError()
	}
	if block == nil {
		block, err = dbase.FetchEBlock(h)
		if err != nil {
			return nil, NewInvalidHashError()
		}
		if block == nil {
			return nil, NewBlockNotFoundError()
		}
	}

	e.Header.BlockSequenceNumber = int64(block.GetHeader().GetEBSequence())
	e.Header.ChainID = block.GetHeader().GetChainID().String()
	e.Header.PrevKeyMR = block.GetHeader().GetPrevKeyMR().String()
	e.Header.DBHeight = int64(block.GetHeader().GetDBHeight())

	if dblock, err := dbase.FetchDBlockByHeight(block.GetHeader().GetDBHeight()); err == nil {
		e.Header.Timestamp = dblock.GetHeader().GetTimestamp().GetTimeSeconds()
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
			t := int64(e.Header.Timestamp + 60*int64(n))
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
	hashkey := new(HashRequest)
	err := MapToObject(params, hashkey)
	if err != nil {
		return nil, NewInvalidParamsError()
	}
	e := new(EntryResponse)

	h, err := primitives.HexToHash(hashkey.Hash)
	if err != nil {
		return nil, NewInvalidHashError()
	}

	entry, err := state.FetchEntryByHash(h)
	if err != nil {
		return nil, NewInternalError()
	}
	if entry == nil {
		dbase := state.GetAndLockDB()
		defer state.UnlockDB()

		entry, err = dbase.FetchEntry(h)
		if err != nil {
			return nil, NewInvalidHashError()
		}
		if entry == nil {
			return nil, NewEntryNotFoundError()
		}
	}

	e.ChainID = entry.GetChainIDHash().String()
	e.Content = hex.EncodeToString(entry.GetContent())
	for _, v := range entry.ExternalIDs() {
		e.ExtIDs = append(e.ExtIDs, hex.EncodeToString(v))
	}

	return e, nil
}

func HandleV2ChainHead(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	chainid := new(ChainIDRequest)
	err := MapToObject(params, chainid)
	if err != nil {
		return nil, NewInvalidParamsError()
	}
	h, err := primitives.HexToHash(chainid.ChainID)
	if err != nil {
		return nil, NewInvalidHashError()
	}

	dbase := state.GetAndLockDB()
	defer state.UnlockDB()

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
	ecadr := new(AddressRequest)
	err := MapToObject(params, ecadr)
	if err != nil {
		return nil, NewInvalidParamsError()
	}

	var adr []byte

	if primitives.ValidateECUserStr(ecadr.Address) {
		adr = primitives.ConvertUserStrToAddress(ecadr.Address)
	} else {
		adr, err = hex.DecodeString(ecadr.Address)
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

func HandleV2EntryCreditRate(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	resp := new(EntryCreditRateResponse)
	resp.Rate = int64(state.GetPredictiveFER())

	return resp, nil
}

func HandleV2FactoidSubmit(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	t := new(TransactionRequest)
	err := MapToObject(params, t)
	if err != nil {
		return nil, NewInvalidParamsError()
	}

	msg := new(messages.FactoidTransaction)

	p, err := hex.DecodeString(t.Transaction)
	if err != nil {
		return nil, NewUnableToDecodeTransactionError()
	}

	_, err = msg.UnmarshalTransData(p)
	if err != nil {
		return nil, NewUnableToDecodeTransactionError()
	}

	state.IncFCTSubmits()

	state.APIQueue() <- msg

	resp := new(FactoidSubmitResponse)
	resp.Message = "Successfully submitted the transaction"
	resp.TxID = msg.Transaction.GetSigHash().String()

	return resp, nil
}

func HandleV2FactoidBalance(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	fadr := new(AddressRequest)
	err := MapToObject(params, fadr)
	if err != nil {
		return nil, NewInvalidParamsError()
	}

	var adr []byte

	if primitives.ValidateFUserStr(fadr.Address) {
		adr = primitives.ConvertUserStrToAddress(fadr.Address)
	} else {
		adr, err = hex.DecodeString(fadr.Address)
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
	h.Height = int64(state.GetHighestCompletedBlock())
	return h, nil
}

func HandleV2GetHeight(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	h := new(HeightResponse)

	h.DirectoryBlockHeight = int64(state.GetHighestCompletedBlock())
	h.LeaderHeight = int64(state.GetLLeaderHeight())
	h.EntryBlockHeight = int64(state.GetHighestCompletedBlock())
	h.EntryHeight = int64(state.GetEntryDBHeightComplete())
	h.MissingEntryCount = int64(state.GetMissingEntryCount())

	return h, nil
}

func HandleV2Properties(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	vtos := func(f int) string {
		v0 := f / 1000000000
		v1 := (f % 1000000000) / 1000000
		v2 := (f % 1000000) / 1000
		v3 := f % 1000

		return fmt.Sprintf("%d.%d.%d.%d", v0, v1, v2, v3)
	}

	p := new(PropertiesResponse)
	p.FactomdVersion = vtos(state.GetFactomdVersion())
	p.ApiVersion = API_VERSION
	return p, nil
}

func HandleV2SendRawMessage(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	r := new(SendRawMessageRequest)
	err := MapToObject(params, r)
	if err != nil {
		return nil, NewInvalidParamsError()
	}
	data, err := hex.DecodeString(r.Message)
	if err != nil {
		return nil, NewInvalidParamsError()
	}

	_, msg, err := messages.UnmarshalMessageData(data)
	if err != nil {
		return nil, NewInvalidParamsError()
	}

	state.APIQueue() <- msg

	resp := new(SendRawMessageResponse)
	resp.Message = "Successfully sent the message"

	return resp, nil
}

func HandleV2GetTranasction(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	hashkey := new(HashRequest)
	err := MapToObject(params, hashkey)
	if err != nil {
		return nil, NewInvalidParamsError()
	}
	h, err := primitives.HexToHash(hashkey.Hash)
	if err != nil {
		return nil, NewInvalidHashError()
	}

	fTx, err := state.FetchFactoidTransactionByHash(h)
	if err != nil {
		if err.Error() != "Block not found, should not happen" {
			return nil, NewInternalError()
		}
	}

	ecTx, err := state.FetchECTransactionByHash(h)
	if err != nil {
		if err.Error() != "Block not found, should not happen" {
			return nil, NewInternalError()
		}
	}

	e, err := state.FetchEntryByHash(h)
	if err != nil {
		return nil, NewInternalError()
	}

	dbase := state.GetAndLockDB()
	defer state.UnlockDB()

	if fTx == nil {
		fTx, err = dbase.FetchFactoidTransaction(h)
		if err != nil {
			if err.Error() != "Block not found, should not happen" {
				return nil, NewInternalError()
			}
		}
	}

	if ecTx == nil {
		ecTx, err = dbase.FetchECTransaction(h)
		if err != nil {
			if err.Error() != "Block not found, should not happen" {
				return nil, NewInternalError()
			}
		}
	}

	if e == nil {
		e, err = dbase.FetchEntry(h)
		if err != nil {
			return nil, NewInternalError()
		}
	}

	blockHash, err := dbase.FetchIncludedIn(h)
	if err != nil {
		return nil, NewInternalError()
	}

	answer := new(TransactionResponse)
	answer.ECTranasction = ecTx
	answer.FactoidTransaction = fTx
	answer.Entry = e

	answer.IncludedInTransactionBlock = blockHash.String()

	blockHash, err = dbase.FetchIncludedIn(blockHash)
	if err != nil {
		return nil, NewInternalError()
	}

	answer.IncludedInDirectoryBlock = blockHash.String()

	dBlock, err := dbase.FetchDBlock(blockHash)
	if err != nil {
		return nil, NewInternalError()
	}
	answer.IncludedInDirectoryBlockHeight = int64(dBlock.GetDatabaseHeight())

	return answer, nil
}
