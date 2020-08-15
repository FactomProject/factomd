package dataDumpFormatting

import (
	"fmt"

	"github.com/PaulSnow/factom2d/state"
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

func SyncingState(copyDS *state.DisplayState) string {
	str := ""
	for i := 0; i < len(copyDS.SyncingState); i++ {
		idx := (copyDS.SyncingStateCurrent - i)
		if idx < 0 {
			idx = len(copyDS.SyncingState) + idx
		}
		str += fmt.Sprintf("%3d : %s\n", i, copyDS.SyncingState[idx])
	}

	return str
}
