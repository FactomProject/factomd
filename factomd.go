// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"time"

	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/engine"
)

func main() {
	fmt.Println("Command Line Arguments:")

	for _, v := range os.Args[1:] {
		fmt.Printf("\t%s\n", v)
	}

	params := engine.ParseCmdLine(os.Args[1:])
	params.PrettyPrint()

	state := engine.Factomd(params)
	for state.GetRunState() != runstate.Stopped {
		time.Sleep(time.Second)
	}
	fmt.Println("Waiting to Shut Down") // This may not be necessary anymore with the new run state method
	time.Sleep(time.Second * 5)
}
