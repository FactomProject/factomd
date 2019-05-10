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
)

var ValidationDebug bool = false

// This is the tread with access to state. It does process and update state
func (s *State) DoProcessing() {
	s.validatorLoopThreadID = atomic.Goid()
	s.IsRunning = true

	slp := false
	i3 := 0

	for s.IsRunning {

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
		if !p1 && !p2 {
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
	CheckGrants()

	go s.DoProcessing()

	// Look for pending messages, and get one if there is one.
	for { // this is the message sort
		var msg interfaces.IMsg
		select {
		case <-s.ShutdownChan: // Check if we should shut down.
			s.IsRunning = false
			time.Sleep(10 * time.Second) // wait till database close is complete
			return
		case <-s.tickerQueue: // Look for pending messages, and get one if there is one.
			if !s.RunLeader || !s.DBFinished { // don't generate EOM if we are not a leader or are loading the DBState messages
				continue
			}

			eom := new(messages.EOM)
			eom.Timestamp = s.GetTimestamp()
			eom.ChainID = s.GetIdentityChainID()
			{
				// best guess info... may be wrong -- just for debug
				eom.DBHeight = s.LLeaderHeight
				eom.VMIndex = s.LeaderVMIndex
				eom.Minute = byte(s.CurrentMinute)
			}

			eom.Sign(s)
			eom.SetLocal(true) // local EOMs are really just timeout indicators that we need to generate an EOM
			msg = eom
		case msg = <-s.inMsgQueue:
			s.LogMessage("InMsgQueue", "dequeue", msg)
		case msg = <-s.inMsgQueue2:
			s.LogMessage("InMsgQueue2", "dequeue", msg)
		}

		if t := msg.Type(); t == constants.ACK_MSG {
			s.LogMessage("ackQueue", "enqueue", msg)
			s.ackQueue <- msg
		} else {
			s.LogMessage("msgQueue", "enqueue", msg)
			s.msgQueue <- msg
		}
	}
}
