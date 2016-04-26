// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	s "github.com/FactomProject/factomd/state"
)

var _ = (*s.State)(nil)

func Timer(state interfaces.IState) {

	time.Sleep(2 * time.Second)

	billion := int64(1000000000)
	period := int64(state.GetDirectoryBlockInSeconds()) * billion
	tenthPeriod := period / 10

	now := time.Now().UnixNano() // Time in billionths of a second

	wait := tenthPeriod - (now % tenthPeriod)

	next := now + wait + tenthPeriod

	if state.GetOut() {
		state.Print(fmt.Sprintf("Time: %v\r\n", time.Now()))
	}
	time.Sleep(time.Duration(wait))
	lastDBHeight := state.GetLeaderHeight()
	for {
		found, _ := state.GetVirtualServers(lastDBHeight, 0, state.GetIdentityChainID())
		sent := false
	    minloop: for i := 0; i < 10; i++ {
			now = time.Now().UnixNano()
			wait := next - now
			if now > next {
				wait = 1
				for next < now {
					next += tenthPeriod
				}
				wait = next - now
			} else {
				wait = next - now
				next += tenthPeriod
			}
			time.Sleep(time.Duration(wait))
			for found && len(state.InMsgQueue()) > 5000 {
				fmt.Println("Skip Period")
				time.Sleep(time.Duration(tenthPeriod))
			}

			// End of the last period, and this is a server, send messages that
			// close off the minute.

			//fmt.Println("         ", "found",found,"green",state.Green(), "sent",sent,"i", i,"dbheight",state.GetLeaderHeight())
			if found && state.Green() && (sent || i == 0) {
				if i == 0 {
					if lastDBHeight != state.GetLeaderHeight() || state.GetLeaderHeight() == 0 {
						lastDBHeight = state.GetLeaderHeight()
					}
				} else if lastDBHeight < state.GetLeaderHeight() {
					break minloop // If the state progresses while we were generating messages, skip
				}
				_, indexes := state.GetVirtualServers(lastDBHeight, i, state.GetIdentityChainID())
				for vmIndex, _ := range indexes {
					sent = true
					eom := new(messages.EOM)
					eom.Minute = byte(i)
					eom.Timestamp = state.GetTimestamp()
					eom.ChainID = state.GetIdentityChainID()
					eom.VMIndex = vmIndex
					eom.Sign(state)
					eom.DBHeight = lastDBHeight
					eom.SetLocal(true)
					if i == 9 {
						DBS := new(messages.DirectoryBlockSignature)
						DBS.ServerIdentityChainID = state.GetIdentityChainID()
						DBS.SetLocal(true)
						DBS.DBHeight = lastDBHeight
						DBS.VMIndex = vmIndex
						state.TimerMsgQueue() <- eom
						state.TimerMsgQueue() <- DBS
          			} else {
						state.TimerMsgQueue() <- eom
     				}
				}

			}
		}
	}
}

func Throttle(state interfaces.IState) {
	time.Sleep(2 * time.Second)

	throttlePeriod := time.Duration(int64(state.GetDirectoryBlockInSeconds()) * 50000000)

	for {
		time.Sleep(throttlePeriod)
		state.Dethrottle()
	}

}

func PrintBusy(state interfaces.IState, i int) {

	s := state.(*s.State)

	if len(s.ShutdownChan) == 0 {
		if state.GetOut() {
			state.Print(fmt.Sprintf("\r%19s: %s %s",
				"Timer",
				state.String(),
				(string)((([]byte)("-\\|/-\\|/-="))[i])))
		}
	}

}
