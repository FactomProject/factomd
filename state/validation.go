// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"time"
)

func (state *State) ValidatorLoop() {
	timeStruct := new(Timer)

	var msg interfaces.IMsg

	// Sort the messages.
	sortMsg := func() {
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

	checkCtrlC := func() {
		// Check if we should shut down.
		select {
		case <-state.ShutdownChan:
			fmt.Println("Closing the Database on", state.GetFactomNodeName())
			state.DB.Close()
			fmt.Println(state.GetFactomNodeName(), "closed")
			return
		default:
		}
	}

	for {

		checkCtrlC()
		// Process any messages we might have queued up.
		for i := 0; i < 5; i++ {
			p, b := state.Process(), state.UpdateState()
			if !p && !b {
				break
			}
			checkCtrlC()
		}

		select {
		case min := <-state.tickerQueue:
			timeStruct.timer(state, min)
		default:
		}

		select {
		case msg = <-state.TimerMsgQueue():
			state.JournalMessage(msg)
			sortMsg()
		default:
		}

		select {
		case msg = <-state.InMsgQueue(): // Get message from the timer or input queue
			state.JournalMessage(msg)
			sortMsg()
		default: // No messages? Sleep for a bit
			time.Sleep(10 * time.Millisecond)
		}
	}
}

type Timer struct {
	lastMin      int
	lastDBHeight uint32
}

func (t *Timer) timer(state *State, min int) {

	t.lastMin = min

	eom := new(messages.EOM)
	eom.Timestamp = state.GetTimestamp()
	eom.ChainID = state.GetIdentityChainID()
	eom.Sign(state)
	eom.SetLocal(true)
	state.TimerMsgQueue() <- eom
}
