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

	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/wsapi"
)

var _ = fmt.Print

func SimControl(listenTo int) {

	var _ = time.Sleep
	// fmt.Printf("%d -- SimControl.SimControl CHECKPOINT 20\n", os.Getpid())

	for {
		// fmt.Printf("%d -- SimControl.SimControl CHECKPOINT 21\n", os.Getpid())

		l := make([]byte, 100)
		var err error
		if _, err = os.Stdin.Read(l); err != nil {
			// fmt.Printf("%d -- SimControl.SimControl CHECKPOINT 22\n Error is: %+v\n", os.Getpid(), err)
			l = []byte("no command")
			time.Sleep(10 * time.Second)

			// fmt.Printf("%d -- SimControl.SimControl CHECKPOINT 23\n", os.Getpid())
		}
		// fmt.Printf("%d -- SimControl.SimControl CHECKPOINT 24\n", os.Getpid())

		// JAYJAY probably should get rid of this and do a readline plus tokenize the line for more complext commands.
		// for i, c := range b {
		// 	if c <= 32 {
		// 		b = b[:i]
		// 		break
		// 	}
		// }
		parseFunc := func(c rune) bool {
			return !unicode.IsLetter(c) && !unicode.IsNumber(c) && !unicode.IsPunct(c)
		}

		cmd := strings.FieldsFunc(string(l), parseFunc)
		if 0 == len(cmd) {
			cmd = []string{"+"}
		}
		b := string(cmd[0])
		v, err := strconv.Atoi(string(b))
		if err == nil && v >= 0 && v < len(fnodes) {
			for _, fnode := range fnodes {
				fnode.State.SetOut(false)
			}
			listenTo = v
			fmt.Print("\r\nSwitching to Node ", listenTo, "\r\n")
			wsapi.SetState(fnodes[listenTo].State)
		} else {
			// fmt.Printf("Parsing command, found %d elements.  The first element is: %+v / %s \n Full command: %+v\n", len(cmd), b[0], string(b), cmd)
			switch {
			case '+' == b[0]:
				mLog.all = false
				fmt.Println("-------------------------------------------------------------------------------")
				for _, f := range fnodes {
					f.State.SetOut(false)
					fmt.Printf("%8s %s\n", f.State.FactomNodeName, f.State.ShortString())
				}
				fmt.Println("Fed Servers:")
				for _, fnode := range fnodes {
					for _, fed := range fnode.State.GetFedServers() {
						fmt.Printf("   %10s %x %x\n", fnode.State.FactomNodeName, fnode.State.GetIdentityChainID().Bytes()[:5], fed.GetChainID().Bytes()[:5])
					}
				}
			case 0 == strings.Compare(strings.ToLower(string(b)), "a"):
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
			case 0 == strings.Compare(strings.ToLower(string(b)), "f"):
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
			case 'D' == b[0]:
				mLog.all = true
				os.Stderr.WriteString("Dump all messages\n")
				for _, fnode := range fnodes {
					fnode.State.SetOut(true)
				}
			case 0 == strings.Compare(strings.ToLower(string(b)), "m"):
				os.Stderr.WriteString(fmt.Sprintf("Print all messages for node: %d\n", listenTo))
				for _, fnode := range fnodes {
					fnode.State.SetOut(false)
				}
				fnodes[listenTo].State.SetOut(true)
				mLog.all = false
			case 32 == b[0]:
				mLog.all = false
				fnodes[listenTo].State.SetOut(false)
				listenTo++
				if listenTo >= len(fnodes) {
					listenTo = 0
				}
				fnodes[listenTo].State.SetOut(true)
				os.Stderr.WriteString("Print all messages\n")
				os.Stderr.WriteString(fmt.Sprint("\r\nSwitching to Node ", listenTo, "\r\n"))
				wsapi.SetState(fnodes[listenTo].State)
				mLog.all = false
			case 0 == strings.Compare(strings.ToLower(string(b)), "s"):
				for _, fnode := range fnodes {
					fnode.State.SetOut(false)
				}
				mLog.all = false
				msg := messages.NewAddServerMsg(fnodes[listenTo].State)
				fnodes[listenTo].State.InMsgQueue() <- msg
				os.Stderr.WriteString(fmt.Sprintln("Attempting to make", fnodes[listenTo].State.GetFactomNodeName(), "a Leader"))
			case 0 == strings.Compare(strings.ToLower(string(b)), "netpeer"):
				i1, err1 := strconv.Atoi(string(cmd[1]))
				i2, err2 := strconv.Atoi(string(cmd[2]))
				if 0 > i1 || 0 > i2 || len(fnodes) < i1 || len(fnodes) < i2 || nil != err1 || nil != err2 {
					os.Stderr.WriteString(fmt.Sprintf("Error creating net peer.  either i1: %d or i2: %d is out of range: 0 - %d or it was \ndue to error1:\n %+v \n or error 2:\n %+v\n", i1, i2, len(fnodes), err1, err2))
					continue
				}
				fmt.Printf("Adding netpeer between: %d, and %d", i1, i2)
				AddPeer(nodeStyle, fnodes, i1, i2)
			case 0 == strings.Compare(strings.ToLower(string(b)), "serve"):
				fmt.Println("--------------RemoteServe")
				RemoteServe(fnodes)
			case 0 == strings.Compare(strings.ToLower(string(b)), "connect"):
				fmt.Println("--------------RemoteConnect")
				RemoteConnect(fnodes, cmd[1])
			case '?' == b[0], 'H' == b[0], 'h' == b[0]:
				fmt.Println("-------------------------------------------------------------------------------")
				fmt.Println("+ or ENTER    Silence")
				fmt.Println("a             Show Admin blocks.")
				fmt.Println("f             Show Factoid blocks.")
				fmt.Println("d             Show Directory blocks.")
				fmt.Println("D             Dump all messages.")
				fmt.Println("m             Show all messages for the focused node.")
				fmt.Println("\" \" [space] Follow next node, print all messages from it.")
				fmt.Println("s             Make focused node the Leader.")
				fmt.Println("? or h        Show help")
				fmt.Println("netpeer A B   Add network peer connection between A & B eg: peer 1 3")
				fmt.Println("serve         Start a TCP server (address will be printed out)")
				fmt.Println("connect             Connect to a TCP Server (eg: adress from a \"serve\" command)")
				fmt.Println("-------------------------------------------------------------------------------")
			// -- add node (and give its connections or topology)

			default:
			}
		}
	}

}
