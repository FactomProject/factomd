// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/log"
)

var _ = fmt.Print
var _ = log.Printf

func Validator(state interfaces.IState) {

	for {
		msg := <-state.InMsgQueue() // Get message from the input queue

		if state.PrintType(msg.Type()) {
			fmt.Printf("%20s %s", "Validator:", msg.String())
		}

		switch msg.Validate(state.GetDBHeight(), state) { // Validate the message.
		case 1: // Process if valid
			state.NetworkOutMsgQueue() <- msg
			if state.PrintType(msg.Type()) {
				fmt.Printf(" Valid\n")
			}
			if msg.Leader(state) {
				if state.PrintType(msg.Type()) {
					fmt.Printf("%20s %s\n", "Leader:", msg.String())
				}
				msg.LeaderExecute(state)
				state.UpdateState()
			} else if msg.Follower(state) {
				if state.PrintType(msg.Type()) {
					fmt.Printf("%20s %s\n", "Follower:", msg.String())
				}
				msg.FollowerExecute(state)
				state.UpdateState()
			} else {
				fmt.Printf(" Message ignored\n")
			}
		case 0: // Hold for later if unknown.
			if state.PrintType(msg.Type()) {
				fmt.Printf(" Hold\n")
			}
		default:
			if state.PrintType(msg.Type()) {
				fmt.Printf(" Invalid\n")
			}
			state.NetworkInvalidMsgQueue() <- msg
		}
	}

}
