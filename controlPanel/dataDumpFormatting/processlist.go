package dataDumpFormatting

import (
	"github.com/FactomProject/factomd/state"
)

func RawProcessList(st state.State) string {
	if st.IdentityChainID == nil {
		return ""
	}
	b := st.GetHighestRecordedBlock()
	pl := st.ProcessLists.Get(b)
	if pl == nil {
		return ""
	}
	return pl.String()
}

func RawPrintMap(st state.State) string {
	if st.DBStates == nil || st.IdentityChainID == nil {
		return ""
	}
	nprt := ""
	b := st.GetHighestRecordedBlock()
	pl := st.ProcessLists.Get(b)
	if pl != nil {
		nprt = nprt + pl.PrintMap()
	}

	return nprt
}
