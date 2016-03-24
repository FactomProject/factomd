// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/log"
	ss "github.com/FactomProject/factomd/state"
	"os"
)

var _ = fmt.Print
var _ = log.Printf
var _ = os.Exit

func Validator(state interfaces.IState) {
	s := state.(*ss.State)
	var _ = s

	for {
		state.SetString()
		select {
		case _ = <-s.ShutdownChan:
			fmt.Println("Closing the Database on", state.GetFactomNodeName())
			state.GetDB().(interfaces.IDatabase).Close()
			fmt.Println(state.GetFactomNodeName(), "closed")
			return
		default:
		}

		msg := <-state.InMsgQueue() // Get message from the input queue

		if state.PrintType(msg.Type()) {
			state.Println(fmt.Sprintf("%20s %s", "Validator:", msg.String()))
		}

		// TODO:  Height here is problematic.  We ensure that the leader doesn't step
		// past the follower...
		switch msg.Validate(state) { // Validate the message.
		case 1: // Process if valid

			if !msg.IsPeer2peer() {
				state.NetworkOutMsgQueue() <- msg
			}

			if state.PrintType(msg.Type()) {
				state.Print(" Valid\n")
			}
			if msg.Leader(state) {
				if state.PrintType(msg.Type()) {
					state.Println(fmt.Sprintf("%20s %s\n", "Leader:", msg.String()))
				}
				msg.LeaderExecute(state)
				state.UpdateState()
			} else if msg.Follower(state) {
				if state.PrintType(msg.Type()) {
					state.Println(fmt.Sprintf("%20s %s\n", "Follower:", msg.String()))
				}
				msg.FollowerExecute(state)
				state.UpdateState()
			} else {
				state.Print(" Message ignored\n")
			}
		case 0: // Hold for later if unknown.
			if state.PrintType(msg.Type()) {
				state.Print(" Hold\n")
			}
		default:
			if state.PrintType(msg.Type()) {
				state.Print(" Invalid\n")
			}
			state.NetworkInvalidMsgQueue() <- msg
		}
	}

}
