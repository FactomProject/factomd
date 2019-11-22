// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"reflect"
	"runtime"

	"github.com/FactomProject/factomd/engine"
)

//go:generate go run ./factomgenerate/generate.go

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Println("Command Line Arguments:")

	for _, v := range os.Args[1:] {
		fmt.Printf("\t%s\n", v)
	}

	params := engine.ParseCmdLine(os.Args[1:])

	fmt.Println("Parameter:")
	s := reflect.ValueOf(params).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fmt.Printf("%30s %s = %v\n", typeOfT.Field(i).Name, f.Type(), f.Interface())
	}

	fmt.Println()
	engine.Run(params)
}
