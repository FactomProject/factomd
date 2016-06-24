// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"math/rand"

	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/wsapi"
)

var _ = fmt.Print

func SimControl(listenTo int) {

	var _ = time.Sleep
	var summary int
	var watchPL int
	var watchMessages int
	var rotate int
	var wsapiNode int

	for {
		l := make([]byte, 100)
		var err error
		// When running as a detatched process, this routine becomes a very tight loop and starves other goroutines.
		// So, we will sleep before letting it check to see if Stdin has been reconnected
		if _, err = os.Stdin.Read(l); err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		// This splits up the command at anycodepoint that is not a letter, number of punctuation, so usually by spaces.
		parseFunc := func(c rune) bool {
			return !unicode.IsLetter(c) && !unicode.IsNumber(c) && !unicode.IsPunct(c)
		}
		// cmd is not a list of the parameters, much like command line args show up in args[]
		cmd := strings.FieldsFunc(string(l), parseFunc)
		if 0 == len(cmd) {
			cmd = []string{"h"}
		}
		b := string(cmd[0])
		v, err := strconv.Atoi(string(b))
		if err == nil && v >= 0 && v < len(fnodes) && fnodes[listenTo].State != nil {
			listenTo = v
			os.Stderr.WriteString(fmt.Sprintf("Switching to Node %d\n", listenTo))
		} else {
			// fmt.Printf("Parsing command, found %d elements.  The first element is: %+v / %s \n Full command: %+v\n", len(cmd), b[0], string(b), cmd)
			switch {
			case 'w' == b[0]:
				if listenTo >= 0 && listenTo < len(fnodes) {
					wsapiNode = listenTo
					wsapi.SetState(fnodes[wsapiNode].State)
					os.Stderr.WriteString(fmt.Sprintf("--Listen to %s --\n", fnodes[wsapiNode].State.FactomNodeName))
				}
			case 's' == b[0]:
				summary++
				if summary%2 == 1 {
					os.Stderr.WriteString("--Print Summary On--\n")
					go printSummary(&summary, summary, &listenTo)
				} else {
					os.Stderr.WriteString("--Print Summary Off--\n")
				}
			case 'p' == b[0]:
				watchPL++
				if watchPL%2 == 1 {
					os.Stderr.WriteString("--Print Process Lists On--\n")
					go printProcessList(&watchPL, watchPL, &listenTo)
				} else {
					os.Stderr.WriteString("--Print Process Lists Off--\n")
				}
			case 'r' == b[0]:
				rotate++
				if rotate%2 == 1 {
					os.Stderr.WriteString("--Rotate the WSAPI around the nodes--\n")
					go rotateWSAPI(&rotate, rotate)
				} else {
					os.Stderr.WriteString("--Stop Rotation of the WSAPI around the nodes.  Now --\n")
					wsapi.SetState(fnodes[wsapiNode].State)
				}
			case 'a' == b[0]:
				mLog.all = false
				for _, fnode := range fnodes {
					fnode.State.SetOut(false)
				}
				if listenTo < 0 || listenTo > len(fnodes) {
					fmt.Println("Select a node first")
					break
				}
				f := fnodes[listenTo]
				fmt.Println("-----------------------------", f.State.FactomNodeName, "--------------------------------------", string(b[:len(b)]))
				if len(b) < 2 {
					break
				}
				ht, err := strconv.Atoi(string(b[1:]))
				if err != nil {
					fmt.Println(err, "Dump Adminblock block with an  where n = blockheight, i.e. 'a10'")
				} else {
					msg, err := f.State.LoadDBState(uint32(ht))
					if err == nil && msg != nil {
						dsmsg := msg.(*messages.DBStateMsg)
						ABlock := dsmsg.AdminBlock
						fmt.Println(ABlock.String())
					} else {
						fmt.Println("Error: ", err, msg)
					}
				}
			case 'f' == b[0]:
				mLog.all = false
				for _, fnode := range fnodes {
					fnode.State.SetOut(false)
				}
				if listenTo < 0 || listenTo > len(fnodes) {
					fmt.Println("Select a node first")
					break
				}
				f := fnodes[listenTo]
				fmt.Println("-----------------------------", f.State.FactomNodeName, "--------------------------------------", string(b[:len(b)]))
				if len(b) < 2 {
					break
				}
				ht, err := strconv.Atoi(string(b[1:]))
				if err != nil {
					fmt.Println(err, "Dump Factoid block with fn  where n = blockheight, i.e. 'f10'")
				} else {
					msg, err := f.State.LoadDBState(uint32(ht))
					if err == nil && msg != nil {
						dsmsg := msg.(*messages.DBStateMsg)
						FBlock := dsmsg.FactoidBlock
						fmt.Printf(FBlock.String())
					} else {
						fmt.Println("Error: ", err, msg)
					}
				}
			case 'd' == b[0]:
				mLog.all = false
				for _, fnode := range fnodes {
					fnode.State.SetOut(false)
				}
				if listenTo < 0 || listenTo > len(fnodes) {
					fmt.Println("Select a node first")
					break
				}
				f := fnodes[listenTo]
				fmt.Println("-----------------------------", f.State.FactomNodeName, "--------------------------------------", string(b[:len(b)]))
				if len(b) < 2 {
					break
				}
				ht, err := strconv.Atoi(string(b[1:]))
				if err != nil {
					fmt.Println(err, "Dump Directory block with dn  where n = blockheight, i.e. 'd10'")
				} else {
					msg, err := f.State.LoadDBState(uint32(ht))
					if err == nil && msg != nil {
						dsmsg := msg.(*messages.DBStateMsg)
						DBlock := dsmsg.DirectoryBlock
						fmt.Printf(DBlock.String())
					} else {
						fmt.Println("Error: ", err, msg)
					}
				}
			case 'x' == b[0]:

				if listenTo >= 0 && listenTo < len(fnodes) {
					f := fnodes[listenTo]
					v := f.State.GetNetStateOff()
					if v {
						os.Stderr.WriteString("Bring " + f.State.FactomNodeName + " Back onto the network\n")
					} else {
						os.Stderr.WriteString("Take  " + f.State.FactomNodeName + " off the network\n")
					}
					f.State.SetNetStateOff(!v)
				}

			case 'm' == b[0]:
				watchMessages++
				if watchMessages%2 == 1 {
					os.Stderr.WriteString("--Print Messages On--\n")
					go printMessages(&watchMessages, watchMessages, &listenTo)
				} else {
					os.Stderr.WriteString("--Print Messages Off--\n")
				}
			case 'o' == b[0]: // Add Audit server and Add Leader fall through to 'n', switch to next node.
				msg := messages.NewAddServerMsg(fnodes[listenTo].State, 1)
				fnodes[listenTo].State.InMsgQueue() <- msg
				os.Stderr.WriteString(fmt.Sprintln("Attempting to make", fnodes[listenTo].State.GetFactomNodeName(), "a Audit Server"))
				fallthrough
			case 'l' == b[0]: // Add Audit server and Add Leader fall through to 'n', switch to next node.
				if b[0] == 'l' { // (Don't do anything if just passing along the audit server)
					msg := messages.NewAddServerMsg(fnodes[listenTo].State, 0)
					fnodes[listenTo].State.InMsgQueue() <- msg
					os.Stderr.WriteString(fmt.Sprintln("Attempting to make", fnodes[listenTo].State.GetFactomNodeName(), "a Leader"))
				}
				fallthrough
			case 'n' == b[0]:
				fnodes[listenTo].State.SetOut(false)
				listenTo++
				if listenTo >= len(fnodes) {
					listenTo = 0
				}
				fnodes[listenTo].State.SetOut(true)
				os.Stderr.WriteString(fmt.Sprint("\r\nSwitching to Node ", listenTo, "\r\n"))
			case 'c' == b[0]:
				c := !fnodes[0].State.DebugConsensus
				for _, n := range fnodes {
					n.State.DebugConsensus = fnodes[0].State.DebugConsensus
				}
				if c {
					os.Stderr.WriteString(fmt.Sprint("\r\nTrace Consensus\n"))
				} else {
					os.Stderr.WriteString(fmt.Sprint("\r\nTurn off Consensus Trace \n"))
				}

				for _, f := range fnodes {
					f.State.DebugConsensus = c
				}

			case 'i' == b[0]:

				for _, i := range fnodes[listenTo].State.Identities {
					os.Stderr.WriteString("-------------------------------------------------------------------------------\n")
					os.Stderr.WriteString(fmt.Sprint("Server Status: ", i.Status, "\n"))
					os.Stderr.WriteString(fmt.Sprint("Identity Chain: ", i.IdentityChainID, "\n"))
					os.Stderr.WriteString(fmt.Sprint("Management Chain: ", i.ManagementChainID, "\n"))
					os.Stderr.WriteString(fmt.Sprint("Matryoshka Hash: ", i.MatryoshkaHash, "\n"))
					os.Stderr.WriteString(fmt.Sprint("Key 1: ", i.Key1, "\n"))
					os.Stderr.WriteString(fmt.Sprint("Key 2: ", i.Key2, "\n"))
					os.Stderr.WriteString(fmt.Sprint("Key 3: ", i.Key3, "\n"))
					os.Stderr.WriteString(fmt.Sprint("Key 4: ", i.Key4, "\n"))
					os.Stderr.WriteString(fmt.Sprint("Signing Key: ", i.SigningKey, "\n"))
					os.Stderr.WriteString(fmt.Sprint("Anchor Key: ", i.AnchorKeys, "\n"))
				}
				//os.Stderr.WriteString(fmt.Sprint(fnodes[listenTo].State.Identities))

			case 'h' == b[0]:
				os.Stderr.WriteString("-------------------------------------------------------------------------------\n")
				os.Stderr.WriteString("h or ENTER    Shows this help\n")
				os.Stderr.WriteString("aN            Show Admin block     N. Indicate node eg:\"a5\" to shows blocks for that node.\n")
				os.Stderr.WriteString("fN            Show Factoid block   N. Indicate node eg:\"f5\" to shows blocks for that node.\n")
				os.Stderr.WriteString("dN            Show Directory block N. Indicate node eg:\"d5\" to shows blocks for that node.\n")
				os.Stderr.WriteString("m             Show Messages as they are passed through the simulator.\n")
				os.Stderr.WriteString("c             Trace the Consensus Process\n")
				os.Stderr.WriteString("s             Show the state of all nodes as their state changes in the simulator.\n")
				os.Stderr.WriteString("p             Show the process lists and directory block states as they change.\n")
				os.Stderr.WriteString("n             Change the focus to the next node.\n")
				os.Stderr.WriteString("l             Make focused node the Leader.\n")
				os.Stderr.WriteString("x             Take the given node out of the netork or bring an offline node back in.\n")
				os.Stderr.WriteString("w             Point the WSAPI to send API calls to the current node.")
				os.Stderr.WriteString("h or <enter>  Show help\n")
				os.Stderr.WriteString("\n")
				os.Stderr.WriteString("Most commands are case insensitive.\n")
				os.Stderr.WriteString("-------------------------------------------------------------------------------\n\n")

			default:
			}
		}
	}
}

// Allows us to scatter transactions across all nodes.
//
func rotateWSAPI(rotate *int, value int) {
	for *rotate == value { // Only if true
		fnode := fnodes[rand.Int()%len(fnodes)]

		wsapi.SetState(fnode.State)
		os.Stderr.WriteString("\rAPI now directed to " + fnode.State.GetFactomNodeName() + "   ")

		time.Sleep(3 * time.Second)
	}
}

func printSummary(summary *int, value int, listenTo *int) {
	out := ""

	if *listenTo < 0 || *listenTo >= len(fnodes) {
		return
	}

	for *summary == value {
		prt := "===SummaryStart===\n"
		for _, f := range fnodes {
			f.State.Status = true
		}

		time.Sleep(time.Second)

		for _, f := range fnodes {

			prt = prt + fmt.Sprintf("%s \n", f.State.ShortString())
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
			list = list + fmt.Sprintf(" %3d", len(f.State.XReview))
		}
		prt = prt + fmt.Sprintf(fmtstr, "Review", list)

		list = ""
		for _, f := range fnodes {
			list = list + fmt.Sprintf(" %3d", len(f.State.Holding))
		}
		prt = prt + fmt.Sprintf(fmtstr, "Holding", list)

		list = ""
		for _, f := range fnodes {
			list = list + fmt.Sprintf(" %3d", len(f.State.Acks))
		}
		prt = prt + fmt.Sprintf(fmtstr, "Acks", list)

		prt = prt + "\n"

		list = ""
		for _, f := range fnodes {
			list = list + fmt.Sprintf(" %3d", len(f.State.MsgQueue()))
		}
		prt = prt + fmt.Sprintf(fmtstr, "MsgQueue", list)

		list = ""
		for _, f := range fnodes {
			list = list + fmt.Sprintf(" %3d", len(f.State.InMsgQueue()))
		}
		prt = prt + fmt.Sprintf(fmtstr, "InMsgQueue", list)

		list = ""
		for _, f := range fnodes {
			list = list + fmt.Sprintf(" %3d", len(f.State.APIQueue()))
		}
		prt = prt + fmt.Sprintf(fmtstr, "APIQueue", list)

		list = ""
		for _, f := range fnodes {
			list = list + fmt.Sprintf(" %3d", len(f.State.AckQueue()))
		}
		prt = prt + fmt.Sprintf(fmtstr, "AckQueue", list)

		prt = prt + "\n"

		list = ""
		for _, f := range fnodes {
			list = list + fmt.Sprintf(" %3d", len(f.State.TimerMsgQueue()))
		}
		prt = prt + fmt.Sprintf(fmtstr, "TimerMsgQueue", list)

		list = ""
		for _, f := range fnodes {
			list = list + fmt.Sprintf(" %3d", len(f.State.NetworkOutMsgQueue()))
		}
		prt = prt + fmt.Sprintf(fmtstr, "NetworkOutMsgQueue", list)

		list = ""
		for _, f := range fnodes {
			list = list + fmt.Sprintf(" %3d", len(f.State.NetworkInvalidMsgQueue()))
		}
		prt = prt + fmt.Sprintf(fmtstr, "NetworkInvalidMsgQueue", list)

		prt = prt + "===SummaryEnd===\n"

		if prt != out {
			fmt.Println(prt)
			out = prt
		}

		time.Sleep(time.Second)
	}
}

func printProcessList(watchPL *int, value int, listenTo *int) {
	out := ""
	for *watchPL == value {
		fnode := fnodes[*listenTo]
		nprt := fnode.State.DBStates.String()
		b := fnode.State.GetHighestRecordedBlock()
		nprt = nprt + fnode.State.ProcessLists.String()
		pl := fnode.State.ProcessLists.Get(b)
		if pl != nil {
			nprt = nprt + pl.PrintMap()
			if out != nprt {
				fmt.Println(nprt)
				out = nprt
			}
		}
		time.Sleep(time.Second * 1)
	}
}

func printMessages(Messages *int, value int, listenTo *int) {
	fmt.Println("Printing Messages")
	for *Messages == value {
		fnode := fnodes[*listenTo]
		fnode.MLog.PrtMsgs(fnode.State)

		time.Sleep(2 * time.Second)
	}
}
