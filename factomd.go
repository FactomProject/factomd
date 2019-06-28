// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"reflect"
	"runtime"
	"time"

	"github.com/FactomProject/factomd/common/constants/runstate"
	. "github.com/FactomProject/factomd/engine"
)

func main() {
	// uncomment StartProfiler() to run the pprof tool (for testing)

	//  Go Optimizations...
	runtime.GOMAXPROCS(runtime.NumCPU()) // TODO: should be *2 to use hyperthreadding? -- clay

	fmt.Println("Command Line Arguments:")

	for _, v := range os.Args[1:] {
		fmt.Printf("\t%s\n", v)
	}

	params := ParseCmdLine(os.Args[1:])
	fmt.Println()

	fmt.Println("Parameter:")
	s := reflect.ValueOf(params).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fmt.Printf("%d: %25s %s = %v\n", i,
			typeOfT.Field(i).Name, f.Type(), f.Interface())
	}

	fmt.Println()
	sim_Stdin := params.Sim_Stdin

	state := Factomd(params, sim_Stdin)
	for state.GetRunState() != runstate.Stopped {
		time.Sleep(time.Second)
	}
	fmt.Println("Waiting to Shut Down") // This may not be necessary anymore with the new run state method
	time.Sleep(time.Second * 5)
}
