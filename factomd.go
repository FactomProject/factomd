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
	log "github.com/sirupsen/logrus"

	"strings"
)

func main() {
	// uncomment StartProfiler() to run the pprof tool (for testing)
	params := ParseCmdLine(os.Args[1:])

	if params.StdoutLog != "" || params.StderrLog != "" {
		HandleLogfiles(params.StdoutLog, params.StderrLog)
	}

	log.Print("//////////////////////// Copyright 2017 Factom Foundation")
	log.Print("//////////////////////// Use of this source code is governed by the MIT")
	log.Print("//////////////////////// license that can be found in the LICENSE file.")

	if !isCompilerVersionOK() {
		log.Println("!!! !!! !!! ERROR: unsupported compiler version !!! !!! !!!")
		time.Sleep(3 * time.Second)
		os.Exit(1)
	}

	// launch debug console if requested
	if params.DebugConsole != "" {
		LaunchDebugServer(params.DebugConsole)
	}

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

func isCompilerVersionOK() bool {
	goodenough := false

	if strings.Contains(runtime.Version(), "1.6") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.7") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.8") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.9") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.10") {
		goodenough = true
	}

	return goodenough
}
