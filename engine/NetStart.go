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
	"flag"
	"bufio"
	"io"
	"encoding/hex"
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

	listenToPtr := flag.Int("node", 0, "Node Number the simulator will set as the focus")
	cntPtr      := flag.Int("count", 1, "The number of nodes to generate")
	netPtr      := flag.String("net", "tree", "The default algorithm to build the network connections")
	journalPtr  := flag.String("journal","", "Rerun a Journal of messages")
	followerPtr := flag.Bool("follower",false,"If true, force node to be a follower.  Only used when replaying a journal.")
	
	flag.Parse()
	
	listenTo := *listenToPtr
	cnt      := *cntPtr
	net      := *netPtr
	journal  := *journalPtr
	follower := *followerPtr
	
	os.Stderr.WriteString(fmt.Sprintf("node     %d\n",listenTo))
	os.Stderr.WriteString(fmt.Sprintf("count    %d\n",cnt))
	os.Stderr.WriteString(fmt.Sprintf("net      %s\n",net))
	os.Stderr.WriteString(fmt.Sprintf("journal  \"%s\"\n",journal))
	os.Stderr.WriteString(fmt.Sprintf("follower \"%v\"\n",follower))
	
	
	if journal != "" {
		cnt = 1
	}
	
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

	s.LoadConfig(FactomConfigFilename)
	if journal != "" {
		if follower {
			s.NodeMode = "FULL"
		}else{
			s.NodeMode = "SERVER"
		}
	}
	s.SetOut(false)
	s.Init()
	
	mLog.init(cnt)

	//************************************************
	// Actually setup the Network
	//************************************************

	
	// Make cnt Factom nodes
	for i := 0; i < cnt; i++ {
		makeServer(s) // We clone s to make all of our servers
	}
	
	switch net {
	case "long":
		fmt.Println("Using long Network")
		for i := 1; i < cnt; i++ {
			AddSimPeer(fnodes, i-1, i)
		}
	case "loops":
		fmt.Println("Using loops Network")
		for i := 1; i < cnt; i++ {
			AddSimPeer(fnodes, i-1, i)
		}
		for i := 0; (i+17)*2 < cnt; i += 17 {
			AddSimPeer(fnodes, i%cnt, (i+5)%cnt)
		}
		for i := 0; (i+13)*2 < cnt; i += 13 {
			AddSimPeer(fnodes, i%cnt, (i+7)%cnt)
		}
	case "tree":
		index := 0
		row := 1
	treeloop:
		for i := 0; true; i++ {
			for j := 0; j <= i; j++ {
				AddSimPeer(fnodes, index, row)
				AddSimPeer(fnodes, index, row+1)
				row++
				index++
				if index >= len(fnodes) {
					break treeloop
				}
			}
			row += 1
		}
	case "circles":
		circleSize := 7
		index := 0
		for {
			AddSimPeer(fnodes, index, index+circleSize-1)
			for i := index; i < index+circleSize-1; i++ {
				AddSimPeer(fnodes, i, i+1)
			}
			index += circleSize

			AddSimPeer(fnodes, index, index-circleSize/3)
			AddSimPeer(fnodes, index, index-circleSize-circleSize*2/3)

			if index >= len(fnodes) {
				break
			}
		}
	default:
		fmt.Println("Didn't understand network type. Known types: mesh, long, circles, tree, loops.  Using a Long Network")
		for i := 1; i < cnt; i++ {
			AddSimPeer(fnodes, i-1, i)
		}

	}
	if journal != "" {
		f, err := os.Open(journal)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		r := bufio.NewReaderSize(f, 4*1024)
		for {
			str := ""
			line, isPrefix, err := r.ReadLine()
			for err == nil && isPrefix {
				str = str+string(line)
				line, isPrefix, err = r.ReadLine()
			}
			if err == io.EOF {
				break
			}
			str = str+string(line)
			
			os.Stderr.WriteString("Message: "+str+" ")
			
			str = ""
			line, isPrefix, err = r.ReadLine()
			for err == nil && isPrefix {
				str = str+string(line)
				line, isPrefix, err = r.ReadLine()
			}
			if err == io.EOF {
				break
			}
			str = str+string(line)
			
			binary,err := hex.DecodeString(str)
			if err != nil {
				fmt.Println(err)
				return
			}
			
			msg,err := messages.UnmarshalMessage(binary)
			if err != nil {
				fmt.Println(err)
				return
			}
			
			s.InMsgQueue() <- msg
		}
		startServers(false)
	}else{
		startServers(true)
	}
	
	go wsapi.Start(fnodes[0].State)

	var _ = time.Sleep

	for {

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
				b = append(b, '+')
			}
			switch b[0] {
			case '+':
				mLog.all = false
				fmt.Println("-------------------------------------------------------------------------------")
				for _, f := range fnodes {
					f.State.SetOut(false)
					fmt.Printf("%8s %s\n", f.State.FactomNodeName, f.State.ShortString())
				}
			case 'a':
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
			case 'D':
				mLog.all = false
				os.Stderr.WriteString("Dump all messages\n")
				for _, fnode := range fnodes {
					fnode.State.SetOut(true)
				}
			case 'm':
				os.Stderr.WriteString(fmt.Sprintf("Print all messages for node: %d\n", listenTo))
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

func startServers(load bool) {
	for i, fnode := range fnodes {
		if i > 0 {
			fnode.State.Init()
		}
		go NetworkProcessorNet(fnode)
		if load {
			go state.LoadDatabase(fnode.State)
		}
		go Timer(fnode.State)
		go Validator(fnode.State)
	}
}
