// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"fmt"
	"github.com/FactomProject/factomd/engine"
	"time"
)

func main() {
	// uncomment StartProfiler() to run the pprof tool (for testing)
	params := engine.ParseCmdLine(os.Args[1:])
	state := engine.Factomd(params, true)
	for state.Running() {
		time.Sleep(time.Second)
	}
	fmt.Println("Waiting to Shut Down")
	time.Sleep(time.Second * 5)
}
