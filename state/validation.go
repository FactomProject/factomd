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

		state.SetString() // Set the string for the state so we can print it later if we like.

		// Process any messages we might have queued up.
		state.Process()

		// Check if we should shut down.
		select {
		case _ = <-state.ShutdownChan:
			fmt.Println("Closing the Database on", state.GetFactomNodeName())
			state.GetDB().(interfaces.IDatabase).Close()
			fmt.Println(state.GetFactomNodeName(), "closed")
			return
		default:
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
				state.JournalMessage(msg)
				break loop
			default:
			}

			select {
			case msg = <-state.InMsgQueue(): // Get message from the timer or input queue
				fmt.Println("VVVVVVVVVVVVVVVVVVVVV", msg.String())
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
			if msg.Leader(state) {
				state.LeaderMsgQueue() <- msg
			} else if msg.Follower(state) {
				state.FollowerMsgQueue() <- msg
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

	stateheight := state.GetLeaderHeight()

	if min == 0 {
		if t.lastDBHeight > 0 && t.lastDBHeight == stateheight {
			t.lastDBHeight = stateheight + 1
		} else {
			t.lastDBHeight = stateheight
		}
	}

	found, vmIndex := state.GetVirtualServers(t.lastDBHeight, min, state.GetIdentityChainID())
	if found {
		eom := new(messages.EOM)
		eom.Minute = byte(min)
		eom.Timestamp = state.GetTimestamp()
		eom.ChainID = state.GetIdentityChainID()
		eom.VMIndex = vmIndex
		eom.Sign(state)
		eom.DBHeight = t.lastDBHeight
		eom.SetLocal(true)
		if min == 9 {
			DBS := new(messages.DirectoryBlockSignature)
			DBS.ServerIdentityChainID = state.GetIdentityChainID()
			DBS.SetLocal(true)
			DBS.DBHeight = t.lastDBHeight
			DBS.VMIndex = vmIndex
			state.TimerMsgQueue() <- eom
			state.TimerMsgQueue() <- DBS
		} else {
			state.TimerMsgQueue() <- eom
		}
	}
}
