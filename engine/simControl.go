// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/controlPanel"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/wsapi"
)

var _ = fmt.Print
var sortByID bool
var verboseFaultOutput = false
var verboseAuthoritySet = false
var verboseAuthorityDeltas = false
var totalServerFaults int
var lastcmd []string

// Used for signing messages
var LOCAL_NET_PRIV_KEY string = "4c38c72fc5cdad68f13b74674d3ffb1f3d63a112710868c9b08946553448d26d"

func SimControl(listenTo int) {
	var _ = time.Sleep
	var summary int
	var watchPL int
	var watchMessages int
	var rotate int
	var wsapiNode int
	var faulting bool

	for {
		l := make([]byte, 100)
		var err error
		// When running as a detatched process, this routine becomes a very tight loop and starves other goroutines.
		// So, we will sleep before letting it check to see if Stdin has been reconnected
		if _, err = os.Stdin.Read(l); err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		// This splits up the command at anycodepoint that is not a letter, number or punctuation, so usually by spaces.
		parseFunc := func(c rune) bool {
			return !unicode.IsLetter(c) && !unicode.IsNumber(c) && !unicode.IsPunct(c)
		}
		// cmd is not a list of the parameters, much like command line args show up in args[]
		cmd := strings.FieldsFunc(string(l), parseFunc)
		// fmt.Printf("Parsing command, found %d elements.  The first element is: %+v / %s \n Full command: %+v\n", len(cmd), b[0], string(b), cmd)

		switch {
		case 0 < len(cmd):
			lastcmd = cmd
		default:
			switch {
			case 0 < len(lastcmd):
				cmd = lastcmd
			default: // no last commands
				cmd = []string{"?"}
			}
		}
		b := string(cmd[0])

		v, err := strconv.Atoi(string(b))
		if err == nil && v >= 0 && v < len(fnodes) && fnodes[listenTo].State != nil {
			listenTo = v
			os.Stderr.WriteString(fmt.Sprintf("Switching to Node %d\n", listenTo))
			// Update which node will be displayed on the controlPanel page
			connectionMetricsChannel := make(chan interface{}, p2p.StandardChannelSize)
			go controlPanel.ServeControlPanel(fnodes[listenTo].State.ControlPanelChannel, fnodes[listenTo].State, connectionMetricsChannel, p2pNetwork, Build)
		} else {
			switch {
			case '!' == b[0]:
				if listenTo < 0 || listenTo > len(fnodes) {
					break
				}
				s := fnodes[listenTo].State
				os.Stderr.WriteString("Reset Node: " + s.FactomNodeName + "\n")
				s.Reset()

			case 'g' == b[0]:
				if len(b) > 1 {
					if b[1] == 'c' {
						copyOver(fnodes[listenTo].State)
						break
					}
				}
				if listenTo < 0 || listenTo > len(fnodes) {
					break
				}
				wsapiNode = listenTo
				wsapi.SetState(fnodes[wsapiNode].State)

				if nextAuthority == -1 {
					err := fundWallet(fnodes[wsapiNode].State, 2e7)
					if err != nil {
						os.Stderr.WriteString(fmt.Sprintf("Error in funding the wallet, %s\n", err.Error()))
						break
					}
					setUpAuthorites(fnodes[wsapiNode].State, true)
					os.Stderr.WriteString(fmt.Sprintf("%d Authorities added to the stack and funds are in wallet\n", len(authStack)))
				}
				if len(b) == 1 {
					os.Stderr.WriteString(fmt.Sprint("Authorities are ready to be made. 'gN' where N is the number to be made.\n"))
				}
				if len(b) > 1 {
					count, err := strconv.Atoi(b[1:])
					if err != nil {
						os.Stderr.WriteString(fmt.Sprintf("Error in input bN, %s\n", err.Error()))
					} else {
						if count > 100 {
							os.Stderr.WriteString(fmt.Sprint("You can only pop a max of 100 off the stack at a time."))
							count = 100
						}
						err := fundWallet(fnodes[wsapiNode].State, uint64(count*5e7))
						if err != nil {
							os.Stderr.WriteString(fmt.Sprintf("Error in funding the wallet, %s\n", err.Error()))
							break
						}
						auths, skipped, err := authorityToBlockchain(count, fnodes[wsapiNode].State)
						if err != nil {
							os.Stderr.WriteString(fmt.Sprintf("Error making authorites, %s\n", err.Error()))
						}
						os.Stderr.WriteString(fmt.Sprintf("=== %d Identities added to blockchain, %d remain in stack, %d skipped (Added by someone else) ===\n", len(auths), len(authStack), skipped))
						for _, ele := range auths {
							fmt.Println(ele.ChainID.String())
						}
					}
				}
			case '/' == b[0]:
				sortByID = !sortByID
				if sortByID {
					os.Stderr.WriteString("Sort Status by Chain IDs\n")
				} else {
					os.Stderr.WriteString("Sort Status by Node Name\n")
				}
			case 'w' == b[0]:
				if listenTo >= 0 && listenTo < len(fnodes) {
					wsapiNode = listenTo
					wsapi.SetState(fnodes[wsapiNode].State)
					os.Stderr.WriteString(fmt.Sprintf("--Listen to %s --\n", fnodes[wsapiNode].State.FactomNodeName))
				}
			case 's' == b[0]:

				if len(b) > 1 {
					ht, err := strconv.Atoi(string(b[1:]))
					if err != nil {
						os.Stderr.WriteString("type snnn where nnn is the number of status messages to print\n")
						break
					}

					if listenTo < 0 || listenTo > len(fnodes) {
						os.Stderr.WriteString("Select a node first\n")
						break
					}

					f := fnodes[listenTo]
					f.State.StatusMutex.Lock()
					os.Stderr.WriteString("----------------------------- " + f.State.FactomNodeName + " -------------------------------------- " + string(b) + "\n")
					l := len(f.State.StatusStrs)
					if l < ht {
						ht = l
					}
					for i := 0; i < ht; i++ {
						os.Stderr.WriteString(f.State.StatusStrs[l-1] + "\n")
						l--
					}
					f.State.StatusMutex.Unlock()
					break
				}

				summary++
				if summary%2 == 1 {
					os.Stderr.WriteString("--Print Summary On--\n")
					go printSummary(&summary, summary, &listenTo, &wsapiNode)
				} else {
					os.Stderr.WriteString("--Print Summary Off--\n")
				}
			case 'p' == b[0]:
				if len(b) > 1 {
					ht, err := strconv.Atoi(string(b[1:]))
					if err != nil {
						os.Stderr.WriteString("Dump Process List with pn  where n = blockheight, i.e. 'p10'")
						break
					}

					if listenTo < 0 || listenTo > len(fnodes) {
						os.Stderr.WriteString("Select a node first")
						break
					}

					f := fnodes[listenTo]

					pl := f.State.ProcessLists.Get(uint32(ht))
					if pl == nil {
						os.Stderr.WriteString("No Process List found")
					} else {
						fmt.Println("-----------------------------", f.State.FactomNodeName, "--------------------------------------", string(b))
						fmt.Println(pl.String())
					}

					break
				}

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
					go rotateWSAPI(&rotate, rotate, &wsapiNode)
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
					os.Stderr.WriteString(fmt.Sprintln("Select a node first"))
					break
				}
				f := fnodes[listenTo]
				os.Stderr.WriteString(fmt.Sprintln("-----------------------------", f.State.FactomNodeName, "--------------------------------------", string(b[:len(b)])))
				if len(b) < 2 {
					break
				}
				ht, err := strconv.Atoi(string(b[1:]))
				if err != nil {
					os.Stderr.WriteString(fmt.Sprintln(err, "Dump Adminblock block with an  where n = blockheight, i.e. 'a10'"))
				} else {
					msg, err := f.State.LoadDBState(uint32(ht))
					if err == nil && msg != nil {
						dsmsg := msg.(*messages.DBStateMsg)
						ABlock := dsmsg.AdminBlock
						os.Stderr.WriteString(fmt.Sprintln(ABlock.String()))
					} else {
						pl := f.State.ProcessLists.Get(uint32(ht))
						if pl == nil || pl.AdminBlock == nil {
							os.Stderr.WriteString(fmt.Sprintln("Could not find this Admin block"))
						} else {
							os.Stderr.WriteString(fmt.Sprintln(pl.AdminBlock.String()))
						}
					}
				}
			case 'e' == b[0]:
				mLog.all = false
				for _, fnode := range fnodes {
					fnode.State.SetOut(false)
				}
				if listenTo < 0 || listenTo > len(fnodes) {
					os.Stderr.WriteString(fmt.Sprintln("Select a node first"))
					break
				}
				f := fnodes[listenTo]
				os.Stderr.WriteString(fmt.Sprintln("-----------------------------", f.State.FactomNodeName, "--------------------------------------", string(b[:len(b)])))
				if len(b) < 2 {
					break
				}
				ht, err := strconv.Atoi(string(b[1:]))
				if err != nil {
					os.Stderr.WriteString(fmt.Sprintln(err, "Dump Entry Credit block with fn  where n = blockheight, i.e. 'e10'"))
				} else {
					msg, err := f.State.LoadDBState(uint32(ht))
					if err == nil && msg != nil {
						dsmsg := msg.(*messages.DBStateMsg)
						ECBlock := dsmsg.EntryCreditBlock
						os.Stderr.WriteString(fmt.Sprint(ECBlock.String()))
					} else {
						pl := f.State.ProcessLists.Get(uint32(ht))
						if pl == nil || pl.EntryCreditBlock == nil {
							os.Stderr.WriteString(fmt.Sprintln("Could not find this Entry Credit Block"))
						} else {
							os.Stderr.WriteString(fmt.Sprintln(pl.EntryCreditBlock.String()))
						}
					}
				}
			case 'f' == b[0]:
				mLog.all = false
				for _, fnode := range fnodes {
					fnode.State.SetOut(false)
				}
				if listenTo < 0 || listenTo > len(fnodes) {
					os.Stderr.WriteString(fmt.Sprintln("Select a node first"))
					break
				}
				f := fnodes[listenTo]
				os.Stderr.WriteString(fmt.Sprintln("-----------------------------", f.State.FactomNodeName, "--------------------------------------", string(b[:len(b)])))
				if len(b) < 2 {
					break
				}
				ht, err := strconv.Atoi(string(b[1:]))
				if err != nil {
					os.Stderr.WriteString(fmt.Sprintln(err, "Dump Factoid block with fn  where n = blockheight, i.e. 'f10'"))
				} else {
					msg, err := f.State.LoadDBState(uint32(ht))
					if err == nil && msg != nil {
						dsmsg := msg.(*messages.DBStateMsg)
						FBlock := dsmsg.FactoidBlock
						os.Stderr.WriteString(fmt.Sprint(FBlock.String()))
					} else {
						dbstate := f.State.DBStates.Get(ht)
						if dbstate == nil || dbstate.FactoidBlock == nil {
							os.Stderr.WriteString(fmt.Sprintln("Could not find this Factoid block"))
						} else {
							os.Stderr.WriteString(fmt.Sprint(dbstate.FactoidBlock.String()))
						}
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
				os.Stderr.WriteString(fmt.Sprintln("-----------------------------", f.State.FactomNodeName, "--------------------------------------", string(b[:len(b)])))
				if len(b) < 2 {
					break
				}
				ht, err := strconv.Atoi(string(b[1:]))
				if err != nil {
					os.Stderr.WriteString(fmt.Sprintln(err, "Dump Directory block with dn  where n = blockheight, i.e. 'd10'"))
				} else {
					msg, err := f.State.LoadDBState(uint32(ht))
					if err == nil && msg != nil {
						dsmsg := msg.(*messages.DBStateMsg)
						DBlock := dsmsg.DirectoryBlock
						os.Stderr.WriteString(fmt.Sprint(DBlock.String()))
					} else {
						pl := f.State.ProcessLists.Get(uint32(ht))
						if pl == nil || pl.DirectoryBlock == nil {
							os.Stderr.WriteString(fmt.Sprintln("Could not find this directory block"))
						} else {
							os.Stderr.WriteString(fmt.Sprintln(pl.DirectoryBlock.String()))
						}
					}
				}
			case 'v' == b[0]:
				if len(b) > 1 {
					nnn, err := strconv.Atoi(string(b[1:]))
					if err != nil || nnn < 0 || nnn > 99 {
						os.Stderr.WriteString("Specifiy a FaultWait between 0 and 100\n")
						break
					}
					for _, fn := range fnodes {
						fn.State.SetFaultWait(nnn)
						os.Stderr.WriteString(fmt.Sprintf("Setting FaultWait of %10s to %d\n", fn.State.FactomNodeName, nnn))
					}
				} else {
					if verboseFaultOutput {
						verboseFaultOutput = false
						os.Stderr.WriteString("Vnnn          Set full fault timeout to the given number of seconds. Helps debugging.\n")
					} else {
						verboseFaultOutput = true
						os.Stderr.WriteString("--VerboseFaultOutput On--\n")
					}
				}
			case 'V' == b[0]:
				if len(b) == 1 {
					os.Stderr.WriteString("Vnnn -- Set the timeout for faulting a server to nnn seconds\n")
					os.Stderr.WriteString("Vt   -- Start an automated Faulting Test, or stop a running Faulting Test\n")
					os.Stderr.WriteString("VL   -- Display all the authority sets in the Status update\n")
					os.Stderr.WriteString("Vr   -- Reset all nodes in the simulation\n")
					break
				}

				if b[1] == 't' {
					faulting = !faulting
					if faulting {
						os.Stderr.WriteString("Start Faulting Test\n")
						go faultTest(&faulting)
					} else {
						os.Stderr.WriteString("Stop Faulting Test\n")
					}
					break
				}

				// Reset Everything
				if b[1] == 'r' {
					os.Stderr.WriteString("Reset all nodes in the simulation!\n")
					for _, f := range fnodes {
						f.State.Reset()
					}
					break
				}

				if b[1] == 'l' || b[1] == 'L' {
					if verboseAuthoritySet {
						verboseAuthoritySet = false
						os.Stderr.WriteString("--VerboseAuthoritySet Off--\n")
					} else {
						verboseAuthoritySet = true
						os.Stderr.WriteString("--VerboseAuthoritySet On--\n")
					}
					break
				}

				if b[1] == 'd' || b[1] == 'D' {
					if verboseAuthorityDeltas {
						verboseAuthorityDeltas = false
						os.Stderr.WriteString("--VerboseAuthorityDeltas Off--\n")
					} else {
						verboseAuthorityDeltas = true
						os.Stderr.WriteString("--VerboseAuthorityDeltas On--\n")
					}
					break
				}

				nnn, err := strconv.Atoi(string(b[1:]))
				if err != nil || nnn < 0 || nnn > 99 {
					os.Stderr.WriteString("Specifiy a FaultTimeout between 0 and 100\n")
					break
				}
				for _, fn := range fnodes {
					fn.State.FaultTimeout = nnn
					os.Stderr.WriteString(fmt.Sprintf("Setting FaultTimeout of %10s to %d\n", fn.State.FactomNodeName, nnn))
				}

			case 'k' == b[0]:
				mLog.all = false
				for _, fnode := range fnodes {
					fnode.State.SetOut(false)
				}
				if listenTo < 0 || listenTo > len(fnodes) {
					fmt.Println("Select a node first")
					break
				}
				f := fnodes[listenTo]
				fmt.Println("-----------------------------", f.State.FactomNodeName, "--------------------------------------", string(b))
				if len(b) < 2 {
					fmt.Println("No Parms found.")
					break
				}

				parms := strings.Split(string(b[1:]), ".")
				if len(parms) >= 2 {
					os.Stderr.WriteString("Print Entry String with:  k<db #>.<entry #>\n")
				}
				db, err1 := strconv.Atoi(string(parms[0]))
				entry, err2 := strconv.Atoi(string(parms[1]))
				if err1 != nil || err2 != nil {
					os.Stderr.WriteString("Bad Parameters")
				} else {
					msg, err := f.State.LoadDBState(uint32(db))
					if err == nil && msg != nil {
						dsmsg := msg.(*messages.DBStateMsg)
						DBlock := dsmsg.DirectoryBlock
						entries := DBlock.GetDBEntries()
						if entry < len(entries) && entry >= 0 {
							fmt.Println(fmt.Sprintf("Looking for EB with KeyMR %x", entries[entry].GetKeyMR().Bytes()))
							eb1, _ := f.State.DB.FetchEBlockHead(entries[entry].GetChainID())
							if eb1 != nil {
								fmt.Println("Chain Head:")
								fmt.Println(eb1.String())
							}
							eb2, _ := f.State.DB.FetchEBlock(entries[entry].GetKeyMR())
							if eb2 != nil {
								fmt.Println("Entry Block:")
								fmt.Println(eb2.String())
							} else {
								fmt.Println("No Entry Block Found")
							}
						}
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

			case 'y' == b[0]:
				if listenTo >= 0 && listenTo < len(fnodes) {
					f := fnodes[listenTo]
					fmt.Println("Holding:")
					for k := range f.State.Holding {
						v := f.State.Holding[k]
						if v != nil {
							os.Stderr.WriteString((v.String()) + "\n")
						} else {
							os.Stderr.WriteString("<nul>\n")
						}
					}
				}

			case 'm' == b[0]:
				watchMessages++
				if watchMessages%2 == 1 {
					os.Stderr.WriteString("--Print Messages On--\n")
					go printMessages(&watchMessages, watchMessages, &listenTo)
				} else {
					os.Stderr.WriteString("--Print Messages Off--\n")
				}
			case 'M' == b[0]:
				if !fnodes[listenTo].State.MessageTally {
					os.Stderr.WriteString("--Print Message Tallies On--\n")
					fnodes[listenTo].State.MessageTally = true
				} else {
					os.Stderr.WriteString("--Print Message Tallies Off--\n")
					fnodes[listenTo].State.MessageTally = false
				}
			case 'z' == b[0]: // Add Audit server, Remove server, and Add Leader fall through to 'n', switch to next node.
				var msg interfaces.IMsg
				if len(b) > 1 && b[1] == 'a' {
					msg = messages.NewRemoveServerMsg(fnodes[listenTo].State, fnodes[listenTo].State.IdentityChainID, 1)
				} else {
					msg = messages.NewRemoveServerMsg(fnodes[listenTo].State, fnodes[listenTo].State.IdentityChainID, 0)
				}

				priv, err := primitives.NewPrivateKeyFromHex(LOCAL_NET_PRIV_KEY)
				if err != nil {
					os.Stderr.WriteString(fmt.Sprintln("Could not remove server,", err.Error()))
					break
				}
				err = msg.(*messages.RemoveServerMsg).Sign(priv)
				if err != nil {
					os.Stderr.WriteString(fmt.Sprintln("Could not remove server,", err.Error()))
					break
				}

				fnodes[listenTo].State.InMsgQueue() <- msg
				os.Stderr.WriteString(fmt.Sprintln("Attempting to remove", fnodes[listenTo].State.GetFactomNodeName(), "as a server"))

				fallthrough
			case 'o' == b[0]: // Add Audit server and Add Leader fall through to 'n', switch to next node.
				if b[0] == 'o' { // (Don't do anything if just passing along the remove server)
					if len(b) > 1 && b[1] == 'n' {
						index := 0
						for index < len(authKeyLibrary) {
							if authKeyLibrary[index].Taken == false {
								authKeyLibrary[index].Taken = true
								fnodes[listenTo].State.IdentityChainID = authKeyLibrary[index].ChainID
								key, pKey, _ := authKeyLookup(fnodes[listenTo].State.IdentityChainID)
								fnodes[listenTo].State.LocalServerPrivKey = key
								fnodes[listenTo].State.SimSetNewKeys(pKey)
								os.Stderr.WriteString(fmt.Sprintf("Identity of " + fnodes[listenTo].State.GetFactomNodeName() + " changed to [" + authKeyLibrary[index].ChainID.String()[:10] + "]\n"))
								break
							}
							index++
						}
					}

					msg := messages.NewAddServerMsg(fnodes[listenTo].State, 1)
					priv, err := primitives.NewPrivateKeyFromHex(LOCAL_NET_PRIV_KEY)
					if err != nil {
						os.Stderr.WriteString(fmt.Sprintln("Could not make an audit server,", err.Error()))
						break
					}
					err = msg.(*messages.AddServerMsg).Sign(priv)
					if err != nil {
						os.Stderr.WriteString(fmt.Sprintln("Could not make a audit server,", err.Error()))
						break
					}
					fnodes[listenTo].State.InMsgQueue() <- msg
					os.Stderr.WriteString(fmt.Sprintln("Attempting to make", fnodes[listenTo].State.GetFactomNodeName(), "a Audit Server"))
				}
				fallthrough
			case 'l' == b[0]: // Add Audit server, Remove server, and Add Leader fall through to 'n', switch to next node.
				if b[0] == 'l' { // (Don't do anything if just passing along the audit server)
					feds := fnodes[listenTo].State.LeaderPL.FedServers
					exists := false
					for _, fed := range feds {
						if fed.GetChainID().IsSameAs(fnodes[listenTo].State.IdentityChainID) {
							exists = true
						}
					}
					if len(b) > 1 && b[1] == 't' && fnodes[listenTo].State.IdentityChainID.String()[:6] != "888888" && !exists {
						index := 0
						for index < len(authKeyLibrary) {
							if authKeyLibrary[index].Taken == false {
								authKeyLibrary[index].Taken = true
								fnodes[listenTo].State.IdentityChainID = authKeyLibrary[index].ChainID
								key, pKey, _ := authKeyLookup(fnodes[listenTo].State.IdentityChainID)
								fnodes[listenTo].State.LocalServerPrivKey = key
								fnodes[listenTo].State.SimSetNewKeys(pKey)
								os.Stderr.WriteString(fmt.Sprintf("Identity of " + fnodes[listenTo].State.GetFactomNodeName() + " changed to [" + authKeyLibrary[index].ChainID.String()[:10] + "]\n"))
								break
							}
							index++
						}
						if index >= len(authKeyLibrary) {
							os.Stderr.WriteString(fmt.Sprint("Did not make a leader, ran out of identities. Type 'g1' for one more identity.\n"))
							break
						}
					}

					msg := messages.NewAddServerMsg(fnodes[listenTo].State, 0)
					priv, err := primitives.NewPrivateKeyFromHex(LOCAL_NET_PRIV_KEY)
					if err != nil {
						os.Stderr.WriteString(fmt.Sprintln("Could not make a leader,", err.Error()))
						break
					}
					err = msg.(*messages.AddServerMsg).Sign(priv)
					if err != nil {
						os.Stderr.WriteString(fmt.Sprintln("Could not make a leader,", err.Error()))
						break
					}
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
				show := 0
				amt := -1
				if len(b) > 1 && (b[1] == 'H' || b[1] == 'h') {
					os.Stderr.WriteString("------------------------   Identity Commands   --------------------------------\n")
					os.Stderr.WriteString("iH           -Show this help\n")
					os.Stderr.WriteString("gN           -Adds 'N' identities to your identity pool from the identity stack. Cannot add \n")
					os.Stderr.WriteString("              identities that another instance of Factomd has added to their identity pool.\n")
					os.Stderr.WriteString("              Each Factomd instance has its own identity pool it can use (t), but everyone\n")
					os.Stderr.WriteString("              will share the identities in the identity stack. This stack is fixed and will\n")
					os.Stderr.WriteString("              be the same each time Factomd launches. It may grow in the future. Used for testing\n")
					os.Stderr.WriteString("tN           -Attaches Nth identity in pool(0 indexed) to current node. If that identity is taken, \n")
					os.Stderr.WriteString("              will grab the next available identity in the local identity pool. Can also just type \n")
					os.Stderr.WriteString("              't' and it will grab the next available identity.\n")
					os.Stderr.WriteString("tm           -Shows the current node's identity information. \n")
					os.Stderr.WriteString("tc           -Changes the current node's signing key. \n")
					os.Stderr.WriteString("t[CHAINID]   -Attaches the identity associated with the root chainID given to current node \n")
					os.Stderr.WriteString("u             Shows the authorities being monitored for change.\n")
					os.Stderr.WriteString("i             Shows the identities being monitored for change.\n")
					os.Stderr.WriteString("i[t/m/b/a][N] Shows only the Chains, Mhash, block signing key, or anchor key up to the Nth identity\n")
					os.Stderr.WriteString("isN           Shows only Nth identity\n")
					os.Stderr.WriteString("-------------------------------------------------------------------------------\n\n")

				} else {
					if len(b) > 1 {
						if b[1] == 't' {
							show = 1
						} else if b[1] == 'm' {
							show = 2
						} else if b[1] == 'b' {
							show = 3
						} else if b[1] == 'a' {
							show = 4
						}
						if len(b) > 2 {
							amt, err = strconv.Atoi(b[2:])
							if b[1] == 's' {
								show = 5
							} else if err == nil {
							} else {
								show = 0
								amt = -1
							}
						}
					}
					if amt == -1 {
						os.Stderr.WriteString(fmt.Sprintf("=== Identity List === Total: %d Displaying: All\n", len(fnodes[listenTo].State.Identities)))

					} else if show == 5 {
						os.Stderr.WriteString(fmt.Sprintf("=== Identity List === Total: %d Displaying Only: %d\n", len(fnodes[listenTo].State.Identities), amt))
					} else {
						os.Stderr.WriteString(fmt.Sprintf("=== Identity List === Total: %d Displaying: %d\n", len(fnodes[listenTo].State.Identities), amt))
					}
					for c, i := range fnodes[listenTo].State.Identities {
						if amt != -1 && c == amt+1 {
							break
						}
						stat := returnStatString(i.Status)
						if show == 5 {
							if c != amt {

							} else {
								os.Stderr.WriteString(fmt.Sprint("-----------------------------------Identity: ", amt, "---------------------------------------\n"))
							}
						} else {
							os.Stderr.WriteString(fmt.Sprint("-----------------------------------Identity: ", c, "---------------------------------------\n"))
						}
						if show == 0 || show == 5 {
							if show == 0 || c == amt {
								os.Stderr.WriteString(fmt.Sprint("Server Status: ", stat, "\n"))
								os.Stderr.WriteString(fmt.Sprint("Identity Chain: ", i.IdentityChainID, "\n"))
								os.Stderr.WriteString(fmt.Sprint("Management Chain: ", i.ManagementChainID, "\n"))
								os.Stderr.WriteString(fmt.Sprint("Matryoshka Hash: ", i.MatryoshkaHash, "\n"))
								os.Stderr.WriteString(fmt.Sprint("Key 1: ", i.Key1, "\n"))
								os.Stderr.WriteString(fmt.Sprint("Key 2: ", i.Key2, "\n"))
								os.Stderr.WriteString(fmt.Sprint("Key 3: ", i.Key3, "\n"))
								os.Stderr.WriteString(fmt.Sprint("Key 4: ", i.Key4, "\n"))
								os.Stderr.WriteString(fmt.Sprint("Signing Key: ", i.SigningKey, "\n"))
								for _, a := range i.AnchorKeys {
									os.Stderr.WriteString(fmt.Sprintf("Anchor Key: {'%s' L%x T%x K:%x}\n", a.BlockChain, a.KeyLevel, a.KeyType, a.SigningKey))
								}
							}
						} else if show == 1 {
							os.Stderr.WriteString(fmt.Sprint("Server Salt:     ", fnodes[c].State.Salt.String()[:16], "\n"))
							os.Stderr.WriteString(fmt.Sprint("Server Status: ", stat, "\n"))
							os.Stderr.WriteString(fmt.Sprint("Identity Chain: ", i.IdentityChainID, "\n"))
							os.Stderr.WriteString(fmt.Sprint("Management Chain: ", i.ManagementChainID, "\n"))
						} else if show == 2 {
							os.Stderr.WriteString(fmt.Sprint("Server Salt:     ", fnodes[c].State.Salt.String()[:16], "\n"))
							os.Stderr.WriteString(fmt.Sprint("Server Status:   ", stat, "\n"))
							os.Stderr.WriteString(fmt.Sprint("Identity Chain:  ", i.IdentityChainID, "\n"))
							os.Stderr.WriteString(fmt.Sprint("Matryoshka Hash: ", i.MatryoshkaHash, "\n"))
						} else if show == 3 {
							os.Stderr.WriteString(fmt.Sprint("Server Salt:     ", fnodes[c].State.Salt.String()[:16], "\n"))
							os.Stderr.WriteString(fmt.Sprint("Server Status: ", stat, "\n"))
							os.Stderr.WriteString(fmt.Sprint("Identity Chain: ", i.IdentityChainID, "\n"))
							os.Stderr.WriteString(fmt.Sprint("Signing Key: ", i.SigningKey, "\n"))
						} else if show == 4 {
							os.Stderr.WriteString(fmt.Sprint("Server Salt:     ", fnodes[c].State.Salt.String()[:16], "\n"))
							os.Stderr.WriteString(fmt.Sprint("Server Status: ", stat, "\n"))
							os.Stderr.WriteString(fmt.Sprint("Identity Chain: ", i.IdentityChainID, "\n"))
							for _, a := range i.AnchorKeys {
								os.Stderr.WriteString(fmt.Sprintf("Anchor Key: {'%s' L%x T%x K:%x}\n", a.BlockChain, a.KeyLevel, a.KeyType, a.SigningKey))
							}
						}
					}
				}
			case 't' == b[0]:
				if len(b) == 2 && b[1] == 'm' {
					_, _, auth := authKeyLookup(fnodes[listenTo].State.IdentityChainID)
					if auth == nil {
						break
					}
					fullSk := []byte{0x4d, 0xb6, 0xc9}
					fullSk = append(fullSk[:], auth.Sk1[:32]...)
					shadSk := shad(fullSk)
					fullSk = append(fullSk[:], shadSk[:4]...)

					os.Stderr.WriteString(fmt.Sprint("Identity of Current Node Information\n"))
					os.Stderr.WriteString(fmt.Sprintf("Server Salt:   %s\n", fnodes[listenTo].State.Salt.String()[:16]))
					os.Stderr.WriteString(fmt.Sprintf("Root Chain ID: %s\n", fnodes[listenTo].State.IdentityChainID.String()))
					os.Stderr.WriteString(fmt.Sprintf("Sub Chain ID : %s\n", auth.ManageChain))
					os.Stderr.WriteString(fmt.Sprintf("Sk1 Key (hex): %x\n", fullSk))
					os.Stderr.WriteString(fmt.Sprintf("Signing Key (hex): %s\n", fnodes[listenTo].State.SimGetSigKey()))
					p := fnodes[listenTo].State.GetServerPrivateKey()
					str := hex.EncodeToString((p.Key)[:32])
					os.Stderr.WriteString(fmt.Sprintf("Private Key (hex): %s\n", str))

					break
				} else if len(b) == 2 && b[1] == 'c' {
					_, _, auth := authKeyLookup(fnodes[listenTo].State.IdentityChainID)
					if auth == nil {
						break
					}
					wsapiNode = listenTo
					wsapi.SetState(fnodes[wsapiNode].State)
					err := fundWallet(fnodes[listenTo].State, 1e8)
					if err != nil {
						os.Stderr.WriteString(fmt.Sprintf("Error in funding the wallet, %s\n", err.Error()))
						break
					}
					newKey, err := changeSigningKey(fnodes[listenTo].State.IdentityChainID, fnodes[listenTo].State)
					if err != nil {
						os.Stderr.WriteString(fmt.Sprintf("Error: %s\n", err.Error()))
						break
					}
					fnodes[listenTo].State.LocalServerPrivKey = newKey.PrivateKeyString()
					fnodes[listenTo].State.SetPendingSigningKey(newKey)
					os.Stderr.WriteString(fmt.Sprintf("New public key for [%s]: %s\n", fnodes[listenTo].State.IdentityChainID.String()[:8], newKey.Pub.String()))
					break
				}
				index := 0
				if len(b) == 65 {
					hash, err := primitives.HexToHash(b[1:])
					if err != nil {
						os.Stderr.WriteString(fmt.Sprintf("Error: %s\n", err.Error()))
					} else {
						fnodes[listenTo].State.IdentityChainID = hash
						key, pKey, _ := authKeyLookup(fnodes[listenTo].State.IdentityChainID)
						if len(key) == 64 {
							fnodes[listenTo].State.LocalServerPrivKey = key
							fnodes[listenTo].State.SimSetNewKeys(pKey)
						}
						os.Stderr.WriteString(fmt.Sprintf("Identity of " + fnodes[listenTo].State.GetFactomNodeName() + " changed to [" + hash.String()[:10] + "]\n"))
					}
					break
				} else if len(authKeyLibrary) == 0 {
					os.Stderr.WriteString(fmt.Sprint("There are no available identities in this node. Type 'g1' to claim another identity\n"))
					break
				} else if len(b) > 1 {
					index, err = strconv.Atoi(string(b[1:]))
					if err != nil {
						os.Stderr.WriteString(fmt.Sprint("Incorrect input. bN where N is a number\n"))
						break
					}
				}
				if index >= len(authKeyLibrary) {
					os.Stderr.WriteString(fmt.Sprintf("Identity index out of bounds, only %d in the list.\n", len(authKeyLibrary)))
					break
				}
				if authKeyLibrary[index].Taken == true {
					os.Stderr.WriteString(fmt.Sprintf("Identity %d already taken, taking next available identity in list\n", index))
				}
				for index < len(authKeyLibrary) {
					if authKeyLibrary[index].Taken == false {
						authKeyLibrary[index].Taken = true
						fnodes[listenTo].State.IdentityChainID = authKeyLibrary[index].ChainID
						key, pKey, _ := authKeyLookup(fnodes[listenTo].State.IdentityChainID)
						fnodes[listenTo].State.LocalServerPrivKey = key
						fnodes[listenTo].State.SimSetNewKeys(pKey)
						os.Stderr.WriteString(fmt.Sprintf("Identity of " + fnodes[listenTo].State.GetFactomNodeName() + " changed to [" + authKeyLibrary[index].ChainID.String()[:10] + "]\n"))
						break
					}
					index++
				}
				if index >= len(authKeyLibrary) {
					os.Stderr.WriteString(fmt.Sprint("There are no more available identities in this node. Type 'g1' to claim another identity\n"))
				}
			case 'u' == b[0]:
				os.Stderr.WriteString(fmt.Sprintf("=== Authority List ===  Total: %d Displaying: All\n", len(fnodes[listenTo].State.Authorities)))
				for _, i := range fnodes[listenTo].State.Authorities {
					os.Stderr.WriteString("-------------------------------------------------------------------------------\n")
					var stat string
					stat = returnStatString(i.Status)
					os.Stderr.WriteString(fmt.Sprint("Server Status: ", stat, "\n"))
					os.Stderr.WriteString(fmt.Sprint("Identity Chain: ", i.AuthorityChainID, "\n"))
					os.Stderr.WriteString(fmt.Sprint("Management Chain: ", i.ManagementChainID, "\n"))
					os.Stderr.WriteString(fmt.Sprint("Matryoshka Hash: ", i.MatryoshkaHash, "\n"))
					os.Stderr.WriteString(fmt.Sprint("Signing Key: ", i.SigningKey.String(), "\n"))
					for _, a := range i.AnchorKeys {
						os.Stderr.WriteString(fmt.Sprintf("Anchor Key: {'%s' L%x T%x K:%x}\n", a.BlockChain, a.KeyLevel, a.KeyType, a.SigningKey))
					}
				}
			case 'q' == b[0]:
				var eHashes interface{}
				if len(b) > 1 {
					eHashes = fnodes[listenTo].State.GetPendingEntries(b[1])
				} else {
					eHashes = fnodes[listenTo].State.GetPendingEntries("")
				}
				os.Stderr.WriteString("Pending Entry Hash\n")
				os.Stderr.WriteString("------------------\n")
				//for _, eh := range eHashes {
				os.Stderr.WriteString(fmt.Sprint(eHashes, "\n"))
				//}
			case 'j' == b[0]:
				var fpl []interfaces.IPendingTransaction
				if len(b) > 1 {
					fpl = fnodes[listenTo].State.GetPendingTransactions(b[1])
				} else {
					fpl = fnodes[listenTo].State.GetPendingTransactions("")
				}
				fmt.Println(fpl)
			case 'S' == b[0]:
				nnn, err := strconv.Atoi(string(b[1:]))
				if err != nil || nnn < 0 || nnn > 999 {
					os.Stderr.WriteString("Specifiy a drop amount between 0 and 1000\n")
					break
				}
				for _, fn := range fnodes {
					fn.State.DropRate = nnn
					os.Stderr.WriteString(fmt.Sprintf("Setting drop rate of %10s to %2d.%01d\n", fn.State.FactomNodeName, nnn/10, nnn%10))
				}

			case 'O' == b[0]:
				if listenTo < 0 || listenTo > len(fnodes) {
					os.Stderr.WriteString("No Factom Node selected\n")
					break
				}
				nnn, err := strconv.Atoi(string(b[1:]))
				if err != nil || nnn < 0 || nnn > 999 {
					os.Stderr.WriteString("Specifiy a drop amount between 0 and 1000\n")
					break
				}

				fnodes[listenTo].State.DropRate = nnn
				os.Stderr.WriteString(fmt.Sprintf("Setting drop rate of %10s to %2d.%01d percent\n", fnodes[listenTo].State.FactomNodeName, nnn/10, nnn%10))

			case 'T' == b[0]:
				nn, err := strconv.Atoi(string(b[1:]))
				if err != nil || nn < 5 || nn > 800 {
					os.Stderr.WriteString("Specify a block time between 5 and 600 seconds\n")
					break
				}
				os.Stderr.WriteString(fmt.Sprint("Setting the block time for all nodes to ", nn, "\n"))
				for _, f := range fnodes {
					f.State.SetDirectoryBlockInSeconds(nn)
				}
			case 'F' == b[0]:
				nn, err := strconv.Atoi(string(b[1:]))
				nnn := int64(nn)
				if err != nil || nnn < 0 || nnn > 99999 {
					os.Stderr.WriteString("Specifiy a delay amount in milliseconds less than 100 seconds\n")
					break
				}

				for _, fn := range fnodes {
					fn.State.Delay = nnn
					os.Stderr.WriteString(fmt.Sprintf("Setting Delay on communications from %10s to %2d.%03d Seconds\n", fn.State.FactomNodeName, nnn/1000, nnn%1000))
				}

				for _, f := range fnodes {
					for _, p := range f.Peers {
						sim, ok := p.(*SimPeer)
						if ok {
							sim.Delay = nnn
						}
					}
				}

			case 'D' == b[0]:
				if listenTo < 0 || listenTo > len(fnodes) {
					os.Stderr.WriteString("No Factom Node selected\n")
					break
				}
				s := fnodes[listenTo].State
				for i, dbs := range s.DBStates.DBStates {
					if dbs == nil {
						os.Stderr.WriteString(fmt.Sprintf("%2d DBState            nil\n", i))
					} else {
						os.Stderr.WriteString(fmt.Sprintf("%2d DBState                          Eht: [%5d] IsNew[%5v]  ReadyToSave [%5v] Locked [%5v] Signed [%5v] Saved [%5v]\n%v", i,
							s.EntryDBHeightComplete,
							dbs.IsNew,
							dbs.ReadyToSave,
							dbs.Locked,
							dbs.Signed,
							dbs.Saved,
							dbs.String()))
					}
				}

			case 'h' == b[0]:
				os.Stderr.WriteString("-------------------------------------------------------------------------------\n")
				os.Stderr.WriteString("<enter>       Running Enter with nothing repeats the previous command.\n\n")
				os.Stderr.WriteString("Vtest         Run the fault test.  Faults 1 to n/2-1 servers. Waits for next block + 60 sec. Repeats.\n")
				os.Stderr.WriteString("nnn           For some number nnn < the number of nodes:  Set focus on that node\n")
				os.Stderr.WriteString("n             increment (with wrap) the node under focus.  i.e. if on 1, focus is set to 2\n")
				os.Stderr.WriteString("Vtest         Run the fault test.  Faults 1 to n/2-1 servers. Waits for next block + 60 sec. Repeats.\n")
				os.Stderr.WriteString("aN            Show Admin block    			 N. Indicate node eg:\"a5\" to shows blocks for that node.\n")
				os.Stderr.WriteString("eN            Show Entry Credit Block   N. Indicate node eg:\"f5\" to shows blocks for that node.\n")
				os.Stderr.WriteString("fN            Show Factoid block  			 N. Indicate node eg:\"f5\" to shows blocks for that node.\n")
				os.Stderr.WriteString("dN            Show Directory block			 N. Indicate node eg:\"d5\" to shows blocks for that node.\n")
				os.Stderr.WriteString("D             Print a Directory Block for blocks in DBStates.\n")
				os.Stderr.WriteString("kN.M          Show Entry Block and Chain Head.  N is the directory block, and M is the Entry in that block.\n")
				os.Stderr.WriteString("                 So K3.6 gets the directory block at height 3, and prints the entry at index 6.\n")
				os.Stderr.WriteString("y             Dump what is in the Holding Map.  Can crash, but oh well.\n")
				os.Stderr.WriteString("m             Show Messages as they are passed through the simulator.\n")
				os.Stderr.WriteString("Tnnn          Set the block time to the given number of seconds.\n")
				os.Stderr.WriteString("c             Trace the Consensus Process\n")
				os.Stderr.WriteString("s             Show the state of all nodes as their state changes in the simulator.\n")
				os.Stderr.WriteString("Snnn          Print the last nnn status messages from the current node.\n")
				os.Stderr.WriteString("p             Show the process lists and directory block states as they change.\n")
				os.Stderr.WriteString("n             Change the focus to the next node.\n")
				os.Stderr.WriteString("l             Make focused node the Leader.\n")
				os.Stderr.WriteString("lt            Attach the next available identity to node and make focused node the Leader.\n")
				os.Stderr.WriteString("z             Attempt to remove focused node as a federated server\n")
				os.Stderr.WriteString("za            Attempt to remove focused node as a audit server\n")
				os.Stderr.WriteString("o             Make focused an audit server.\n")
				os.Stderr.WriteString("x             Take the given node out of the netork or bring an offline node back in.\n")
				os.Stderr.WriteString("w             Point the WSAPI to send API calls to the current node.\n")
				os.Stderr.WriteString("iH            To learn about identity control through simulator.\n")
				os.Stderr.WriteString("gN            Adds 'N' identities to your identity pool. (Cannot add identities already taken)\n")
				os.Stderr.WriteString("tN            Attaches Nth identity in pool to current node. Can also just press 't' to grab the next\n")
				os.Stderr.WriteString("i             Shows the identities being monitored for change.\n")
				os.Stderr.WriteString("u             Shows the current Authorities (federated or audit servers)\n")
				os.Stderr.WriteString("v             Verbose Fault Debug Output\n")
				os.Stderr.WriteString("Vnnn          Set full fault timeout to the given number of seconds. Helps debugging.\n")
				os.Stderr.WriteString("Vtest         Run the fault test.  Faults 1 to n/2-1 servers. Waits for next block + 60 sec. Repeats.\n")
				os.Stderr.WriteString("Vreset        Reset all fnodes in the simulation.\n")
				os.Stderr.WriteString("!             Reset the current node with the focus (i.e. the 'f' by it)\n")
				os.Stderr.WriteString("Snnn          Set Drop Rate to nnn on everyone\n")
				os.Stderr.WriteString("Onnn          Set Drop Rate to nnn on this node\n")
				os.Stderr.WriteString("Dnnn          Set the Delay on messages from the current node to nnn milliseconds\n")
				os.Stderr.WriteString("Fnnn          Set the Delay on messages from all nodes to nnn milliseconds\n")
				os.Stderr.WriteString("/             Toggle the sort order between ChainID and Factom Node Name\n")

				//os.Stderr.WriteString("i[m/b/a][N]   Shows only the Mhash, block signing key, or anchor key up to the Nth identity\n")
				//os.Stderr.WriteString("isN           Shows only Nth identity\n")
				os.Stderr.WriteString("h or <enter>  Show help\n")
				os.Stderr.WriteString("\n")
				os.Stderr.WriteString("Commands are case sensitive.\n")
				os.Stderr.WriteString("-------------------------------------------------------------------------------\n\n")

			default:
			}
		}
	}
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
		stat = "Skeleton Identity"
	}
	return stat
}

// Allows us to scatter transactions across all nodes.
//
func rotateWSAPI(rotate *int, value int, wsapiNode *int) {
	for *rotate == value { // Only if true
		*wsapiNode = rand.Int() % len(fnodes)
		fnode := fnodes[*wsapiNode]
		wsapi.SetState(fnode.State)
		time.Sleep(3 * time.Second)
	}
}

func printProcessList(watchPL *int, value int, listenTo *int) {
	out := ""
	for *watchPL == value {
		fnode := fnodes[*listenTo]
		nprt := fnode.State.DBStates.String()
		b := fnode.State.GetHighestSavedBlk()
		fnode.State.ProcessLists.SetString = true
		nprt = nprt + fnode.State.ProcessLists.Str
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
