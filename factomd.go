// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/state"
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
	if len(os.Args) > 0 {
		cfgFilename = os.Args[0]
	}

	state := new(state.State)
	state.Init(cfgFilename)

	go NetworkProcessor(state)
	go Timer(state)
	go Validator(state)
	go Leader(state)
	go Follower(state)

	for {
		time.Sleep(time.Duration(5) * time.Second)
	}

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
