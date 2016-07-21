package controlPanel

import (
	"encoding/json"

	dd "github.com/FactomProject/factomd/controlPanel/dataDumpFormatting"
)

type DataDump struct {
	DataDump1 struct { // State Summary
		ShortDump string
		RawDump   string
	}
	DataDump2 struct { // Process List
		RawDump string
	}
	DataDump3 struct { // Process List
		RawDump string
	}
	DataDump4 struct { // Process List
		Authorities string
		Identities  string
		MyNode      string
	}
}

func getDataDumps() []byte {
	holder := new(DataDump)
	holder.DataDump1.ShortDump = dd.ShortSummary(fnodes)
	holder.DataDump1.RawDump = dd.RawSummary(fnodes)

	holder.DataDump2.RawDump = dd.RawProcessList(*statePointer)

	holder.DataDump3.RawDump = dd.RawPrintMap(*statePointer)

	holder.DataDump4.Authorities = dd.Authorities(*statePointer)
	holder.DataDump4.Identities = dd.Identities(*statePointer)
	holder.DataDump4.MyNode = dd.MyNodeInfo(*statePointer)

	ret, err := json.Marshal(holder)
	if err != nil {
		return []byte(`{"list":"none"}`)
	}
	return ret
}
