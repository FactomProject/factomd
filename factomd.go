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

	//  Go Optimizations...
	runtime.GOMAXPROCS(runtime.NumCPU())

	state0 := new(state.State)
	
	fmt.Println("len(Args)",len(os.Args))
	if len(os.Args) == 1 {
		OneStart(state0)
	} else {
		NetStart(state0)
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
