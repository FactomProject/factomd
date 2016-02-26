// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"fmt"
	"github.com/FactomProject/factomd/btcd"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/log"
	ss "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/wsapi"
	"time"
)

var _ = fmt.Print

type FactomNode struct {
	State		*ss.State
	Peers		[]*FactomPeer
	
}

type FactomPeer struct {	
	// A connection to this node:
	name 			  string 
	// Channels that define the connection:
	BroadcastOut      chan interfaces.IMsg
	BroadcastIn       chan interfaces.IMsg
}

func (f *FactomPeer) init(name string) *FactomPeer {
	f.name = name
	f.BroadcastOut = make(chan interfaces.IMsg,10000)
	return f
}

func AddPeer(f1, f2 *FactomNode) {
	peer12 := new(FactomPeer).init(f2.State.FactomNodeName)
	peer21 := new(FactomPeer).init(f1.State.FactomNodeName)
	peer12.BroadcastIn = peer21.BroadcastOut
	peer21.BroadcastIn = peer12.BroadcastOut
	
	f1.Peers = append(f1.Peers,peer12)
	f2.Peers = append(f2.Peers,peer21)
}

func NetStart(state *ss.State) {
	
	var states []*ss.State
	
	state.SetOut(false)
	
	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Net Sim Start!!!!!")
	fmt.Println(">>>>>>>>>>>>>>>>")
	
	AddInterruptHandler(func() {
		fmt.Print("<Break>\n")
		fmt.Print("Gracefully shutting down the server...\n")
		for i,one_state := range states {
			fmt.Println("Shutting Down: ",i, one_state.FactomNodeName)
			one_state.ShutdownChan <- 0
		}
		os.Exit(0)
	})
		
	pcfg, _, err := btcd.LoadConfig()
	if err != nil {
		log.Println(err.Error())
	}
	FactomConfigFilename := pcfg.FactomConfigFile
	
	if len(FactomConfigFilename) == 0 {
		FactomConfigFilename = util.GetConfigFilename("m2")
	}
	fmt.Println(fmt.Sprintf("factom config: %s", FactomConfigFilename))
	
	startServer := func(clone bool, number string) *FactomNode{
		newState := state
		if clone {
			newState = state.Clone(number).(*ss.State)
			newState.Init()
		} 
		
		states = append(states,newState)
		
		fnode := new(FactomNode)
		fnode.State = newState
		go NetworkProcessorNet(fnode)
		go loadDatabase(newState)
		go Timer(newState)
		go Validator(newState)
		return fnode
	}

	state.LoadConfig(FactomConfigFilename)
	state.Init()
//	fnode4 := startServer(true,"4")
//	fnode3 := startServer(true,"3")
//	fnode2 := startServer(true,"2")
	fnode1 := startServer(true,"1")
//	AddPeer(fnode2, fnode3)
//	AddPeer(fnode1, fnode2)
//	AddPeer(fnode2, fnode4)
	fnode0 := startServer(false,"0")
//	AddPeer(fnode0, fnode3)
	AddPeer(fnode0, fnode1)
	go wsapi.Start(fnode1.State)

	fnode1.State.SetOut(true)
	
	// Web API runs independent of Factom Servers

	for {
		time.Sleep(100000000)
	}
	
}
