// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
    "os"
	"fmt"
    "time"
    "bufio"
    "strings"
    fct "github.com/FactomProject/factoid"
   // "golang.org/x/crypto/ssh/terminal"
)
var _ = fmt.Println
var _ fct.Transaction
var _ = time.Now 
 
func main() {
    state := NewState()
    r := bufio.NewReader(os.Stdin)
    for {
        fmt.Print(" Factom Wallet$ ")
        line,_,_ := r.ReadLine()
        args := strings.Fields(string(line))    
        err := state.Execute(args)
        if err != nil {
            fmt.Println(err)
            c,_ := state.GetCommand(args)
            if c != nil {
                fmt.Println(c.ShortHelp())
            }
        }
    }
}



