// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"time"
)

// Timer
// Provides a tick add inserts it into the TickerQueue to trigger EOM generation by
// leaders.
func (s *State) Timer() {

	var last int64
	for {
		tenthPeriod := s.GetMinuteDuration().Nanoseconds() // The length of the minute can change, so do this each time
		now := time.Now().UnixNano()                       // Get the current time
		sleep := tenthPeriod - now%tenthPeriod
		time.Sleep(time.Duration(sleep)) // Sleep the length of time from now to the next minute

		// Delay some number of milliseconds.  This is a debugging tool for testing how well we handle
		// Leaders running with slightly different minutes in test environments.
		time.Sleep(time.Duration(s.GetTimeOffset().GetTimeMilli()) * time.Millisecond)

		if s.Leader {
			now = time.Now().UnixNano()
			issueTime := last
			if s.EOMSyncEnd > s.EOMIssueTime {
				issueTime = s.EOMIssueTime
			}
			last = s.EOMIssueTime

			if s.EOMIssueTime-issueTime > tenthPeriod*8/100 {
				s.EOMIssueTime = 0 // Don't skip more than one ticker
				s.LogPrintf("ticker", "10%s Skip ticker Sleep %4.2f", s.FactomNodeName, float64(sleep)/1000000000)
				continue
			}
		}
		s.LogPrintf("ticker", "10%s send ticker Sleep %4.2f", s.FactomNodeName, float64(sleep)/1000000000)
		s.EOMSyncEnd = 0

		s.TickerQueue() <- -1 // -1 indicated this is real minute cadence
	}
}
