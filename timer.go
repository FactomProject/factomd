// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	s "github.com/FactomProject/factomd/state"
	"time"
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

	state.Print(fmt.Sprintf("Time: %v\r\n", time.Now()))
	time.Sleep(time.Duration(wait))
	for {
		found, index := state.GetFedServerIndex(state.GetLeaderHeight())
		for i := 0; i < 10; i++ {
			now = time.Now().UnixNano()
			wait := next - now
			next += tenthPeriod
			time.Sleep(time.Duration(wait))

			// PrintBush(state,i)

			// End of the last period, and this is a server, send messages that
			// close off the minute.
			if found {
				eom := new(messages.EOM)
				eom.Minute = byte(i)
				eom.Timestamp = state.GetTimestamp()
				eom.ChainID = state.GetIdentityChainID()
				eom.ServerIndex = index
				state.TimerMsgQueue() <- eom
			}
		}
	}

}

func PrintBusy(state interfaces.IState, i int) {

	s := state.(*s.State)

	if len(s.ShutdownChan) == 0 {
		state.Print(fmt.Sprintf("\r%19s: %s %s",
			"Timer",
			state.String(),
			(string)((([]byte)("-\\|/-\\|/-="))[i])))
	}

}
