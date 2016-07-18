package dataDumpFormatting

import (
	"github.com/FactomProject/factomd/state"
)

func RawProcessList(st *state.State) string {
	b := st.GetHighestRecordedBlock()
	pl := st.ProcessLists.Get(b)
	if pl == nil {
		return ""
	}
	return pl.String()
}

func RawPrintMap(st *state.State) string {
	if st.DBStates == nil {
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
