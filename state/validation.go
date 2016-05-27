// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

func (state *State) ValidatorLoop() {
	timeStruct := new(Timer)
	for {

		// Check if we should shut down.
		select {
		case _ = <-state.ShutdownChan:
			state.DBMutex.Lock()
			defer state.DBMutex.Unlock()
			fmt.Println("Closing the Database on", state.GetFactomNodeName())
			state.DB.Close()
			fmt.Println(state.GetFactomNodeName(), "closed")
			return
		default:
		}

		state.SetString() // Set the string for the state so we can print it later if we like.
		// Process any messages we might have queued up.
		for state.Process() {
			state.UpdateState()
		}

		// Look for pending messages, and get one if there is one.
		var msg interfaces.IMsg
	loop:
		for i := 0; i < 100; i++ {
			state.UpdateState()

			select {
			case min := <-state.tickerQueue:
				timeStruct.timer(state, min)
			default:
			}

			select {
			case msg = <-state.TimerMsgQueue():
				if state.Leader {
					state.JournalMessage(msg)
					break loop
				}
			default:
			}

			select {
			case msg = <-state.InMsgQueue(): // Get message from the timer or input queue
				state.JournalMessage(msg)
				break loop
			default: // No messages? Sleep for a bit.
				state.SetString()
				time.Sleep(10 * time.Millisecond)
			}
		}

		// Sort the messages.
		if msg != nil {
			if state.IsReplaying == true {
				state.ReplayTimestamp = msg.GetTimestamp()
			}
			if _, ok := msg.(*messages.Ack); ok {
				state.ackQueue <- msg
			} else {
				state.msgQueue <- msg
			}
		}
	}
}

type Timer struct {
	lastMin      int
	lastDBHeight uint32
}

func (t *Timer) timer(state *State, min int) {

	state.UpdateState()

	t.lastMin = min

	stateheight := state.LLeaderHeight

	if stateheight != t.lastDBHeight && min != 0 {
		return
	} else {
		t.lastDBHeight = stateheight
	}

	eom := new(messages.EOM)
	eom.Minute = byte(min)
	eom.Timestamp = state.GetTimestamp()
	eom.ChainID = state.GetIdentityChainID()
	eom.Sign(state)
	eom.SetLocal(true)
	state.TimerMsgQueue() <- eom
}
