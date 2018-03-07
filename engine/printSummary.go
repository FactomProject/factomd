package engine

import (
	"bytes"
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/constants"
)

func printSummary(summary *int, value int, listenTo *int, wsapiNode *int) {

	if *listenTo < 0 || *listenTo >= len(fnodes) {
		return
	}

	for *summary == value {
		PrintOneStatus(*listenTo, *wsapiNode)

		time.Sleep(time.Second)
	}
}

var out string // previous status

func PrintOneStatus(listenTo int, wsapiNode int) {
	f := fnodes[listenTo]

	prt := "===SummaryStart===\n\n"
	loc := time.Now().Location()
	nettype := ""
	if len(fnodes) > 1 {
		nettype = "Network: " + networkpattern
	}
	prt = fmt.Sprintf("%sTime: %d %s %s\n", prt, time.Now().Unix(), loc, nettype)
	for i, f := range fnodes {
		f.Index = i
	}
	var pnodes []*FactomNode
	pnodes = append(pnodes, fnodes...)
	if sortByID {
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
	time.Sleep(time.Second)
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
		list = list + fmt.Sprintf(" %3d", f.State.Commits.Len())
	}
	prt = prt + fmt.Sprintf(fmtstr, "Commits", list)
	list = ""
	// Nil pointer exception at start up -- clay
	for _, f := range pnodes {
		var i int
		if f.State.LeaderPL.NewEBlocks != nil {
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
		list = list + fmt.Sprintf(" %3d", f.State.InMsgQueue().Length())
	}
	prt = prt + fmt.Sprintf(fmtstr, "InMsgQueue", list)
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
		for i := 0; i < int(constants.NUM_MESSAGES); i++ {
			prt = prt + fmt.Sprintf("%5d ", i)
		}
		prt = prt + "\nRecd:"

		for i := 0; i < int(constants.NUM_MESSAGES); i++ {
			prt = prt + fmt.Sprintf("%5d ", f.State.GetMessageTalliesReceived(i))
		}
		prt = prt + "\nSent:"
		for i := 0; i < int(constants.NUM_MESSAGES); i++ {
			prt = prt + fmt.Sprintf("%5d ", f.State.GetMessageTalliesSent(i))
		}

	}
	prt = prt + "\n" + systemFaults(f)
	prt = prt + faultSummary()
	lastdiff := ""
	if verboseAuthoritySet {
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
	if verboseAuthorityDeltas {
		prt = prt + "AuthorityDeltas:"

		for _, f := range pnodes {
			prt = prt + "\n"
			prt = prt + f.State.FactomNodeName
			prt = prt + f.State.GetAuthorityDeltas()
			prt = prt + "\n"
		}
	}
	prt = prt + "===SummaryEnd===\n"
	if prt != out {
		fmt.Println(prt)
		out = prt
	}
}

func systemFaults(f *FactomNode) string {
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

func faultSummary() string {
	prt := ""
	headerTitle := "Faults"
	headerLabel := "Fed   "
	currentlyFaulted := "."

	for i, fnode := range fnodes {
		if verboseFaultOutput || !fnode.State.GetNetStateOff() { // Don't show fault info for nodes that are off
			b := fnode.State.GetHighestSavedBlk()
			pl := fnode.State.ProcessLists.Get(b + 1)
			if pl == nil {
				pl = fnode.State.ProcessLists.Get(b)
			}
			if pl != nil {
				if i == 0 {
					prt = prt + fmt.Sprintf("%s\n", headerTitle)
					prt = prt + fmt.Sprintf("%7s", headerLabel)
					for headerNum, _ := range pl.FedServers {
						prt = prt + fmt.Sprintf(" %3d", headerNum)
					}
					prt = prt + fmt.Sprintf("\n")
				}
				if fnode.State.Leader {
					prt = prt + fmt.Sprintf("%7s ", fnode.State.FactomNodeName)
					for _, fed := range pl.FedServers {
						currentlyFaulted = "."
						if !fed.IsOnline() {
							currentlyFaulted = "F"
						}
						prt = prt + fmt.Sprintf("%3s ", currentlyFaulted)
					}

					prt = prt + fmt.Sprintf("| Current Fault:")
					ff := pl.CurrentFault()
					if !ff.IsNil() {
						pledgeDoneString := "N"
						if ff.PledgeDone {
							pledgeDoneString = "Y"
						}
						prt = prt + fmt.Sprintf(" %x/%x:%d/%d/%d(%s)", ff.ServerID.Bytes()[:8], ff.AuditServerID.Bytes()[:8], len(ff.LocalVoteMap), ff.SignatureList.Length, ff.SigTally(fnode.State), pledgeDoneString)
					}

					prt = prt + fmt.Sprintf("| Watch VM: ")
					for i := 0; i < len(pl.FedServers); i++ {
						if pl.VMs[i].WhenFaulted > 0 {
							prt = prt + fmt.Sprintf("%d ", i)
						}
					}
					prt = prt + " "
					for i := 0; i < len(pl.FedServers); i++ {
						if pl.VMs[i].WhenFaulted > 0 {
							prt = prt + fmt.Sprintf("(%d) ", pl.VMs[i].FaultFlag)
						}
					}

					prt = prt + fmt.Sprintf("\n")
				}
			}
		}
	}
	return prt
}
