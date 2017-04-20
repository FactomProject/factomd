// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

func (state *State) SortOutAcks(msg interfaces.IMsg) {
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
	for i := 0; i < 20; i++ {
		p, b := state.Process(), state.UpdateState()
		if !p && !b {
			return
		}
	}
}

func (state *State) ValidatorLoop() {
	timeStruct := new(Timer)
	for {
		select {
		case <-state.ShutdownChan:
			fmt.Println("Closing the Database on", state.GetFactomNodeName())
			state.DB.Close()
			fmt.Println(state.GetFactomNodeName(), "closed")
			return
		case min := <-state.tickerQueue:
			timeStruct.timer(state, min)
		case msg := <-state.TimerMsgQueue():
			state.JournalMessage(msg)
			state.SortOutAcks(msg)
		case msg := <-state.InMsgQueue():
			// Get message from the timer or input queue
			state.JournalMessage(msg)
			state.SortOutAcks(msg)
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
