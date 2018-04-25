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
	var prev time.Time
	for {
		if state.debugExec() {
			status := ""
			now := time.Now()
			if now.Sub(prev).Minutes() > 1 {
				state.LogPrintf("executeMsg", "Timestamp DBh/VMh/h %d/%d/%d", state.LLeaderHeight, state.LeaderVMIndex, state.CurrentMinute)
				pendingEBs := 0
				pendingEntries := 0
				pl := state.ProcessLists.Get(state.LLeaderHeight)
				if pl != nil {
					pendingEBs = len(pl.NewEBlocks)
					pendingEntries = len(pl.NewEntries)
				}
				status += fmt.Sprintf("Review %d ", len(state.XReview))
				status += fmt.Sprintf("Holding %d ", len(state.Holding))
				status += fmt.Sprintf("Commits %d ", state.Commits.Len())
				status += fmt.Sprintf("Pending EBs %d ", pendingEBs)         // cope with nil
				status += fmt.Sprintf("Pending Entries %d ", pendingEntries) // cope with nil
				status += fmt.Sprintf("Acks %d ", len(state.AcksMap))
				status += fmt.Sprintf("MsgQueue %d ", len(state.msgQueue))
				status += fmt.Sprintf("InMsgQueue %d ", state.inMsgQueue.Length())
				status += fmt.Sprintf("APIQueue   %d ", state.apiQueue.Length())
				status += fmt.Sprintf("AckQueue %d ", len(state.ackQueue))
				status += fmt.Sprintf("TimerMsgQueue %d ", len(state.timerMsgQueue))
				status += fmt.Sprintf("NetworkOutMsgQueue %d ", state.networkOutMsgQueue.Length())
				status += fmt.Sprintf("NetworkInvalidMsgQueue %d ", len(state.networkInvalidMsgQueue))
				status += fmt.Sprintf("UpdateEntryHash %d ", len(state.UpdateEntryHash))
				status += fmt.Sprintf("MissingEntries %d ", state.GetMissingEntryCount())
				status += fmt.Sprintf("WriteEntry %d ", len(state.WriteEntry))

				state.LogPrintf("executeMsg", "Status %s", status)
				prev = now
			}
		}
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
			for state.Process() {
			}
			for state.UpdateState() {
			}

			select {
			case min := <-state.tickerQueue:
				timeStruct.timer(state, min)
			default:
			}

			for i := 0; i < 1000; i++ {
				ackRoom := cap(state.ackQueue) - len(state.ackQueue)
				msgRoom := cap(state.msgQueue) - len(state.msgQueue)

				if ackRoom > 1 && msgRoom > 1 {
					msg = state.InMsgQueue().Dequeue()
				}
				// This doesn't block so it intentionally returns nil, don't log nils
				if msg != nil {
					state.LogMessage("InMsgQueue", "dequeue", msg)
				}

				if msg != nil {
					state.JournalMessage(msg)
					break loop
				} else {
					// No messages? Sleep for a bit
					for i := 0; i < 10 && state.InMsgQueue().Length() == 0; i++ {
						time.Sleep(10 * time.Millisecond)
					}
					break
				}
			}
		}

		// Sort the messages.
		if msg != nil {
			if state.IsReplaying == true {
				state.ReplayTimestamp = msg.GetTimestamp()
			}
			if _, ok := msg.(*messages.Ack); ok {
				state.LogMessage("ackQueue", "enqueue", msg)
				state.ackQueue <- msg //
			} else {
				state.LogMessage("msgQueue", "enqueue", msg)
				state.msgQueue <- msg //
			}
		}
	}
}

type Timer struct {
	lastMin      int
	lastDBHeight uint32
}

func (t *Timer) timer(s *State, min int) {
	t.lastMin = min

	if s.RunLeader { // don't generate EOM if we are not a leader or are loading the DBState messages
	eom := new(messages.EOM)
		eom.Timestamp = s.GetTimestamp()
		eom.ChainID = s.GetIdentityChainID()
		{
			// best guess info... may be wrong -- just for debug
			eom.DBHeight = s.LLeaderHeight
			eom.VMIndex = s.LeaderVMIndex
			// EOM.Minute is zerobased, while LeaderMinute is 1 based.  So
			// a simple assignment works.
			eom.Minute = byte(s.CurrentMinute)
		}

		eom.Sign(s)
	eom.SetLocal(true)
		consenLogger.WithFields(log.Fields{"func": "GenerateEOM", "lheight": s.GetLeaderHeight()}).WithFields(eom.LogFields()).Debug("Generate EOM")

		s.MsgQueue() <- eom
	}
}
