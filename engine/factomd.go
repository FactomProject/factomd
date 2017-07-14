// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/state"
)

var _ = fmt.Print

// winServiceMain is only invoked on Windows.  It detects when btcd is running
// as a service and reacts accordingly.
//var winServiceMain func() (bool, error)

// Build sets the factomd build id using git's SHA
// $ go install -ldflags "-X github.com/FactomProject/factomd/engine.Build=`git rev-parse HEAD`"
// It also seems to need to have the previous binary deleted if recompiling to have this message show up if no code has changed.
// Since we are tracking code changes, then there is no need to delete the binary to use the latest message
var Build string

func Factomd(params *FactomParams, listenToStdin bool) interfaces.IState {
	log.Print("//////////////////////// Copyright 2017 Factom Foundation")
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
	state0.IsRunning = true
	state0.SetLeaderTimestamp(primitives.NewTimestampFromMilliseconds(0))
	fmt.Println("len(Args)", len(os.Args))

	go NetStart(state0, params, listenToStdin)
	return state0
}

func isCompilerVersionOK() bool {
	goodenough := false

	if strings.Contains(runtime.Version(), "1.5") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.6") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.7") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.8") {
		goodenough = true
	}
	return goodenough
}
