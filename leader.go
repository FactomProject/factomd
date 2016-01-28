// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/log"
)

var _ = log.Print
var _ = fmt.Print

func Leader(state interfaces.IState) {

	for {
		msg := <-state.LeaderInMsgQueue()
		if state.PrintType(msg.Type()) {
			fmt.Printf("%20s %s\n", "Leader:", msg.String())
		}
		msg.LeaderExecute(state)
	}

}
