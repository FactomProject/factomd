package controlPanel

import (
	"encoding/json"
	"fmt"

	dd "github.com/FactomProject/factomd/controlPanel/dataDumpFormatting"
)

type DataDump struct {
	DataDump1 struct { // State Summary
		ShortDump string
		RawDump   string
	}
	DataDump2 struct {
		RawDump string
	}
	DataDump3 struct {
		RawDump string
	}
	DataDump4 struct {
		Authorities string
		Identities  string
		MyNode      string
	}
	DataDump5 struct {
		RawDump    string
		SortedDump string
	}
}

func GetDataDumps() []byte {
	holder := new(DataDump)
	DisplayStateMutex.RLock()
	DsCopy := DisplayState.Clone()
	DisplayStateMutex.RUnlock()

	holder.DataDump1.ShortDump = "Currently disabled"
	holder.DataDump1.RawDump = DsCopy.RawSummary

	holder.DataDump2.RawDump = DsCopy.ProcessList

	holder.DataDump3.RawDump = DsCopy.PrintMap

	holder.DataDump4.Authorities = dd.Authorities(*DsCopy)
	holder.DataDump4.Identities = dd.Identities(*DsCopy)
	holder.DataDump4.MyNode = dd.MyNodeInfo(*DsCopy)

	holder.DataDump5.RawDump = AllConnectionsString()
	holder.DataDump5.SortedDump = SortedConnectionString()

	ret, err := json.Marshal(holder)
	if err != nil {
		return []byte(`{"list":"none"}`)
	}
	return ret
}

func SortedConnectionString() string {
	arr := AllConnections.SortedConnections()
	str := ""
	for _, con := range arr {
		str += fmt.Sprintf("Connected: %v, Hash:%s, State: %s\n", con.Connected, con.Hash[:8], con.Connection.ConnectionState)
	}
	return str
}

func AllConnectionsString() string {
	str := ""
	con := AllConnections.GetConnectedCopy()
	dis := AllConnections.GetDisconnectedCopy()
	for key := range con {
		str += fmt.Sprintf("   Connected - IP:%s, ST:%s\n", con[key].PeerAddress, con[key].ConnectionState)
	}
	for key := range dis {
		str += fmt.Sprintf("Disconnected - IP:%s, ST:%s\n", dis[key].PeerAddress, dis[key].ConnectionState)
	}
	return str
}
