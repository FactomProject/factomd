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
	/*if st.DBStates == nil || st.IdentityChainID == nil {
		return ""
	}
	nprt := ""
	b := st.GetHighestRecordedBlock()
	pl := st.ProcessLists.Get(b)
	if pl != nil {
		nprt = nprt + pl.PrintMap()
	}

	return nprt*/
	return "Currently undergoing concurrency fixes."
}
