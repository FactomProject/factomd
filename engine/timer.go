// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	s "github.com/FactomProject/factomd/state"
)

var _ = (*s.State)(nil)

func Timer(stateI interfaces.IState) {
	state := stateI.(*s.State)
	time.Sleep(2 * time.Second)

	tenthPeriod := state.GetMinuteDuration()

	now := time.Now().UnixNano() // Time in billionths of a second

	wait := tenthPeriod.Nanoseconds() - (now % tenthPeriod.Nanoseconds())

	next := now + wait + tenthPeriod.Nanoseconds()

	if state.GetOut() {
		state.Print(fmt.Sprintf("Time: %v\r\n", time.Now()))
	}

	time.Sleep(time.Duration(wait))

	for {
		for i := 0; i < 10; i++ {
			// Don't stuff messages into the system if the
			// Leader is behind.
			for j := 0; j < 10 && len(state.AckQueue()) > 1000; j++ {
				time.Sleep(time.Millisecond * 10)
			}

			now = time.Now().UnixNano()
			if now > next {
				next += tenthPeriod.Nanoseconds()
				wait = next - now
			} else {
				wait = next - now
				next += tenthPeriod.Nanoseconds()
			}
			time.Sleep(time.Duration(wait))

			// Delay some number of milliseconds.
			time.Sleep(time.Duration(state.GetTimeOffset().GetTimeMilli()) * time.Millisecond)

			state.TickerQueue() <- -1 // -1 indicated this is real minute cadence

			tenthPeriod = state.GetMinuteDuration()
			state.LogPrintf("ticker", "Tick! %d, wait=%s, tenthPeriod=%s", i, time.Duration(wait), time.Duration(tenthPeriod))
		}
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
