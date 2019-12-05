// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/common/interfaces"
	llog "github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/pubsub"
	"github.com/FactomProject/factomd/util/atomic"
)

var ValidationDebug bool = false

// This is the tread with access to state. It does process and update state
func (s *State) MsgExecute() {
	s.validatorLoopThreadID = atomic.Goid()
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
		// If we haven't process inMessages in over a seconds go process them now
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

func (s *State) MsgSort() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("A panic state occurred in ValidatorLoop.", r)
			llog.LogPrintf("recovery", "A panic state occurred in ValidatorLoop. %v", r)
			shutdown(s)
		}
	}()
	leaderOut := pubsub.SubFactory.Channel(50)
	if s.GetFactomNodeName() == "FNode0" {
		leaderOut.Subscribe(pubsub.GetPath(s.GetFactomNodeName(), event.Path.LeaderMsgOut))
	}

	// Look for pending inMessages, and get one if there is one.
	for { // this is the message sort
		var msg interfaces.IMsg

		select {
		case <-s.ShutdownChan: // Check if we should shut down.
			shutdown(s)
			time.Sleep(10 * time.Second) // wait till database close is complete
			return
		case msg = <-s.inMsgQueue.Channel:
			s.LogMessage("InMsgQueue", "dequeue", msg)
			s.inMsgQueue.Metric(msg).Dec()
		case msg = <-s.inMsgQueue2.Channel:
			s.LogMessage("InMsgQueue2", "dequeue", msg)
			s.inMsgQueue2.Metric(msg).Dec()
		case v := <-leaderOut.Updates:
			msg = v.(interfaces.IMsg)
			s.LogMessage("leader", "out", msg)
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
	state.RunState = runstate.Stopping
	fmt.Println("Closing the Database on", state.GetFactomNodeName())
	state.StateSaverStruct.StopSaving()
	state.DB.Close()
	fmt.Println("Database on", state.GetFactomNodeName(), "closed")
	state.RunState = runstate.Stopped
}
