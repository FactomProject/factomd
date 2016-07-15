package controlPanel

import (
	"encoding/json"
	"fmt"
)

type DataDump struct {
	DataDump1 struct { // State Summary
		Dump string
	}
}

func getDataDumps() []byte {
	holder := new(DataDump)
	holder.DataDump1.Dump = printSummary()

	ret, err := json.Marshal(holder)
	if err != nil {
		return []byte(`{"list":"none"}`)
	}
	return ret
}

func printSummary() string {
	out := ""
	prt := "===SummaryStart===\n"
	for _, f := range fnodes {
		f.Status = true
	}
	for _, f := range fnodes {
		prt = prt + fmt.Sprintf("%s \n", f.ShortString())
	}

	fmtstr := "%22s%s\n"

	var list string

	list = ""
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

	prt = prt + "===SummaryEnd===\n"

	if prt != out {
		out = prt
	}

	return prt
}
