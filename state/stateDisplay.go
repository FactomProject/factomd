package state

import (
	"fmt"
	"github.com/FactomProject/factomd/modules/event"
)

func (s *State) stateUpdate() *event.StateUpdate {
	fnodes := []*State{s}
	nodesSummary := messageLists(fnodes)

	summary := fmt.Sprintf("===SummaryStart===\n%s \n%s\n===SummaryEnd===\n", s.ShortString(), nodesSummary)

	return &event.StateUpdate{
		NodeTime:     s.ProcessTime,
		LeaderHeight: s.LLeaderHeight,
		Summary:      summary,
	}
}

// Data Dump String Creation
func messageLists(fnodes []*State) string {
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
		list = list + fmt.Sprintf(" %3d", len(f.PrioritizedMsgQueue()))
	}
	prt = prt + fmt.Sprintf(fmtstr, "PrioritizedMsgQueue", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", f.InMsgQueue().Length())
	}
	prt = prt + fmt.Sprintf(fmtstr, "InMsgQueue", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", f.APIQueue().Length())
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
		list = list + fmt.Sprintf(" %3d", f.NetworkOutMsgQueue().Length())
	}
	prt = prt + fmt.Sprintf(fmtstr, "NetworkOutMsgQueue", list)

	list = ""
	for _, f := range fnodes {
		list = list + fmt.Sprintf(" %3d", len(f.NetworkInvalidMsgQueue()))
	}
	prt = prt + fmt.Sprintf(fmtstr, "NetworkInvalidMsgQueue", list)

	return prt
}
