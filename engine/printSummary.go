package engine

import (
	"bytes"
	"fmt"
	"time"
)

func printSummary(summary *int, value int, listenTo *int, wsapiNode *int) {
	out := ""

	if *listenTo < 0 || *listenTo >= len(fnodes) {
		return
	}

	for *summary == value {
		prt := "===SummaryStart===\n\n"
		prt = fmt.Sprintf("%sTime: %d\n", prt, time.Now().Unix())

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
			if f.Index == *listenTo {
				in = "f"
			}
			if f.Index == *wsapiNode {
				api = "w"
			}

			prt = prt + fmt.Sprintf("%3d %1s%1s %s \n", i, in, api, f.State.ShortString())
		}

		if *listenTo < len(pnodes) {
			f := pnodes[*listenTo]
			prt = fmt.Sprintf("%s EB Complete %d EB Processing %d Entries Complete %d Faults %d\n", prt, f.State.EntryBlockDBHeightComplete, f.State.EntryBlockDBHeightProcessing, f.State.EntryHeightComplete, totalServerFaults)
		}

		for _, f := range pnodes {
			fctSubmits += f.State.FCTSubmits
			ecCommits += f.State.ECCommits
			eCommits += f.State.ECommits
		}

		totals := fmt.Sprintf("%d/%d/%d", fctSubmits, ecCommits, eCommits)
		prt = prt + fmt.Sprintf("%145s %20s\n", "", totals)

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
			list = list + fmt.Sprintf(" %3d", len(f.State.Commits))
		}
		prt = prt + fmt.Sprintf(fmtstr, "Commits", list)

		list = ""
		for _, f := range pnodes {
			list = list + fmt.Sprintf(" %3d", len(f.State.LeaderPL.NewEBlocks))
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
			list = list + fmt.Sprintf(" %3d", len(f.State.InMsgQueue()))
		}
		prt = prt + fmt.Sprintf(fmtstr, "InMsgQueue", list)

		list = ""
		for _, f := range pnodes {
			list = list + fmt.Sprintf(" %3d", len(f.State.APIQueue()))
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
			list = list + fmt.Sprintf(" %3d", len(f.State.NetworkOutMsgQueue()))
		}
		prt = prt + fmt.Sprintf(fmtstr, "NetworkOutMsgQueue", list)

		list = ""
		for _, f := range pnodes {
			list = list + fmt.Sprintf(" %3d", len(f.State.NetworkInvalidMsgQueue()))
		}
		prt = prt + fmt.Sprintf(fmtstr, "NetworkInvalidMsgQueue", list)

		prt = prt + "\n" + systemFaults(fnodes[*listenTo])

		prt = prt + faultSummary()

		if verboseAuthoritySet {
			for _, f := range pnodes {
				prt = prt + "\n"
				prt = prt + f.State.GetAuthoritySetString()
			}
			prt = prt + "\n"
		}

		prt = prt + "===SummaryEnd===\n"

		if prt != out {
			fmt.Println(prt)
			out = prt
		}

		time.Sleep(time.Second)
	}
}

func systemFaults(f *FactomNode) string {
	dbheight := f.State.LLeaderHeight
	pl := f.State.ProcessLists.Get(dbheight)
	if len(pl.System.List) == 0 {
		return ""
	}
	str := fmt.Sprintf("%5s %s\n", "", "System List")
	for _, ff := range pl.System.List {
		str = fmt.Sprintf("%s%8s%s\n", str, "", ff.String())
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
		if verboseFaultOutput || !fnode.State.GetNetStateOff() {
			b := fnode.State.GetHighestCompletedBlock()
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
					if pl.AmINegotiator {
						lenFaults := pl.LenFaultMap()
						if lenFaults > 0 {
							prt = prt + fmt.Sprintf("| Faults:")
							if lenFaults < 3 {
								faultIDs := pl.GetKeysFaultMap()
								for _, faultID := range faultIDs {
									faultState := pl.GetFaultState(faultID)
									if !faultState.IsNil() {
										prt = prt + fmt.Sprintf(" %x/%x:%d ", faultState.FaultCore.ServerID.Bytes()[2:5], faultState.FaultCore.AuditServerID.Bytes()[2:5], faultState.SigTally(pl.State))

										pledgeDoneString := "N"
										if faultState.PledgeDone {
											pledgeDoneString = "Y"
										}
										prt = prt + pledgeDoneString
									}
								}
							} else {
								//too many, line gets cluttered, just show totals
								faultIDs := pl.GetKeysFaultMap()
								for _, faultID := range faultIDs {
									faultState := pl.GetFaultState(faultID)
									if !faultState.IsNil() {
										//if int(faultState.FaultCore.VMIndex) == pl.NegotiatorVMIndex {
										pledgeDoneString := "N"
										if faultState.PledgeDone {
											pledgeDoneString = "Y"
										}
										prt = prt + fmt.Sprintf(" %x/%x:%d(%s)", faultState.FaultCore.ServerID.Bytes()[2:5], faultState.FaultCore.AuditServerID.Bytes()[2:5], len(faultState.VoteMap), pledgeDoneString)
										//}
									}
								}
							}

							//prt = prt + " |"
						}
					}

					prt = prt + fmt.Sprintf("\n")
				}
			}
		}
	}
	return prt
}
