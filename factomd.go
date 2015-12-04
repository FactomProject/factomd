// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/btcd"
	"github.com/FactomProject/factomd/log"
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

func main() {
	log.Print("//////////////////////// Copyright 2015 Factom Foundation")
	log.Print("//////////////////////// Use of this source code is governed by the MIT")
	log.Print("//////////////////////// license that can be found in the LICENSE file.")

	log.Printf("Go compiler version: %s\n", runtime.Version())

	if !isCompilerVersionOK() {
		for i := 0; i < 30; i++ {
			fmt.Println("!!! !!! !!! ERROR: unsupported compiler version !!! !!! !!!")
		}
		time.Sleep(3 * time.Second)
		os.Exit(1)
	}
	cfgFilename := ""
	if len(os.Args) > 1 {
		cfgFilename = os.Args[1]
	}

	state := new(state.State)
	state.Init(cfgFilename)

	//Todo: gracefully shutting down the database here

	server, err := btcd.NewServer(state)
	if err != nil {
		log.Errorf("Unable to start server on %v: %v",	cfg.Listeners, err)
		return err
	}
	btcd.AddInterruptHandler(func() {
		log.Infof("Gracefully shutting down the server...")
		server.Stop()
		server.WaitForShutdown()
	})
	server.Start()

	//factomForkInit(server)
	go NetworkProcessor(state)
	go Timer(state)
	go Validator(state)
	go Leader(state)
	go Follower(state)
	go wsapi.Start(state)

	shutdownChannel = make(chan struct{})
	go func() {
		server.WaitForShutdown()
		log.Infof("Server shutdown complete")
		shutdownChannel <- struct{}{}
	}()

	// Wait for shutdown signal from either a graceful server stop or from
	// the interrupt handler.
	<-shutdownChannel
	log.Info("Shutdown complete")
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
