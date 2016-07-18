package controlPanel

import (
	"encoding/json"
	"fmt"
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
	}
}

func getDataDumps() []byte {
	holder := new(DataDump)
	holder.DataDump1.ShortDump = shortSummary()
	holder.DataDump1.RawDump = rawSummary()
	holder.DataDump2.RawDump = rawProcessList()
	holder.DataDump3.RawDump = rawPrintMap()
	holder.DataDump4.Authorities = authorities()
	holder.DataDump4.Identities = identities()

	ret, err := json.Marshal(holder)
	if err != nil {
		return []byte(`{"list":"none"}`)
	}
	return ret
}

func identities() string {
	prt := ""
	prt = prt + fmt.Sprintf("=== Identity List ===  Total: %d Displaying: All\n", len(st.Identities))
	for c, i := range st.Identities {
		num := fmt.Sprintf("%d", c)
		prt = prt + "------------------------------------" + num + "----------------------------------------\n"
		stat := returnStatString(i.Status)
		prt = prt + fmt.Sprint("Server Status: ", stat, "\n")
		prt = prt + fmt.Sprint("Identity Chain: ", i.IdentityChainID, "\n")
		prt = prt + fmt.Sprint("Management Chain: ", i.ManagementChainID, "\n")
		prt = prt + fmt.Sprint("Matryoshka Hash: ", i.MatryoshkaHash, "\n")
		prt = prt + fmt.Sprint("Key 1: ", i.Key1, "\n")
		prt = prt + fmt.Sprint("Key 2: ", i.Key2, "\n")
		prt = prt + fmt.Sprint("Key 3: ", i.Key3, "\n")
		prt = prt + fmt.Sprint("Key 4: ", i.Key4, "\n")
		prt = prt + fmt.Sprint("Signing Key: ", i.SigningKey, "\n")
		for _, a := range i.AnchorKeys {
			prt = prt + fmt.Sprintf("Anchor Key: {'%s' L%x T%x K:%x}\n", a.BlockChain, a.KeyLevel, a.KeyType, a.SigningKey)
		}
	}
	return prt
}

func authorities() string {
	prt := ""
	prt = prt + fmt.Sprintf("=== Authority List ===  Total: %d Displaying: All\n", len(st.Authorities))
	for c, i := range st.Authorities {
		num := fmt.Sprintf("%d", c)
		prt = prt + "--------------------------------------" + num + "---------------------------------------\n"
		var stat string
		switch i.Status {
		case 0:
			stat = "Unassigned"
		case 1:
			stat = "Federated Server"
		case 2:
			stat = "Audit Server"
		case 3:
			stat = "Full"
		case 4:
			stat = "Pending Federated Server"
		case 5:
			stat = "Pending Audit Server"
		case 6:
			stat = "Pending Full"
		case 7:
			stat = "Pending"
		}
		prt = prt + fmt.Sprint("Server Status: ", stat, "\n")
		prt = prt + fmt.Sprint("Identity Chain: ", i.AuthorityChainID, "\n")
		prt = prt + fmt.Sprint("Management Chain: ", i.ManagementChainID, "\n")
		prt = prt + fmt.Sprint("Matryoshka Hash: ", i.MatryoshkaHash, "\n")
		prt = prt + fmt.Sprint("Signing Key: ", i.SigningKey.String(), "\n")
		for _, a := range i.AnchorKeys {
			prt = prt + fmt.Sprintf("Anchor Key: {'%s' L%x T%x K:%x}\n", a.BlockChain, a.KeyLevel, a.KeyType, a.SigningKey)
		}
	}
	return prt
}

func rawProcessList() string {
	b := st.GetHighestRecordedBlock()
	pl := st.ProcessLists.Get(b)
	if pl == nil {
		return ""
	}
	return pl.String()
}

func rawPrintMap() string {
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

func shortSummary() string {
	prt := ""
	for _, f := range fnodes {
		f.Status = true
	}
	for _, f := range fnodes {
		fname := f.GetFactomNodeName()
		if len(fname) > 10 {
			fname = fname[:10]
		}
		idStr := "000000000"
		id := f.IdentityChainID
		if id != nil {
			idStr = id.String()[:10]
		}
		name := fmt.Sprintf("%s[%s]", fname, idStr)
		leader := " "
		height := f.GetHighestKnownBlock()
		dblock := st.GetDirectoryBlockByHeight(height)
		heightHash := "0000000000"
		if dblock == nil {

		} else {
			heightHash = dblock.GetFullHash().String()
		}
		if f.IsLeader() {
			leader = "L"
		}
		prt = prt + fmt.Sprintf("%22s %s DB:%d[%s] Fct:%d EC:%d E:%d\n",
			name,
			leader,
			height,
			heightHash[:6],
			f.FactoidTrans,
			f.NewEntryChains,
			f.NewEntries)
	}
	prt = prt + messageLists()
	return prt
}

func rawSummary() string {
	out := ""
	prt := "===SummaryStart===\n"
	for _, f := range fnodes {
		f.Status = true
	}
	for _, f := range fnodes {
		prt = prt + fmt.Sprintf("%s \n", f.ShortString())
	}

	prt = prt + messageLists()
	prt = prt + "===SummaryEnd===\n"

	if prt != out {
		out = prt
	}

	return prt
}

func messageLists() string {
	prt := ""
	list := ""
	fmtstr := "%22s%s\n"
	for i, _ := range fnodes {
		list = list + fmt.Sprintf(" %3d", i)
	}
	prt = prt + fmt.Sprintf(fmtstr, "", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", len(f.XReview))
	}
	prt = prt + fmt.Sprintf(fmtstr, "Review", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", len(f.Holding))
	}
	prt = prt + fmt.Sprintf(fmtstr, "Holding", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", len(f.Acks))
	}
	prt = prt + fmt.Sprintf(fmtstr, "Acks", list)

	prt = prt + "\n"

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", len(f.MsgQueue()))
	}
	prt = prt + fmt.Sprintf(fmtstr, "MsgQueue", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", len(f.InMsgQueue()))
	}
	prt = prt + fmt.Sprintf(fmtstr, "InMsgQueue", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", len(f.APIQueue()))
	}
	prt = prt + fmt.Sprintf(fmtstr, "APIQueue", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", len(f.AckQueue()))
	}
	prt = prt + fmt.Sprintf(fmtstr, "AckQueue", list)

	prt = prt + "\n"

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", len(f.TimerMsgQueue()))
	}
	prt = prt + fmt.Sprintf(fmtstr, "TimerMsgQueue", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", len(f.NetworkOutMsgQueue()))
	}
	prt = prt + fmt.Sprintf(fmtstr, "NetworkOutMsgQueue", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", len(f.NetworkInvalidMsgQueue()))
	}
	prt = prt + fmt.Sprintf(fmtstr, "NetworkInvalidMsgQueue", list)

	return prt
}

func returnStatString(i int) string {
	var stat string
	switch i {
	case 0:
		stat = "Unassigned"
	case 1:
		stat = "Federated Server"
	case 2:
		stat = "Audit Server"
	case 3:
		stat = "Full"
	case 4:
		stat = "Pending Federated Server"
	case 5:
		stat = "Pending Audit Server"
	case 6:
		stat = "Pending Full"
	case 7:
		stat = "Pending"
	}
	return stat
}
