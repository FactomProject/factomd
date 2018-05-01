// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"fmt"
	"runtime"
	"time"

	. "github.com/FactomProject/factomd/engine"
)

func main() {
	// uncomment StartProfiler() to run the pprof tool (for testing)
	params := ParseCmdLine(os.Args[1:])

	//  Go Optimizations...
	runtime.GOMAXPROCS(runtime.NumCPU()) // TODO: should be *2 to use hyperthreadding? -- clay

	fmt.Printf("Arguments\n %+v\n", params)

	sim_Stdin := params.Sim_Stdin

	state := Factomd(params, sim_Stdin)
	for state.Running() {
		time.Sleep(time.Second)
	}
	fmt.Println("Waiting to Shut Down")
	time.Sleep(time.Second * 5)
}
