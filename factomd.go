// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/btcd"
	"github.com/FactomProject/factomd/btcd/limits"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/wsapi"
	"os"
	"runtime"
	"strings"
	"time"
)

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
	log.Print("//////////////////////// Copyright 2015 Factom Foundation")
	log.Print("//////////////////////// Use of this source code is governed by the MIT")
	log.Print("//////////////////////// license that can be found in the LICENSE file.")

	log.Printf("Go compiler version: %s\n", runtime.Version())
	log.Printf("Using build: %s\n", Build)

	if !isCompilerVersionOK() {
		for i := 0; i < 30; i++ {
			fmt.Println("!!! !!! !!! ERROR: unsupported compiler version !!! !!! !!!")
		}
		time.Sleep(3 * time.Second)
		os.Exit(1)
	}
	cfgFilename := ""

	state := new(state.State)
	state.Init(cfgFilename)

	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := limits.SetLimits(); err != nil {
		os.Exit(1)
	}

	btcd.AddInterruptHandler(func() {
		log.Printf("Gracefully shutting down the database...")
		state.GetDB().(interfaces.IDatabase).Close()
	})

	log.Print("Starting server")
	server, _ := btcd.NewServer(state)

	btcd.AddInterruptHandler(func() {
		log.Printf("Gracefully shutting down the server...")
		server.Stop()
		server.WaitForShutdown()
	})
	server.Start()
	state.SetServer(server)

	//factomForkInit(server)
	go NetworkProcessor(state)
	go Timer(state)
	go Validator(state)
	go Leader(state)
	go Follower(state)
	go wsapi.Start(state)

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
