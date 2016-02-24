// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/btcd"
	"github.com/FactomProject/factomd/btcd/limits"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/wsapi"
	"os"
)

var _ = fmt.Print

func OneStart(state *state.State) {
	
	state.SetOut(true)
	
	pcfg, _, err := btcd.LoadConfig()
	if err != nil {
		state.Println(err.Error())
	}
	
	if err := limits.SetLimits(); err != nil {
		os.Exit(1)
	}

	state.Print("Starting server\n")
	server, _ := btcd.NewServer(state)

	btcd.AddInterruptHandler(func() {
		state.Print("<Break>\n")
		state.Print("Gracefully shutting down the server...\n")
		state.GetDB().(interfaces.IDatabase).Close()
		server.Stop()
		server.WaitForShutdown()
	})
	server.Start()
	state.SetServer(server)
	
	FactomConfigFilename := pcfg.FactomConfigFile
	
	if len(FactomConfigFilename) == 0 {
		FactomConfigFilename = util.GetConfigFilename("m2")
	}
	state.Print(fmt.Sprintf("factom config: %s", FactomConfigFilename))
	
	//
	// Start Up Factom here!  
	//    Start Factom
	//    Add the API (don't have to)
	//    Add the network.  
	state.LoadConfig(FactomConfigFilename)
	FactomServerStart(state)
	go wsapi.Start(state)
	go NetworkProcessorOne(state)
	
	// Web API runs independent of Factom Servers

	shutdownChannel := make(chan struct{})
	go func() {
		server.WaitForShutdown()
		state.Print("Server shutdown complete\n")
		shutdownChannel <- struct{}{}
	}()

	// Wait for shutdown signal from either a graceful server stop or from
	// the interrupt handler.
	<-shutdownChannel
	state.Print("Shutdown complete\n")
}
