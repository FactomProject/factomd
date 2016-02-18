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
	"runtime"
	"strings"
	"time"
)

var _ = fmt.Print

// winServiceMain is only invoked on Windows.  It detects when btcd is running
// as a service and reacts accordingly.
//var winServiceMain func() (bool, error)

// Build sets the factomd build id using git's SHA
// by compiling factomd with: -ldflags "-X main.Build=<build sha1>"
// e.g. get  the short version of the sha1 of your latest commit by running
// $ git rev-parse --short HEAD
// $ go install -ldflags "-X main.Build=6c10244"
var Build string

func main() {

//	go StartProfiler()

	log.Print("//////////////////////// Copyright 2015 Factom Foundation")
	log.Print("//////////////////////// Use of this source code is governed by the MIT")
	log.Print("//////////////////////// license that can be found in the LICENSE file.")
	log.Printf("Go compiler version: %s\n", runtime.Version())
	log.Printf("Using build: %s\n", Build)

	if !isCompilerVersionOK() {
		for i := 0; i < 30; i++ {
			log.Println("!!! !!! !!! ERROR: unsupported compiler version !!! !!! !!!")
		}
		time.Sleep(3 * time.Second)
		os.Exit(1)
	}

	pcfg, _, err := btcd.LoadConfig()
	if err != nil {
		log.Println(err.Error())
	}
	sim := pcfg.SimNet
	fcfg := pcfg.FactomConfigFile
	if sim {
		log.Println(" ==========> Simulation Network <===========")
	}

	//var state0, state1, state2, state3, state4 *state.State
	state0 := new(state.State)
	if len(fcfg) > 0 {
		log.Printfln("factom config: %s", fcfg)
		state0.Init(fcfg)
	} else {
		state0.Init(util.GetConfigFilename("m2"))
	} /*
		if sim {
			state1 := new(state.State)
			state1.Init(util.GetConfigFilename("m2-1"))
			state2 := new(state.State)
			state2.Init(util.GetConfigFilename("m2-2"))
			state3 := new(state.State)
			state3.Init(util.GetConfigFilename("m2-3"))
			state4 := new(state.State)
			state4.Init(util.GetConfigFilename("m2-4"))
		}*/

	btcd.AddInterruptHandler(func() {
		log.Printf("Gracefully shutting down the database...")
		state0.GetDB().(interfaces.IDatabase).Close()
		if sim {
			//state1.GetDB().(interfaces.IDatabase).Close()
			//state2.GetDB().(interfaces.IDatabase).Close()
			//state3.GetDB().(interfaces.IDatabase).Close()
			//state4.GetDB().(interfaces.IDatabase).Close()
		}
	})

	// Go Optimizations...
	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := limits.SetLimits(); err != nil {
		os.Exit(1)
	}

	log.Print("Starting server")
	server, _ := btcd.NewServer(state0)

	btcd.AddInterruptHandler(func() {
		log.Printf("Gracefully shutting down the server...")
		server.Stop()
		server.WaitForShutdown()
	})
	server.Start()
	state0.SetServer(server)

	//if sim {
	//go SimNetwork(state0, state1, state2, state3, state4)
	//}else{
	go NetworkProcessor(state0)
	//}

	go Timer(state0)
	go Validator(state0)
	go Leader(state0)
	go Follower(state0)

	go wsapi.Start(state0)

	shutdownChannel := make(chan struct{})
	go func() {
		server.WaitForShutdown()
		log.Printf("Server shutdown complete")
		shutdownChannel <- struct{}{}
	}()

	// Wait for shutdown signal from either a graceful server stop or from
	// the interrupt handler.
	<-shutdownChannel
	log.Printf("Shutdown complete")
}

func isCompilerVersionOK() bool {
	goodenough := false

	if strings.Contains(runtime.Version(), "1.4") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.5") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.6") {
		goodenough = true
	}

	return goodenough
}
