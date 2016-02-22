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
	State	*ss.State
	Peers	[]*FactomPeer
}

type FactomPeer struct {	
	BroadcastOut      chan interfaces.IMsg
	BroadcastIn       chan interfaces.IMsg
	PrivateOut        chan interfaces.IMsg
	PrivateIn         chan interfaces.IMsg	
}

func (f *FactomPeer) init() *FactomPeer {
	f.BroadcastOut = make(chan interfaces.IMsg,10)
	f.BroadcastIn  = make(chan interfaces.IMsg,10)
	f.PrivateOut = make(chan interfaces.IMsg,10)
	f.PrivateIn  = make(chan interfaces.IMsg,10)
	return f
}

func AddPeer(f1, f2 FactomNode) {
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

	FactomServerStart(state)
	
	state1 := state.Clone("1").(*ss.State)
	FactomServerStart(state1)
	
	go wsapi.Start(state)
	go NetworkProcessorNet(state)
	
	// Web API runs independent of Factom Servers

	for {
		time.Sleep(100000000)
	}
	
}
