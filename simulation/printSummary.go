package simulation

import (
	"bytes"
	"fmt"
	"time"

	"github.com/FactomProject/factomd/fnode"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/globals"
)

func printSummary(summary *int, value int, listenTo *int, wsapiNode *int) {

	if *listenTo < 0 || *listenTo >= fnode.Len() {
		return
	}
	// set everyone's ID
	for i, f := range fnode.GetFnodes() {
		f.Index = i
	}

	for *summary == value {
		PrintOneStatus(*listenTo, *wsapiNode)

		time.Sleep(2 * time.Second)
	}
}

func GetSystemStatus(listenTo int, wsapiNode int) string {
	defer func() {
		recover() // KLUDGE: may induce race/condition errors while adding a node
	}()

	fnodes := fnode.GetFnodes()
	f := fnodes[listenTo]
	s := f.State
	prt := "===SummaryStart===\n\n"
	prt = fmt.Sprintf("%sTime: %d %s Elapsed time:%s\n", prt, time.Now().Unix(), time.Now().Format("2006-01-02 15:04:05"), time.Since(globals.StartTime).String())

	var stateProcessCnt, processListProcessCnt, stateUpdateState, validatorLoopSleepCnt int64

	for _, f := range fnodes {
		stateProcessCnt += f.State.StateProcessCnt
		processListProcessCnt += s.ProcessListProcessCnt
		stateUpdateState += s.StateUpdateState
		validatorLoopSleepCnt += s.ValidatorLoopSleepCnt
	}
	downscale := int64(5000 * len(fnodes))
	prt += fmt.Sprintf("P=%8d PL=%8d US=%8d Z=%8d", stateProcessCnt/downscale, processListProcessCnt/downscale, stateUpdateState/downscale, validatorLoopSleepCnt/downscale)

	var pnodes []*fnode.FactomNode
	pnodes = append(pnodes, fnodes...)
	if SortByID {
		for i := 0; i < len(pnodes)-1; i++ {
			for j := 0; j < len(pnodes)-1-i; j++ {
				if bytes.Compare(pnodes[j].State.GetIdentityChainID().Bytes(), pnodes[j+1].State.GetIdentityChainID().Bytes()) > 0 {
					pnodes[j], pnodes[j+1] = pnodes[j+1], pnodes[j]
				}
			}
		}
	}

	fctSubmits := 0
	ecCommits := 0
	eCommits := 0

	for _, f := range pnodes {
		f.State.Status = 1
	}

	prt = prt + "    " + pnodes[0].State.SummaryHeader()

	for i, f := range pnodes {
		in := ""
		api := ""
		if f.Index == listenTo {
			in = "f"
		}
		if f.Index == wsapiNode {
			api = "w"
		}

		prt = prt + fmt.Sprintf("%3d %1s%1s %s \n", i, in, api, f.State.ShortString())
	}

	if listenTo < len(fnodes) {
		prt = fmt.Sprintf("%s EB Complete %d EB Processing %d Entries Complete %d Faults %d\n", prt, f.State.EntryBlockDBHeightComplete, f.State.EntryBlockDBHeightProcessing, f.State.EntryDBHeightComplete, totalServerFaults)
	}

	sumOut := 0
	sumIn := 0
	cnt := len(f.Peers)
	for _, p := range f.Peers {
		peer, ok := p.(*SimPeer)
		if ok && peer != nil {
			sumOut += peer.BytesOut() * 8
			sumIn += peer.BytesIn() * 8
		}
	}
	if cnt > 0 {
		prt = prt + fmt.Sprintf(" #Peers: %d            Avg/Total in Kbps:   Out: %d.%03d/%d.%03d     In: %d.%03d/%d.%03d\n",
			cnt,
			sumOut/cnt/1000, sumOut/cnt%1000,
			sumOut/1000, sumOut%1000,
			sumIn/cnt/1000, sumIn/cnt%1000,
			sumIn/1000, sumIn%1000)
	}

	for _, f := range pnodes {
		fctSubmits += f.State.FCTSubmits
		ecCommits += f.State.ECCommits
		eCommits += f.State.ECommits
	}

	totals := fmt.Sprintf("%d/%d/%d", fctSubmits, ecCommits, eCommits)
	prt = prt + fmt.Sprintf("%147s %20s\n", "", totals)

	fmtstr := "%26s%s\n"

	var list string

	list = ""
	for i, _ := range pnodes {
		list = list + fmt.Sprintf(" %3d", i)
	}
	prt = prt + fmt.Sprintf(fmtstr, "", list)

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", len(f.State.XReview))
	}
	prt = prt + fmt.Sprintf(fmtstr, "Review", list)

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", len(f.State.Holding))
	}
	prt = prt + fmt.Sprintf(fmtstr, "Holding", list)

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", f.State.Hold.GetSize())
	}
	prt = prt + fmt.Sprintf(fmtstr, "DepHolding", list)

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", f.State.Commits.Len())
	}
	prt = prt + fmt.Sprintf(fmtstr, "Commits", list)

	list = ""
	// Nil pointer exception at start up -- clay
	for _, f := range pnodes {
		var i int
		if f.State.LeaderPL != nil && f.State.LeaderPL.NewEBlocks != nil {
			i = len(f.State.LeaderPL.NewEBlocks)
		} else {
			i = 0
		}
		list = list + fmt.Sprintf(" %3d", i)
	}
	prt = prt + fmt.Sprintf(fmtstr, "Pending EBs", list)

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", f.State.LeaderPL.LenNewEntries())
	}
	prt = prt + fmt.Sprintf(fmtstr, "Pending Entries", list)

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", len(f.State.Acks))
	}
	prt = prt + fmt.Sprintf(fmtstr, "Acks", list)

	prt = prt + "\n"

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", len(f.State.MsgQueue()))
	}
	prt = prt + fmt.Sprintf(fmtstr, "MsgQueue", list)

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", len(f.State.PrioritizedMsgQueue()))
	}
	prt = prt + fmt.Sprintf(fmtstr, "PrioritizedMsgQueue", list)

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", f.State.InMsgQueue().Length())
	}
	prt = prt + fmt.Sprintf(fmtstr, "InMsgQueue", list)

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", f.State.InMsgQueue2().Length())
	}
	prt = prt + fmt.Sprintf(fmtstr, "InMsgQueue2", list)

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", len(f.State.MissingMessageResponseHandler.MissingMsgRequests))
	}
	prt = prt + fmt.Sprintf(fmtstr, "MissingMsgQueue", list)

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", f.State.APIQueue().Length())
	}
	prt = prt + fmt.Sprintf(fmtstr, "APIQueue", list)

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", len(f.State.AckQueue()))
	}
	prt = prt + fmt.Sprintf(fmtstr, "AckQueue", list)

	prt = prt + "\n"

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", len(f.State.TimerMsgQueue()))
	}
	prt = prt + fmt.Sprintf(fmtstr, "TimerMsgQueue", list)

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", f.State.NetworkOutMsgQueue().Length())
	}
	prt = prt + fmt.Sprintf(fmtstr, "NetworkOutMsgQueue", list)

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", len(f.State.NetworkInvalidMsgQueue()))
	}
	prt = prt + fmt.Sprintf(fmtstr, "NetworkInvalidMsgQueue", list)

	prt = prt + "\n"

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", len(f.State.UpdateEntryHash))
	}
	prt = prt + fmt.Sprintf(fmtstr, "UpdateEntryHash", list)

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", len(f.State.MissingEntries))
	}
	prt = prt + fmt.Sprintf(fmtstr, "MissingEntries", list)

	list = ""
	for _, f := range pnodes {
		list = list + fmt.Sprintf(" %3d", len(f.State.WriteEntry))
	}
	prt = prt + fmt.Sprintf(fmtstr, "WriteEntry", list)

	if f.State.MessageTally {
		prt = prt + "\nType:"
		NumMsgTypes := int(constants.NUM_MESSAGES)
		for i := 0; i < NumMsgTypes/3; i++ {
			prt = prt + fmt.Sprintf("%8s(%2d) ", constants.ShortMessageName(byte(i)), i)
		}
		prt = prt + "\nRecd:"

		for i := 0; i < NumMsgTypes/3; i++ {
			prt = prt + fmt.Sprintf("%12d ", f.State.GetMessageTalliesReceived(i))
		}
		prt = prt + "\nSent:"
		for i := 0; i < NumMsgTypes/3; i++ {
			prt = prt + fmt.Sprintf("%12d ", f.State.GetMessageTalliesSent(i))
		}
		prt = prt + "\n\nType:"
		for i := NumMsgTypes / 3; i < 2*NumMsgTypes/3; i++ {
			prt = prt + fmt.Sprintf("%8s(%2d) ", constants.ShortMessageName(byte(i)), i)
		}
		prt = prt + "\nRecd:"

		for i := NumMsgTypes / 3; i < 2*NumMsgTypes/3; i++ {
			prt = prt + fmt.Sprintf("%12d ", f.State.GetMessageTalliesReceived(i))
		}
		prt = prt + "\nSent:"
		for i := NumMsgTypes / 3; i < 2*NumMsgTypes/3; i++ {
			prt = prt + fmt.Sprintf("%12d ", f.State.GetMessageTalliesSent(i))
		}

		prt = prt + "\n\nType:"
		for i := 2 * NumMsgTypes / 3; i < NumMsgTypes; i++ {
			prt = prt + fmt.Sprintf("%8s(%2d) ", constants.ShortMessageName(byte(i)), i)
		}
		prt = prt + "\nRecd:"

		for i := 2 * NumMsgTypes / 3; i < NumMsgTypes; i++ {
			prt = prt + fmt.Sprintf("%12d ", f.State.GetMessageTalliesReceived(i))
		}
		prt = prt + "\nSent:"
		for i := 2 * NumMsgTypes / 3; i < NumMsgTypes; i++ {
			prt = prt + fmt.Sprintf("%12d ", f.State.GetMessageTalliesSent(i))
		}

	}
	prt = prt + "\n" + SystemFaults(f)

	prt = prt + FaultSummary()

	lastdiff := ""
	if VerboseAuthoritySet {
		lastdelta := pnodes[0].State.GetAuthoritySetString()
		for i, f := range pnodes {
			prt = prt + "\n"
			ad := f.State.GetAuthoritySetString()
			diff := ""
			adiff := false
			for i := range ad {
				if i >= len(lastdelta) {
					break
				}
				if i < 8 {
					diff = diff + " "
					continue
				}
				if lastdelta[i] != ad[i] {
					diff = diff + "-"
					adiff = true
				} else {
					diff = diff + " "
				}
			}
			if adiff && i > 0 && lastdiff == diff {
				adiff = false
			}
			lastdiff = diff
			if adiff {
				diff = "\n" + diff
			} else {
				diff = ""
			}
			lastdelta = ad
			diff = "\n" /*********************************** replace Diff with new line.*/
			prt = prt + ad + diff
		}
		prt = prt + "\n"
	}

	if VerboseAuthorityDeltas {
		prt = prt + "AuthorityDeltas:"

		for _, f := range pnodes {
			prt = prt + "\n"
			prt = prt + f.State.FactomNodeName
			prt = prt + f.State.GetAuthorityDeltas()
			prt = prt + "\n"
		}
	}

	prt = prt + "===SummaryEnd===\n"
	return prt
}

var out string // previous status

func PrintOneStatus(listenTo int, wsapiNode int) {
	prt := GetSystemStatus(listenTo, wsapiNode)
	if prt != out {
		fmt.Println(prt)
		out = prt
	}

}

func SystemFaults(f *fnode.FactomNode) string {
	dbheight := f.State.LLeaderHeight
	pl := f.State.ProcessLists.Get(dbheight)
	if pl == nil {
		return ""
	}
	if len(pl.System.List) == 0 {
		str := fmt.Sprintf("%5s %13s %6s Length: 0\n", "", "System List", f.State.FactomNodeName)
		return str
	}
	str := fmt.Sprintf("%5s %s\n", "", "System List")
	for _, ff := range pl.System.List {
		if ff != nil {
			str = fmt.Sprintf("%s%8s%s\n", str, "", ff.String())
		}
	}
	str = str + "\n"
	return str
}

func FaultSummary() string {

	return ""
}
