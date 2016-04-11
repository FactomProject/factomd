// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"time"
)

func (state *State) ValidatorLoop() {
	for {
		state.SetString()
		select {
		case _ = <-state.ShutdownChan:
			fmt.Println("Closing the Database on", state.GetFactomNodeName())
			state.GetDB().(interfaces.IDatabase).Close()
			fmt.Println(state.GetFactomNodeName(), "closed")
			return
		default:
		}

		var msg interfaces.IMsg
	loop:
		for {
			state.UpdateState()
			state.UpdateState()
			state.UpdateState()
			state.UpdateState()
			msgProcess := func() {
				state.JournalMessage(msg)

				if state.PrintType(msg.Type()) {
					state.Println(fmt.Sprintf("%20s %s", "Validator:", msg.String()))
				}
			}
			select {
			case msg = <-state.TimerMsgQueue():
				msgProcess()
				break loop
			case msg = <-state.InMsgQueue(): // Get message from the timer or input queue
				msgProcess()
				break loop
			default:
				time.Sleep(time.Millisecond * 100)
			}
		}

		if state.IsReplaying == true {
			state.ReplayTimestamp = msg.GetTimestamp()
		}

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
			} else if msg.Follower(state) {
				if state.PrintType(msg.Type()) {
					state.Println(fmt.Sprintf("%20s %s\n", "Follower:", msg.String()))
				}
				msg.FollowerExecute(state)
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
