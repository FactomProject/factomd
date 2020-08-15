// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"time"

	"github.com/PaulSnow/factom2d/events/eventmessages/generated/eventmessages"

	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/constants/runstate"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/messages"
	"github.com/PaulSnow/factom2d/util/atomic"
)

var ValidationDebug bool = false

// This is the tread with access to state. It does process and update state
func (s *State) DoProcessing() {
	s.validatorLoopThreadID = atomic.Goid()

	s.EventService.EmitNodeInfoMessageF(eventmessages.NodeMessageCode_STARTED, "Node %s startup complete", s.GetFactomNodeName())
	s.RunState = runstate.Running

	slp := false
	i3 := 0

	for s.GetRunState() == runstate.Running {

		p1 := true
		p2 := true
		i1 := 0
		i2 := 0

		if ValidationDebug {
			s.LogPrintf("executeMsg", "start validate.process")
		}

		for i1 = 0; p1 && i1 < 20; i1++ {
			p1 = s.Process()
		}

		if ValidationDebug {
			s.LogPrintf("executeMsg", "start validate.updatestate")
		}
		for i2 = 0; p2 && i2 < 20; i2++ {
			p2 = s.UpdateState()
		}
		// Call process at least every second to insure MMR runs.
		now := s.GetTimestamp()
		p3 := false
		// If we haven't process messages in over a seconds go process them now
		if now.GetTimeMilli()-s.ProcessTime.GetTimeMilli() > int64(s.FactomSecond()/time.Millisecond) {
			for s.LeaderPL.Process(s) {
				p3 = true
			}
			s.ProcessTime = now
		}

		// if we were unable to accomplish any work sleep a bit.
		if !p1 && !p2 && !p3 {
			// No work? Sleep for a bit
			time.Sleep(10 * time.Millisecond)
			s.ValidatorLoopSleepCnt++
			i3++
			slp = true
		} else if slp {
			slp = false
			if ValidationDebug {
				s.LogPrintf("executeMsg", "DoProcessing() slept %d times", i3)
				i3 = 0
			}

		}

	}
	fmt.Println("Closing the Database on", s.GetFactomNodeName())
	s.DB.Close()
	s.StateSaverStruct.StopSaving()
	fmt.Println(s.GetFactomNodeName(), "closed")
}

func (s *State) ValidatorLoop() {
	defer func() {
		if r := recover(); r != nil {
			s.EventService.EmitNodeErrorMessage(eventmessages.NodeMessageCode_GENERAL,
				"A panic state occurred in ValidatorLoop.", r)

			shutdown(s)
		}
	}()

	CheckGrants()

	// We should only generate 1 EOM for each height/minute/vmindex
	lastHeight, lastMinute, lastVM := -1, -1, -1

	go s.DoProcessing()
	// Look for pending messages, and get one if there is one.
	for { // this is the message sort
		var msg interfaces.IMsg

		select {
		case <-s.ShutdownChan: // Check if we should shut down.
			shutdown(s)
			time.Sleep(10 * time.Second) // wait till database close is complete
			return
		case c := <-s.tickerQueue: // Look for pending messages, and get one if there is one.
			if !s.RunLeader || !s.DBFinished { // don't generate EOM if we are not ready to execute as a leader or are loading the DBState messages
				continue
			}
			currentMinute := s.CurrentMinute
			if currentMinute == 10 { // if we are between blocks
				currentMinute = 9 // treat minute 10 as an extension of minute 9
			}
			if lastHeight == int(s.LLeaderHeight) && lastMinute == currentMinute && s.LeaderVMIndex == lastVM {
				// This eom was already generated. We shouldn't generate it again.
				// This does mean we missed an EOM boundary, and the next EOM won't occur for another
				// "minute". This could cause some serious sliding, as minutes could be an addition 100%
				// in length.
				if c == -1 { // This means we received a normal eom cadence timer
					c = 8 // Send 8 retries on a 1/10 of the normal minute period
				}
				if c > 0 {
					go func() {
						// We sleep for 1/10 of a minute, and try again
						time.Sleep(s.GetMinuteDuration() / 10)
						s.tickerQueue <- c - 1
					}()
				}
				s.LogPrintf("timer", "retry %d", c)
				s.LogPrintf("validator", "retry %d  %d-:-%d %d", c, s.LLeaderHeight, currentMinute, s.LeaderVMIndex)
				continue // Already generated this eom
			}

			lastHeight, lastMinute, lastVM = int(s.LLeaderHeight), currentMinute, s.LeaderVMIndex

			eom := new(messages.EOM)
			eom.Timestamp = s.GetTimestamp()
			eom.ChainID = s.GetIdentityChainID()
			{
				// best guess info... may be wrong -- just for debug
				eom.DBHeight = s.LLeaderHeight
				eom.VMIndex = s.LeaderVMIndex
				eom.Minute = byte(currentMinute)
			}

			eom.Sign(s)
			eom.SetLocal(true) // local EOMs are really just timeout indicators that we need to generate an EOM
			msg = eom
			s.LogMessage("validator", fmt.Sprintf("generated c:%d  %d-:-%d %d", c, s.LLeaderHeight, s.CurrentMinute, s.LeaderVMIndex), eom)
		case msg = <-s.inMsgQueue:
			s.LogMessage("InMsgQueue", "dequeue", msg)
		case msg = <-s.inMsgQueue2:
			s.LogMessage("InMsgQueue2", "dequeue", msg)
		}

		if t := msg.Type(); t == constants.ACK_MSG {
			s.LogMessage("ackQueue", "enqueue ValidatorLoop", msg)
			s.ackQueue <- msg
		} else {
			s.LogMessage("msgQueue", "enqueue ValidatorLoop", msg)
			s.msgQueue <- msg
		}
	}
}

func shouldShutdown(state *State) bool {
	select {
	case <-state.ShutdownChan:
		shutdown(state)
		return true
	default:
		return false
	}
}

func shutdown(state *State) {
	state.EventService.EmitNodeInfoMessageF(eventmessages.NodeMessageCode_SHUTDOWN,
		"Node %s is shutting down", state.GetFactomNodeName())

	state.RunState = runstate.Stopping
	fmt.Println("Closing the Database on", state.GetFactomNodeName())
	state.StateSaverStruct.StopSaving()
	state.DB.Close()
	fmt.Println("Database on", state.GetFactomNodeName(), "closed")
	state.RunState = runstate.Stopped
}
