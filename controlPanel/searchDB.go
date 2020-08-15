package controlPanel

import (
	"fmt"
	"strconv"

	"github.com/FactomProject/btcutil/base58"
	"github.com/PaulSnow/factom2d/common/primitives"
	"github.com/PaulSnow/factom2d/state"
	//"github.com/PaulSnow/factom2d/wsapi"
)

type foundItemInterface interface {
	JSONString() (string, error)
}

func newSearchResponse(ftype string, found foundItemInterface) string {
	jsonStr := ""
	if found == nil {
		jsonStr = `"none"`
	} else {
		var err error
		jsonStr, err = found.JSONString()
		if err != nil {
			jsonStr = `"none"`
		}
	}
	searchJson := `{"Type":"` + ftype + `","item":` + jsonStr + `}`
	return searchJson
}

func searchDB(searchitem string, st state.State) (bool, string) {
	if len(searchitem) < 32 {
		heightInt, err := strconv.Atoi(searchitem)
		if err != nil {
			return false, ""
		}
		height := uint32(heightInt)
		if height < DisplayState.CurrentNodeHeight {
			dbase := StatePointer.GetDB()
			dBlock, err := dbase.FetchDBlockByHeight(height)

			if err != nil {
				return false, ""
			}
			resp := `{"Type":"dblockHeight","item":"` + dBlock.GetKeyMR().String() + `"}`
			return true, resp
		}
		return false, ""
	}
	switch searchitem[:2] {
	case "EC":
		if !primitives.ValidateECUserStr(searchitem) {
			break
		}
		hash := base58.Decode(searchitem)
		if len(hash) < 34 {
			break
		}
		var fixed [32]byte
		copy(fixed[:], hash[2:34])
		bal := fmt.Sprintf("%d", st.FactoidState.GetECBalance(fixed))
		return true, `{"Type":"EC","item":` + bal + "}"
	case "FA":
		if !primitives.ValidateFUserStr(searchitem) {
			break
		}
		hash := base58.Decode(searchitem)
		if len(hash) < 34 {
			break
		}
		var fixed [32]byte
		copy(fixed[:], hash[2:34])
		bal := fmt.Sprintf("%.8f", float64(st.FactoidState.GetFactoidBalance(fixed))/1e8)
		return true, `{"Type":"FA","item":` + bal + "}"
	}
	if len(searchitem) == 64 {
		hash, err := primitives.HexToHash(searchitem)
		if err != nil {
			return false, ""
		}

		// Must unlock manually when returining. Function continues to wsapi, who needs the dbase
		dbase := st.GetDB()

		// Search for Entry
		if entry, err := dbase.FetchEntry(hash); err == nil && entry != nil {
			resp := newSearchResponse("entry", entry)
			if len(resp) > 1 {

				return true, resp
			}
		}
		// Search for Chain
		if mr, err := dbase.FetchHeadIndexByChainID(hash); err == nil && mr != nil {
			resp := newSearchResponse("chainhead", mr)
			if len(resp) > 1 {

				return true, resp
			}
		}
		// Search for EBlock
		if eBlock, err := dbase.FetchEBlock(hash); err == nil && eBlock != nil {
			resp := newSearchResponse("eblock", eBlock)
			if len(resp) > 1 {

				return true, resp
			}
		}
		// Search for DBlock
		if dBlock, err := dbase.FetchDBlock(hash); err == nil && dBlock != nil {
			resp := newSearchResponse("dblock", dBlock)
			if len(resp) > 1 {

				return true, resp
			}
		}
		// Search for ABlock
		if aBlock, err := dbase.FetchABlock(hash); err == nil && aBlock != nil {
			resp := newSearchResponse("ablock", aBlock)
			if len(resp) > 1 {

				return true, resp
			}
		}
		// Search for Factoid Block
		if fBlock, err := dbase.FetchFBlock(hash); err == nil && fBlock != nil {
			resp := newSearchResponse("fblock", fBlock)
			if len(resp) > 1 {

				return true, resp
			}
		}
		// Search for Entry Credit Block
		if ecBlock, err := dbase.FetchECBlock(hash); err == nil && ecBlock != nil {
			resp := newSearchResponse("ecblock", ecBlock)
			if len(resp) > 1 {

				return true, resp
			}
		}

		// Search for Factoid Transaction
		if trans, err := dbase.FetchFactoidTransaction(hash); err == nil && trans != nil {
			resp := newSearchResponse("facttransaction", trans)
			if len(resp) > 1 {

				return true, resp
			}
		}

		// Search for Entry Credit Transaction
		if trans, err := dbase.FetchECTransaction(hash); err == nil && trans != nil {
			resp := newSearchResponse("ectransaction", trans)
			if len(resp) > 1 {

				return true, resp
			}
		}

		// This search takes too long to make it worth it
		// Search for Entry Transaction
		/*ackReq := new(wsapi.AckRequest)
		ackReq.TxID = hash.String()
		if entryAck, err := wsapi.HandleV2EntryACK(&st, ackReq); err == nil && entryAck != nil && len(entryAck.(*wsapi.EntryStatus).EntryHash) == 64 {
			resp := newSearchResponse("entryack", nil)
			if len(resp) > 1 {
				return true, resp
			}
		}*/

		// This search takes too long to make it worth it
		// Search for Factoid Transaction
		/*ackReq = new(wsapi.AckRequest)
		ackReq.TxID = hash.String()
		if factoidAck, err := wsapi.HandleV2FactoidACK(&st, ackReq); err == nil && factoidAck != nil && factoidAck.(*wsapi.FactoidTxStatus).BlockDate > 0 {
			resp := newSearchResponse("factoidack", nil)
			if len(resp) > 1 {
				return true, resp
			}
		}*/

	}

	return false, ""
}
