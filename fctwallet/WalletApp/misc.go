// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import(
    "os"
    "fmt"    
    "strconv"
)  

type Exit struct {
    ICommand
}

func (Exit) Execute (state State, args []string) error {
    if len(args)<2 { os.Exit(0) }
    if len(args) != 2 {
        return fmt.Errorf("Too many parameters")
    }
    n, err := strconv.ParseInt(args[1],10,32)
    if err != nil {
        return err
    }
    os.Exit(int(n))
    return nil
}
    
func (Exit) Name() string {
    return "Exit"
}

func (Exit) ShortHelp() string {
    return "Exit <error code> -- Exits Factom Wallet with the given error code"
}

func (Exit) LongHelp() string {
    return `
Exit <error code>                   Exits Factom Wallet with the given error code.
                                    <error code> is an int, and is OS specific
`
}

