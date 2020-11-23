// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"reflect"
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

	state := engine.Factomd(params)
	for state.GetRunState() != runstate.Stopped {
		time.Sleep(time.Second)
	}
	fmt.Println("Waiting to Shut Down") // This may not be necessary anymore with the new run state method
	time.Sleep(time.Second * 5)
}
