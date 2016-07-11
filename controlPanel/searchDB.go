package controlPanel

import (
	"fmt"

	"github.com/FactomProject/btcutil/base58"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/state"
)

type foundItemInterface interface {
	JSONString() (string, error)
}

func newSearchResponse(ftype string, found foundItemInterface) string {
	jsonStr, err := found.JSONString()
	if err != nil {
		return ""
	}
	searchJson := `{"Type":"` + ftype + `","item":` + jsonStr + "}"
	return searchJson
}

func searchDB(searchitem string, st *state.State) (bool, string) {
	switch searchitem[:2] {
	case "EC":
		hash := base58.Decode(searchitem)
		if len(hash) < 34 {
			return false, ""
		}
		var fixed [32]byte
		copy(fixed[:], hash[2:34])
		bal := fmt.Sprintf("%d", st.FactoidState.GetECBalance(fixed))
		return true, newSearchResponse("EC", bal)
	case "FA":
		hash := base58.Decode(searchitem)
		if len(hash) < 34 {
			return false, ""
		}
		var fixed [32]byte
		copy(fixed[:], hash[2:34])
		bal := fmt.Sprintf("%.3f", float64(st.FactoidState.GetFactoidBalance(fixed))/1e8)
		return true, newSearchResponse("FA", bal)
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
	}

	return false, ""
}
