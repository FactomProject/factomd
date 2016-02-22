// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
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
	BroadcastOut      chan interfaces.IMsg
	BroadcastIn       chan interfaces.IMsg
	PrivateOut        chan interfaces.IMsg
	PrivateIn         chan interfaces.IMsg	
}

func (f *FactomPeer) init() *FactomPeer {
	f.BroadcastOut = make(chan interfaces.IMsg,10000)
	f.BroadcastIn  = make(chan interfaces.IMsg,10000)
	f.PrivateOut = make(chan interfaces.IMsg,10000)
	f.PrivateIn  = make(chan interfaces.IMsg,10000)
	return f
}

func AddPeer(f1, f2 *FactomNode) {
	peer12 := new(FactomPeer).init()
	peer21 := new(FactomPeer).init()
	peer12.BroadcastOut = peer21.BroadcastIn
	peer12.BroadcastIn = peer21.BroadcastOut
	peer12.PrivateOut = peer21.PrivateIn
	peer12.PrivateIn = peer21.PrivateOut
	peer21.BroadcastOut = peer12.BroadcastIn
	peer21.BroadcastIn = peer12.BroadcastOut
	peer21.PrivateOut = peer12.PrivateIn
	peer21.PrivateIn = peer12.PrivateOut
	
	f1.Peers = append(f1.Peers,peer12)
	f2.Peers = append(f2.Peers,peer21)
}

func NetStart(state *ss.State) {
	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Net Sim Start!!!!!")
	fmt.Println(">>>>>>>>>>>>>>>>")
	
	btcd.AddInterruptHandler(func() {
		log.Printf("<Break>\n")
		log.Printf("Gracefully shutting down the server...\n")
		state.ShutdownChan <- 0
	})
		
	
	pcfg, _, err := btcd.LoadConfig()
	if err != nil {
		log.Println(err.Error())
	}
	FactomConfigFilename := pcfg.FactomConfigFile
	
	if len(FactomConfigFilename) == 0 {
		FactomConfigFilename = util.GetConfigFilename("m2")
	}
	log.Printfln("factom config: %s", FactomConfigFilename)
	//
	// Start Up Factom here!  
	//    Start Factom
	//    Add the API (don't have to)
	//    Add the network.  
	state.LoadConfig(FactomConfigFilename)

	state.Init()
	
	state2 := state.Clone("1").(*ss.State)
	state2.Init()

	fnode1 := new(FactomNode)
	fnode1.State = state
	fnode2 := new(FactomNode)
	fnode2.State = state2
	
	go NetworkProcessorNet(fnode2)
	go loadDatabase(state2)
	go Timer(state2)
	go Validator(state2)
	
	AddPeer(fnode1, fnode2)
	
	go NetworkProcessorNet(fnode1)
	go loadDatabase(state)
	go Timer(state)
	go Validator(state)
	
	
	go wsapi.Start(state)
	
	// Web API runs independent of Factom Servers

	for {
		time.Sleep(100000000)
	}
	
}
