// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/btcd"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/wsapi"
	"time"
)

var _ = fmt.Print

func NetStart(state *state.State) {
	
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
	go wsapi.Start(state)
	go NetworkProcessorNet(state)
	
	// Web API runs independent of Factom Servers

	for {
		time.Sleep(100000000)
	}
	
}
