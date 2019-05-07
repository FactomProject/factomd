// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/util/atomic"
	log "github.com/sirupsen/logrus"
)

var ValidationDebug bool = true

func (state *State) ValidatorLoop() {
	CheckGrants()
	timeStruct := new(Timer)
	var prev time.Time
	state.validatorLoopThreadID = atomic.Goid()
	for {
		s := state
		if state.DebugExec() {
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
				status += fmt.Sprintf("InMsgQueue2 %d ", state.inMsgQueue2.Length())
				status += fmt.Sprintf("APIQueue   %d ", state.apiQueue.Length())
				status += fmt.Sprintf("AckQueue %d ", len(state.ackQueue))
				status += fmt.Sprintf("TimerMsgQueue %d ", len(state.timerMsgQueue))
				status += fmt.Sprintf("NetworkOutMsgQueue %d ", state.networkOutMsgQueue.Length())
				status += fmt.Sprintf("NetworkInvalidMsgQueue %d ", len(state.networkInvalidMsgQueue))
				status += fmt.Sprintf("UpdateEntryHash %d ", len(state.UpdateEntryHash))
				status += fmt.Sprintf("MissingEntries %d ", state.GetMissingEntryCount())
				status += fmt.Sprintf("WriteEntry %d ", len(state.WriteEntry))

				status += fmt.Sprintf("PL %8d, P %8d, U %8d, S%8d", s.ProcessListProcessCnt, s.StateProcessCnt, s.StateUpdateState, s.ValidatorLoopSleepCnt)
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

		for i := 0; i < 1; i++ {
			//for state.Process() {}
			//for state.UpdateState() {}
			p1 := true
			p2 := true
			i1 := 0
			i2 := 0

			if ValidationDebug {
				s.LogPrintf("executeMsg", "start validate.process")
			}
			for i1 = 0; p1 && i1 < 20; i1++ {
				p1 = state.Process()
			}
			if ValidationDebug {
				s.LogPrintf("executeMsg", "start validate.updatestate")
			}
			for i2 = 0; p2 && i2 < 20; i2++ {
				p2 = state.UpdateState()
			}
			if ValidationDebug {
				s.LogPrintf("executeMsg", "start validate.messagesort")
			}

			if ValidationDebug {
				s.LogPrintf("updateIssues", "Validation messages %3d processlist %3d", i1, i2)
			}

			select {
			case min := <-state.tickerQueue:
				timeStruct.timer(state, min)
			default:
			}

			msgcnt := 0
			for i := 0; i < 50; i++ {
				ackRoom := cap(state.ackQueue) - len(state.ackQueue)
				msgRoom := cap(state.msgQueue) - len(state.msgQueue)

				if ackRoom < 1 || msgRoom < 1 {
					break // no room
				}
				msg = nil // use queue2 every 5th message
				if i%5 != 0 {
					// This doesn't block so it intentionally returns nil, don't log nils
					msg = state.InMsgQueue().Dequeue()
					if msg != nil {
						state.LogMessage("InMsgQueue", "dequeue", msg)
					}
				}
				if msg == nil {
					// This doesn't block so it intentionally returns nil, don't log nils
					msg = state.InMsgQueue2().Dequeue()
					if msg != nil {
						state.LogMessage("InMsgQueue2", "dequeue", msg)
					}
				}

				// This doesn't block so it intentionally returns nil, don't log nils

				if msg != nil {
					msgcnt++
					// Sort the messages.
					if t := msg.Type(); t == constants.ACK_MSG {
						state.LogMessage("ackQueue", "enqueue", msg)
						state.ackQueue <- msg //
					} else {
						state.LogMessage("msgQueue", "enqueue", msg)
						state.msgQueue <- msg //
					}
				}
			}
			if ValidationDebug {
				s.LogPrintf("executeMsg", "stop validate.messagesort sorted %d messages", msgcnt)
			}

			// if we are not making progress and there are no messages to sort  sleep a bit
			if !(p1 || p2) && state.InMsgQueue().Length() == 0 && state.InMsgQueue2().Length() == 0 {
				// No messages? Sleep for a bit
				for i := 0; i < 10 && state.InMsgQueue().Length() == 0 && state.InMsgQueue2().Length() == 0; i++ {
					time.Sleep(10 * time.Millisecond)
					state.ValidatorLoopSleepCnt++
				}

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

	if s.RunLeader && s.DBFinished { // don't generate EOM if we are not a leader or are loading the DBState messages
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
		s.LogMessage("MsgQueue", "enqueue", eom)

		s.MsgQueue() <- eom
	}
}
