// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package leader

import (
	"time"
)

// EOM Timer
func (l *Leader) EOMTimer() {
	l.loaded.Wait()
	for {
		// KLUGE somehow this is wrong?
		//tenthPeriod := 60*l.Config.FactomSecond.Nanoseconds()       // The length of the minute can change, so do this each time
		tenthPeriod := time.Duration(15*time.Second).Nanoseconds() / 10 // hardcode 15s blocks
		now := time.Now().UnixNano()                                    // Get the current time
		sleep := tenthPeriod - now%tenthPeriod                          //
		time.Sleep(time.Duration(sleep))                                // Sleep the length of time from now to the next minute
		l.SendEOM()
	}
}
