// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/btcd"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/wsapi"
	"os"
	"time"
)

var _ = fmt.Print

type FactomNode struct {
	State 	*state.State
	Peers 	[]*FactomPeer
	MLog	*MsgLog
}

type FactomPeer struct {
	// A connection to this node:
	name string
	// Channels that define the connection:
	BroadcastOut chan interfaces.IMsg
	BroadcastIn  chan interfaces.IMsg
}

func (f *FactomPeer) init(name string) *FactomPeer {
	f.name = name
	f.BroadcastOut = make(chan interfaces.IMsg, 10000)
	return f
}

func AddPeer(f1, f2 *FactomNode) {
	peer12 := new(FactomPeer).init(f2.State.FactomNodeName)
	peer21 := new(FactomPeer).init(f1.State.FactomNodeName)
	peer12.BroadcastIn = peer21.BroadcastOut
	peer21.BroadcastIn = peer12.BroadcastOut

	f1.Peers = append(f1.Peers, peer12)
	f2.Peers = append(f2.Peers, peer21)
}

func NetStart(s *state.State) {

	var fnodes []*FactomNode

	s.SetOut(false)

	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Net Sim Start!!!!!")
	fmt.Println(">>>>>>>>>>>>>>>>")

	pcfg, _, err := btcd.LoadConfig()
	if err != nil {
		log.Println(err.Error())
	}
	FactomConfigFilename := pcfg.FactomConfigFile

	if len(FactomConfigFilename) == 0 {
		FactomConfigFilename = util.GetConfigFilename("m2")
	}
	fmt.Println(fmt.Sprintf("factom config: %s", FactomConfigFilename))

	mLog := new(MsgLog)
	mLog.init()
	
	makeServer := func() *FactomNode {
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

	startServers := func() {
		for _, fnode := range fnodes {
			go NetworkProcessorNet(fnode)
			go loadDatabase(fnode.State)
			go Timer(fnode.State)
			go Validator(fnode.State)
		}
	}

	//************************************************
	// Actually setup the Network
	//************************************************

	s.LoadConfig(FactomConfigFilename)
	s.Init()

	for i := 0; i < 10; i++ { // Make 10 nodes
		makeServer()
	}

	AddPeer(fnodes[2], fnodes[3])
	AddPeer(fnodes[1], fnodes[2])
	AddPeer(fnodes[2], fnodes[4])
	AddPeer(fnodes[4], fnodes[5])
	AddPeer(fnodes[4], fnodes[6])
	AddPeer(fnodes[5], fnodes[7])
	AddPeer(fnodes[6], fnodes[7])
	AddPeer(fnodes[7], fnodes[8])
	AddPeer(fnodes[8], fnodes[9])
	AddPeer(fnodes[0], fnodes[3])
	AddPeer(fnodes[0], fnodes[1])
	AddPeer(fnodes[0], fnodes[4])
	
	startServers()

	go wsapi.Start(fnodes[0].State)

	// Web API runs independent of Factom Servers


	var _ = time.Sleep
	p := 0
	var _ = p
	for {
		fmt.Sprintf(">>>>>>>>>>>>>>")
		
		b := make([]byte, 1)
		if _, err := os.Stdin.Read(b); err != nil {
			log.Fatal(err.Error())
		}
		
		fmt.Printf("Got: %X\n",b)
		
		switch b[0] {
			case 27:
				fmt.Print("Gracefully shutting down the server...\r\n")
				for _, fnode := range fnodes {
					fmt.Print("Shutting Down: ", fnode.State.FactomNodeName, "\r\n")
					fnode.State.ShutdownChan <- 0
				}
				fmt.Print("Waiting...\r\n")
				time.Sleep(10 * time.Second)
				os.Exit(0)
			case 32:
				fnodes[p].State.SetOut(false)
				p++
				if p >= len(fnodes) {
					p = 0
				}
				fnodes[p].State.SetOut(true)
				fmt.Print("\r\nSwitching to Node ", p,"\r\n")
				wsapi.SetState(fnodes[p].State)
			default:
		}
	}

}
