package dataDumpFormatting


/*func RawSummary(fnodes []*state.State) string {
	out := ""
	prt := "===SummaryStart===\n"
	for _, f := range fnodes {
		f.Status = true
	}
	for _, f := range fnodes {
		prt = prt + fmt.Sprintf("%s \n", f.ShortString())
	}

	prt = prt + messageLists(fnodes)
	prt = prt + "===SummaryEnd===\n"

	if prt != out {
		out = prt
	}

	return prt
}

func ShortSummary(fnodes []*state.State) string {
	st := fnodes[0]
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
		height := st.GetHighestRecordedBlock()
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
	prt = prt + messageLists(fnodes)
	return prt
}

func messageLists(fnodes []*state.State) string {
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
}*/
