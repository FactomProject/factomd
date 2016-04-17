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

	for {

		l := make([]byte, 100)
		var err error
		if _, err = os.Stdin.Read(l); err != nil {
			l = []byte("no command")  // This is a hack to handle running in the background. (Eg: as a detatched process)
			// Being unable to read from StdIn gives error, this pretends like "no command" was typed, which causes nothing (unlike simply hitting return)
		}

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
				fmt.Println("-------------------------------------------------------------------------------")
			// -- add node (and give its connections or topology)

			default:
			}
		}
	}

}
