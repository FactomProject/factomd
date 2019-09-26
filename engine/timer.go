// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"os"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/state"
)

// Timer
// Provides a tick add inserts it into the TickerQueue to trigger EOM generation by
// leaders.
func Timer(stateI interfaces.IState) {
	s := stateI.(*state.State)

	for {
		for i := 0; i < 10; i++ {
			tenthPeriod := s.GetMinuteDuration().Nanoseconds()       // The length of the minute can change, so do this each time
			now := time.Now().UnixNano()                             // Get the current time
			time.Sleep(time.Duration(tenthPeriod - now%tenthPeriod)) // Sleep the length of time from now to the next minute

			// Delay some number of milliseconds.  This is a debugging tool for testing how well we handle
			// Leaders running with slightly different minutes in test environments.
			time.Sleep(time.Duration(s.GetTimeOffset().GetTimeMilli()) * time.Millisecond)

			if s.LastSyncTime.Nanoseconds() > tenthPeriod*9/10 && s.Leader {
				s.LastSyncTime = 0
				_, _ = os.Stderr.WriteString(fmt.Sprintf("Slipping a minute %10s %d  dbht %5d\n", s.FactomNodeName, s.CurrentMinute, s.LLeaderHeight))

			} else {
				s.TickerQueue() <- -1 // -1 indicated this is real minute cadence

				s.LogPrintf("ticker", "Tick! %d, tenthPeriod=%s", i, time.Duration(tenthPeriod))
			}
		}
	}
}
