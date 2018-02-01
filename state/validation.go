// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	log "github.com/sirupsen/logrus"
)

func (state *State) ValidatorLoop() {
	timeStruct := new(Timer)
	for {
		// Check if we should shut down.
		select {
		case <-state.ShutdownChan:
			fmt.Println("Closing the Database on", state.GetFactomNodeName())
			state.DB.Close()
			state.StateSaverStruct.StopSaving()
			fmt.Println(state.GetFactomNodeName(), "closed")
			state.IsRunning = false
			return
		default:
		}

		// Look for pending messages, and get one if there is one.
		var msg interfaces.IMsg
	loop:
		for i := 0; i < 10; i++ {
			// Process any messages we might have queued up.
			for i = 0; i < 10; i++ {
				p, b := state.Process(), state.UpdateState()
				if !p && !b {
					break
				}
				//fmt.Printf("dddd %20s %10s --- %10s %10v %10s %10v\n", "Validation", state.FactomNodeName, "Process", p, "Update", b)
			}

			for i := 0; i < 10; i++ {
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

				msg = state.InMsgQueue().Dequeue()

				if msg != nil {
					state.JournalMessage(msg)
					break loop
				} else {
					// No messages? Sleep for a bit
					for i := 0; i < 10 && state.InMsgQueue().Length() == 0; i++ {
						time.Sleep(10 * time.Millisecond)
					}
				}
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
	t.lastMin = min

	eom := new(messages.EOM)
	eom.Timestamp = state.GetTimestamp()
	eom.ChainID = state.GetIdentityChainID()
	eom.Sign(state)
	eom.SetLocal(true)
	consenLogger.WithFields(log.Fields{"func": "GenerateEOM", "lheight": state.GetLeaderHeight()}).WithFields(eom.LogFields()).Debug("Generate EOM")

	state.TimerMsgQueue() <- eom
}
