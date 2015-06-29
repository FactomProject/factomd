// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import(
    "fmt"    
    "strings"
)  

type Help struct {
    ICommand
}

func (h Help) Execute (state State, args []string) error {
    if len(args)==1 || len(args)>2 { 
        fmt.Println(h.ShortHelp())
        return nil
    }
    if strings.ToLower(args[1]) == "all" {
        keys := make([]string,0,len(state.commands))
        for key, _ := range(state.commands) {
            keys = append(keys,key)
        }
        for i:=0;i<len(keys)-1; i++ {
            for j:=0;j<len(keys)-1-i;j++ {
                if keys[j]>keys[j+1] {
                    t := keys[j]
                    keys[j]=keys[j+1]
                    keys[j+1]=t
                }
            }
        }
        for _,key := range(keys) {
            fmt.Println(state.commands[key].LongHelp())
        }
    }
    c := state.commands[strings.ToLower(args[1])]
    if c != nil {
        fmt.Println(c.LongHelp())
    }else{
        return fmt.Errorf("Unknown Command")
    }
    return nil
}
    
func (Help) Name() string {
    return "Help"
}

func (Help) ShortHelp() string {
    return "Help <all|cmd> -- Help all prints help on all commands.\n"+  
           "                  Help <cmd> prints the detailed help on the given command."
}

func (Help) LongHelp() string {
    return `
Help <all|cmd>                      Help facility for the Factom Wallet App.  Returns 
                                    help on all commands, or on individual specified
                                    commands.
`
}

