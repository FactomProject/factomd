// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi
/*
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
	"github.com/FactomProject/factomd/receipts"
	"github.com/hoisie/web"
)

func HandleV2(ctx *web.Context) {
	j, err := ParseJSON2Request("")
	if err != nil {
		HandleV2Error(ctx, j, err)
		return
	}

	state := ctx.Server.Env["state"].(interfaces.IState)

	var resp interface{}
	switch j.Method {
	case "factoid-submit":
		resp, err = HandleV2FactoidSubmit(state, params)
		break
	case "commit-chain":
		resp, err = HandleV2CommitChain(state, params)
		break
	case "reveal-chain":
		resp, err = HandleV2RevealChain(state, params)
		break
	case "commit-entry":
		resp, err = HandleV2CommitEntry(state, params)
		break
	case "reveal-entry":
		resp, err = HandleV2RevealEntry(state, params)
		break
	case "directory-block-head":
		resp, err = HandleV2DirectoryBlockHead(state, params)
		break
	case "get-raw-data":
		resp, err = HandleV2GetRaw(state, params)
		break
	case "get-receipt":
		resp, err = HandleV2GetReceipt(state, params)
		break
	case "directory-block-by-keymr":
		resp, err = HandleV2DirectoryBlock(state, params)
		break
	case "entry-block-by-keymr":
		resp, err = HandleV2EntryBlock(state, params)
		break
	case "entry-by-hash":
		resp, err = HandleV2Entry(state, params)
		break
	case "chain-head":
		resp, err = HandleV2ChainHead(state, params)
		break
	case "entry-credit-balance":
		resp, err = HandleV2EntryCreditBalance(state, params)
		break
	case "factoid-balance":
		resp, err = HandleV2FactoidBalance(state, params)
		break
	case "factoid-get-fee":
		resp, err = HandleV2GetFee(state, params)
		break
	default:
		//TODO: do
		break
	}
	if err != nil {
		HandleV2Error(ctx, j, err)
		return
	}

	resp := NewJSON2Response()
	resp.ID = j.ID
	j.Result = resp

	ctx.Write([]byte(j.String()))
}

func HandleV2Error(ctx *web.Context, j *JSON2Request, err error) {
	resp := NewJSON2Response()
	if j != nil {
		resp.ID = j.ID
	} else {
		resp.ID = ""
	}
	j.AddError(-1, err.Error())

	ctx.WriteHeader(httpBad)
	ctx.Write([]byte(j.String()))
}

func HandleV2CommitChain(state interfaces.IState, params interface{}) (interface{}, error) {
	type commitchain struct {
		CommitChainMsg string
	}

	c := new(commitchain)
	if p, err := ioutil.ReadAll(ctx.Request.Body); err != nil {
		return nil, fmt.Errorf("Bad commit message")
	} else {
		if err := json.Unmarshal(p, c); err != nil {
			return nil, fmt.Errorf("Bad commit message")
		}
	}

	commit := entryCreditBlock.NewCommitChain()
	if p, err := hex.DecodeString(c.CommitChainMsg); err != nil {
			return nil, fmt.Errorf("Bad commit message")
	} else {
		_, err := commit.UnmarshalBinaryData(p)
		if err != nil {
			return nil, fmt.Errorf("Bad commit message")
		}
	}

	msg := new(messages.CommitChainMsg)
	msg.CommitChain = commit
	msg.Timestamp = state.GetTimestamp()
	state.InMsgQueue() <- msg

	return "Chain Commit Success", nil
}

func HandleV2RevealChain(state interfaces.IState, params interface{}) (interface{}, error) {
	return HandleV2RevealEntry(state, params)
}

func HandleV2CommitEntry(state interfaces.IState, params interface{}) (interface{}, error) {
	return nil, nil
}

func HandleV2RevealEntry(state interfaces.IState, params interface{}) (interface{}, error) {
	e := params.(string)

	entry := entryBlock.NewEntry()
	if p, err := hex.DecodeString(e); err != nil {
		return nil, fmt.Errorf("Error Reveal Entry: %v", err.Error())
	} else {
		_, err := entry.UnmarshalBinaryData(p)
		if err != nil {
			return nil, fmt.Errorf("Error Reveal Entry: %v", err.Error())
		}
	}

	state := ctx.Server.Env["state"].(interfaces.IState)

	msg := new(messages.RevealEntryMsg)
	msg.Entry = entry
	msg.Timestamp = state.GetTimestamp()
	state.InMsgQueue() <- msg

	return "Entry Reveal Success", nil
}

func HandleV2DirectoryBlockHead(state interfaces.IState, params interface{}) (interface{}, error) {
	return state.GetPreviousDirectoryBlock().GetKeyMR().String(), nil
}

func HandleV2GetRaw(state interfaces.IState, params interface{}) (interface{}, error) {
	//TODO: var block interfaces.BinaryMarshallable
	hashkey := params.(string)
	d := new(RawData)

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		return nil, err
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
	} else {
		return nil, fmt.Errorf("Entry not found")
	}

	return hex.EncodeToString(b), nil
}

func HandleV2GetReceipt(state interfaces.IState, params interface{}) (interface{}, error) {
	hashkey := params.(string)

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		return nil, err
	}

	dbase := state.GetDB()

	return receipts.CreateFullReceipt(dbase, h)
}

func HandleV2DirectoryBlock(state interfaces.IState, params interface{}) (interface{}, error) {
	hashkey := params.(string)
	d := new(DBlock)

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		return nil, err
	}

	dbase := state.GetDB()

	block, err := dbase.FetchDBlockByKeyMR(h)
	if err != nil {
		return nil, err
	}
	if block == nil {
		block, err = dbase.FetchDBlockByHash(h)
		if err != nil {
			return nil, err
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

	return d
}

func HandleV2EntryBlock(state interfaces.IState, params interface{}) (interface{}, error) {
	hashkey := params.(string)
	e := new(EBlock)

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		return nil, err
	}

	dbase := state.GetDB()

	block, err := dbase.FetchEBlockByKeyMR(h)
	if err != nil {
		return nil, err
	}
	if block == nil {
		block, err = dbase.FetchEBlockByHash(h)
		if err != nil {
			return nil, err
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

	return e, nil
}

func HandleV2Entry(state interfaces.IState, params interface{}) (interface{}, error) {
	hashkey := params.(string)
	e := new(EntryStruct)

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		return nil, err
	}

	dbase := state.GetDB()

	entry, err := dbase.FetchEntryByHash(h)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, fmt.Errorf("Entry not found")
	}

	e.ChainID = entry.GetChainIDHash().String()
	e.Content = hex.EncodeToString(entry.GetContent())
	for _, v := range entry.ExternalIDs() {
		e.ExtIDs = append(e.ExtIDs, hex.EncodeToString(v))
	}

	return e, nil
}

func HandleV2ChainHead(state interfaces.IState, params interface{}) (interface{}, error) {
	hashkey := params.(string)
	c := new(CHead)

	h, err := primitives.HexToHash(hashkey)
	if err != nil {
		return nil, err
	}

	dbase := state.GetDB()

	mr, err := dbase.FetchHeadIndexByChainID(h)
	if err != nil {
		return nil, err
	}
	if mr == nil {
		return nil, fmt.Errorf("Missing Chain Head")
	}
	return mr.String(), nil
}

func HandleV2EntryCreditBalance(state interfaces.IState, params interface{}) (interface{}, error) {
	eckey := params.(string)
	var b FactoidBalance
	adr, err := primitives.HexToHash(eckey)
	if err != nil {
		return nil, err
	}
	if adr == nil {
		return nil, fmt.Errorf("Invalid Address")
	}
	if err != nil {
		return nil, err
	}
	return state.GetFactoidState().GetECBalance(adr.Fixed()), nil

}

func HandleV2GetFee(state interfaces.IState, params interface{}) (interface{}, error) {
	type x struct{ Fee int64 }
	b := new(x)
	return int64(state.GetFactoidState().GetFactoshisPerEC()), nil
}

func HandleV2FactoidSubmit(state interfaces.IState, params interface{}) (interface{}, error) {
	t := params.(string)

	msg := new(messages.FactoidTransaction)

	if p, err = hex.DecodeString(t); err != nil {
		return nil, fmt.Errorf("Unable to decode the transaction")
	}

	_, err = msg.UnmarshalTransData(p)
	if err != nil {
		return nil, err
	}

	err = state.GetFactoidState().Validate(1, msg.Transaction)
	if err != nil {
		return nil, err
	}

	state.InMsgQueue() <- msg

	return "Successfully submitted the transaction", nil

}

func HandleV2FactoidBalance(state interfaces.IState, params interface{}) (interface{}, error) {
	eckey := params.(string)

	adr, err := hex.DecodeString(eckey)
	if err == nil && len(adr) != constants.HASH_LENGTH {
		return nil, fmt.Errorf("Invalid Address")
	}
	if err == nil {
		v := int64(state.GetFactoidState().GetFactoidBalance(factoid.NewAddress(adr).Fixed()))
		return v, nil
	} else {
		return nil, err
	}
	return nil, nil
}
*/