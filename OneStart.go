// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/btcd"
	"github.com/FactomProject/factomd/btcd/limits"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/wsapi"
	"os"
)

var _ = fmt.Print

func OneStart(state *state.State) {

	pcfg, _, err := btcd.LoadConfig()
	if err != nil {
		log.Println(err.Error())
	}
	fcfg := pcfg.FactomConfigFile

	if len(fcfg) > 0 {
		log.Printfln("factom config: %s", fcfg)
		state.Init(fcfg)
	} else {
		state.Init(util.GetConfigFilename("m2"))
	}

	btcd.AddInterruptHandler(func() {
		log.Printf("Gracefully shutting down the database...\n")
		state.GetDB().(interfaces.IDatabase).Close()
	})

	if err := limits.SetLimits(); err != nil {
		os.Exit(1)
	}

	log.Print("Starting server\n")
	server, _ := btcd.NewServer(state)

	btcd.AddInterruptHandler(func() {
		log.Printf("Gracefully shutting down the server...\n")
		server.Stop()
		server.WaitForShutdown()
	})
	server.Start()
	state.SetServer(server)

	// Network runs independent of Factom
	go NetworkProcessor(state)
	// Timer runs periodically, and inserts eom messages into the stream
	go Timer(state)
	// Validator is the gateway.
	go Validator(state)
	// Web API runs independent of Factom
	go wsapi.Start(state)

	shutdownChannel := make(chan struct{})
	go func() {
		server.WaitForShutdown()
		log.Printf("Server shutdown complete\n")
		shutdownChannel <- struct{}{}
	}()

	// Wait for shutdown signal from either a graceful server stop or from
	// the interrupt handler.
	<-shutdownChannel
	log.Printf("Shutdown complete\n")
}
