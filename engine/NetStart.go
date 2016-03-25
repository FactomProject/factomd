// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/wsapi"
	"os"
	"strconv"
	"time"
)

var _ = fmt.Print

type FactomNode struct {
	State *state.State
	Peers []interfaces.IPeer
	MLog  *MsgLog
}

var fnodes []*FactomNode
var mLog = new(MsgLog)

func NetStart(s *state.State) {

	listenTo := -1

	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Net Sim Start!!!!!")
	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Listening to Node", listenTo)
	fmt.Println(">>>>>>>>>>>>>>>>")

	AddInterruptHandler(func() {
		fmt.Print("<Break>\n")
		fmt.Print("Gracefully shutting down the server...\n")
		for _, fnode := range fnodes {
			fmt.Print("Shutting Down: ", fnode.State.FactomNodeName, "\r\n")
			fnode.State.ShutdownChan <- 0
		}
		fmt.Print("Waiting...\r\n")
		time.Sleep(3 * time.Second)
		os.Exit(0)
	})

	FactomConfigFilename := util.GetConfigFilename("m2")

	fmt.Println(fmt.Sprintf("factom config: %s", FactomConfigFilename))

	// Figure out how many nodes I am going to generate.  Default 10
	cnt := 3
	net := "long"
	if len(os.Args) > 2 {
		cnt1, err := strconv.Atoi(os.Args[1])
		if err == nil && cnt1 != 0 {
			cnt = cnt1
		}
		net = os.Args[2]
	}

	mLog.init(cnt)

	//************************************************
	// Actually setup the Network
	//************************************************

	s.LoadConfig(FactomConfigFilename)
	s.SetOut(false)
	s.Init()

	// Make cnt Factom nodes
	for i := 0; i < cnt; i++ {
		makeServer(s) // We clone s to make all of our servers
	}

	primes := []int{7, 13, 17, 23, 37, 43, 47, 53, 67, 73, 83, 97, 103}

	var p1 []int
	// Pick 3 primes
	for _, p := range primes {
		a := cnt
		b := p
		if a < b {
			a = p
			b = cnt
		}
		if a%b != 0 {
			p1 = append(p1, p)
			if len(p1) > 3 {
				break
			}
		}
	}

	fmt.Println("factomd <node count> <network config: mesh/long/loops>")

	switch net {
	case "mesh":

		h := 0
		for index, p := range p1 {
			fmt.Println()
			for i := 0; i < cnt/(index+1); i++ {
				h2 := (h + p) % cnt
				AddSimPeer(fnodes, h, h2)
				h = h2
			}
		}
	case "long":
		for i := 1; i < cnt; i++ {
			AddSimPeer(fnodes, i-1, i)
		}
	case "loops":
		for i := 1; i < cnt; i++ {
			AddSimPeer(fnodes, i-1, i)
		}
		for i := 0; i+5 < cnt; i += 6 {
			AddSimPeer(fnodes, i, i+5)
		}
		for i := 0; i+7 < cnt; i += 3 {
			AddSimPeer(fnodes, i, i+7)
		}
	}

	if len(fnodes) > listenTo && listenTo >= 0 {
		fnodes[listenTo].State.SetOut(true)
	}

	// Web API runs independent of Factom Servers
	startServers()

	go wsapi.Start(fnodes[0].State)

	var _ = time.Sleep

	for {
		fmt.Sprintf(">>>>>>>>>>>>>>")

		b := make([]byte, 100)
		var err error
		if _, err = os.Stdin.Read(b); err != nil {
			log.Fatal(err.Error())
		}
		for i, c := range b {
			if c <= 32 {
				b = b[:i]
				break
			}
		}
		v, err := strconv.Atoi(string(b))
		if err == nil && v >= 0 && v < len(fnodes) {
			for _, fnode := range fnodes {
				fnode.State.SetOut(false)
			}
			listenTo = v
			fmt.Print("\r\nSwitching to Node ", listenTo, "\r\n")
			wsapi.SetState(fnodes[listenTo].State)
		} else {
			if len(b) == 0 {
				b = append(b, 'a')
			}
			switch b[0] {
			case 'a':
				mLog.all = false
				fmt.Println("-------------------------------------------------------------------------------")
				for _, f := range fnodes {
					f.State.SetOut(false)
					fmt.Printf("%8s %s\n", f.State.FactomNodeName, f.State.ShortString())
				}
			case 'f':
				mLog.all = false
				for _, fnode := range fnodes {
					fnode.State.SetOut(false)
				}
				if listenTo < 0 || listenTo > len(fnodes) {
					fmt.Println("Select a node first")
					break
				}
				f := fnodes[listenTo]
				fmt.Println("-----------------------------", f.State.FactomNodeName, "--------------------------------------",string(b[:len(b)]))
				if len(b)<2 { break }
				ht,err := strconv.Atoi(string(b[1:]))
				if err != nil {
					fmt.Println(err,"Dump Factoid block with Fn  where n = blockheight, i.e. 'F10'")
				}else{
					msg,err := f.State.LoadDBState(uint32(ht))
					if err == nil && msg != nil{
						dsmsg := msg.(*messages.DBStateMsg)
						FBlock := dsmsg.FactoidBlock
						fmt.Printf(FBlock.String())
					}else{
						fmt.Println("Error: ",err,msg)
					}
				}
			case 'd':
				mLog.all = false
				for _, fnode := range fnodes {
					fnode.State.SetOut(false)
				}
				if listenTo < 0 || listenTo > len(fnodes) {
					fmt.Println("Select a node first")
					break
				}
				f := fnodes[listenTo]
				fmt.Println("-----------------------------", f.State.FactomNodeName, "--------------------------------------",string(b[:len(b)]))
				if len(b)<2 { break }
				ht,err := strconv.Atoi(string(b[1:]))
				if err != nil {
					fmt.Println(err,"Dump Directory block with Dn  where n = blockheight, i.e. 'D10'")
				}else{
					msg,err := f.State.LoadDBState(uint32(ht))
					if err == nil && msg != nil{
						dsmsg := msg.(*messages.DBStateMsg)
						DBlock := dsmsg.DirectoryBlock
						fmt.Printf(DBlock.String())
					}else{
						fmt.Println("Error: ",err,msg)
					}
				}
			case 'D':
				mLog.all = false
				os.Stderr.WriteString("Dump all messages\n")
				for _, fnode := range fnodes {
					fnode.State.SetOut(true)
				}
			case 'm':
				os.Stderr.WriteString(fmt.Sprintf("Print all messages for node: %d\n",listenTo))
				for _, fnode := range fnodes {
					fnode.State.SetOut(false)
				}
				fnodes[listenTo].State.SetOut(true)
				mLog.all = true
			case 27:
				mLog.all = false
				os.Stderr.WriteString((fmt.Sprint("Gracefully shutting down the servers...\r\n")))
				for _, fnode := range fnodes {
					os.Stderr.WriteString(fmt.Sprint("Shutting Down: ", fnode.State.FactomNodeName, "\r\n"))
					fnode.State.ShutdownChan <- 0
				}
				os.Stderr.WriteString("Waiting...\r\n")
				time.Sleep(time.Duration(len(fnodes)/8+1) * time.Second)
				fmt.Println()
				os.Exit(0)
			case 32:
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
			case 's', 'S':
				for _, fnode := range fnodes {
					fnode.State.SetOut(false)
				}
				mLog.all = false
				msg := messages.NewAddServerMsg(fnodes[listenTo].State)
				fnodes[listenTo].State.InMsgQueue() <- msg
				fnodes[listenTo].State.SetOut(true)
				os.Stderr.WriteString(fmt.Sprintln("Attempting to make", fnodes[listenTo].State.GetFactomNodeName(), "a Leader"))
			default:
			}
		}
	}

}

//**********************************************************************
// Functions that access variables in this method to set up Factom Nodes
// and start the servers.
//**********************************************************************
func makeServer(s *state.State) *FactomNode {
	// All other states are clones of the first state.  Which this routine
	// gets passed to it.
	newState := s

	if len(fnodes) > 0 {
		number := fmt.Sprintf("%d", len(fnodes))
		newState = s.Clone(number).(*state.State)
		newState.Init()
	}

	fnode := new(FactomNode)
	fnode.State = newState
	fnodes = append(fnodes, fnode)
	fnode.MLog = mLog

	return fnode
}

func startServers() {
	for i, fnode := range fnodes {
		if i > 0 {
			fnode.State.Init()
		}
		go NetworkProcessorNet(fnode)
		go state.LoadDatabase(fnode.State)
		go Timer(fnode.State)
		go Validator(fnode.State)
	}
}
