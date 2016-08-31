package dataDumpFormatting

import (
	"github.com/FactomProject/factomd/state"
)

func RawProcessList(copyDS state.DisplayState) string {
	/*if st.IdentityChainID == nil {
		return ""
	}
	b := st.CurrentNodeHeight
	pl := st.ProcessLists.Get(b)
	if pl == nil {
		return ""
	}
	return pl.String()*/
	return "Currently undergoing concurrency fixes."
}

func RawPrintMap(copyDS state.DisplayState) string {
	return copyDS.PrintMap
}
