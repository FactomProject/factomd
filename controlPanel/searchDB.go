package controlPanel

import (
	"fmt"

	"github.com/FactomProject/btcutil/base58"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/wsapi"
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
	searchJson := `{"Type":"` + ftype + `","item":` + jsonStr + "}"
	return searchJson
}

func searchDB(searchitem string, st state.State) (bool, string) {
	if len(searchitem) < 10 {
		return false, ""
	}
	switch searchitem[:2] {
	case "EC":
		hash := base58.Decode(searchitem)
		if len(hash) < 34 {
			return false, ""
		}
		var fixed [32]byte
		copy(fixed[:], hash[2:34])
		bal := fmt.Sprintf("%d", st.FactoidState.GetECBalance(fixed))
		return true, `{"Type":"EC","item":` + bal + "}"
	case "FA":
		hash := base58.Decode(searchitem)
		if len(hash) < 34 {
			return false, ""
		}
		var fixed [32]byte
		copy(fixed[:], hash[2:34])
		bal := fmt.Sprintf("%.3f", float64(st.FactoidState.GetFactoidBalance(fixed))/1e8)
		return true, `{"Type":"FA","item":` + bal + "}"
	}
	if len(searchitem) == 64 {
		hash, err := primitives.HexToHash(searchitem)
		if err != nil {
			return false, ""
		}
		// Search for Entry
		if entry, err := st.DB.FetchEntry(hash); err == nil && entry != nil {
			resp := newSearchResponse("entry", entry)
			if len(resp) > 1 {
				return true, resp
			}
		}
		// Search for Chain
		if mr, err := st.DB.FetchHeadIndexByChainID(hash); err == nil && mr != nil {
			resp := newSearchResponse("chainhead", mr)
			if len(resp) > 1 {
				return true, resp
			}
		}
		// Search for EBlock
		if eBlock, err := st.DB.FetchEBlockByPrimary(hash); err == nil && eBlock != nil {
			resp := newSearchResponse("eblock", eBlock)
			if len(resp) > 1 {
				return true, resp
			}
		}
		// Search for DBlock
		if dBlock, err := st.DB.FetchDBlockByPrimary(hash); err == nil && dBlock != nil {
			resp := newSearchResponse("dblock", dBlock)
			if len(resp) > 1 {
				return true, resp
			}
		}
		// Search for ABlock
		if aBlock, err := st.DB.FetchABlock(hash); err == nil && aBlock != nil {
			resp := newSearchResponse("ablock", aBlock)
			if len(resp) > 1 {
				return true, resp
			}
		}
		// Search for Factoid Block
		if fBlock, err := st.DB.FetchFBlock(hash); err == nil && fBlock != nil {
			resp := newSearchResponse("fblock", fBlock)
			if len(resp) > 1 {
				return true, resp
			}
		}
		// Search for Entry Credit Block
		if ecBlock, err := st.DB.FetchECBlock(hash); err == nil && ecBlock != nil {
			resp := newSearchResponse("ecblock", ecBlock)
			if len(resp) > 1 {
				return true, resp
			}
		}

		// Search for Factoid Transaction
		/if trans, err := st.DB.FetchFactoidTransaction(hash); err == nil && trans != nil {
			resp := newSearchResponse("facttransaction", trans)
			if len(resp) > 1 {
				return true, resp
			}
		}

		// Search for Entry Credit Transaction
		if trans, err := st.DB.FetchECTransaction(hash); err == nil && trans != nil {
			resp := newSearchResponse("ectransaction", trans)
			if len(resp) > 1 {
				return true, resp
			}
		}

		// Search for Entry Transaction
		/*ackReq := new(wsapi.AckRequest)
		ackReq.TxID = hash.String()
		if entryAck, err := wsapi.HandleV2EntryACK(&st, ackReq); err == nil && entryAck != nil && len(entryAck.(*wsapi.EntryStatus).EntryHash) == 64 {
			resp := newSearchResponse("entryack", nil)
			if len(resp) > 1 {
				return true, resp
			}
		}

		// Search for Factoid Transaction
		ackReq = new(wsapi.AckRequest)
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
