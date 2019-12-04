// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package leader

import (
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/pubsub"
	"github.com/FactomProject/factomd/worker"
	"time"
)

// EOM Timer
func (l *Leader) EOMTimer(w *worker.Thread) {
	movedToHt := l.mkChan()
	w.OnReady(func() {
		movedToHt.Subscribe(pubsub.GetPath("FNode0", event.Path.Seq))
	})

	w.OnRun(func() {
		l.loaded.Wait() // Should wait for dbht +1?
		for {
			// KLUGE somehow this is wrong?
			//tenthPeriod := 60*l.Config.FactomSecond.Nanoseconds()       // The length of the minute can change, so do this each time
			tenthPeriod := time.Duration(15*time.Second).Nanoseconds() / 10 // hardcode 15s blocks
			now := time.Now().UnixNano()                                    // Get the current time
			sleep := tenthPeriod - now%tenthPeriod                          //
			time.Sleep(time.Duration(sleep))                                // Sleep the length of time from now to the next minute
			<-movedToHt.Updates                                             // wait for older EOM to be consumed
			l.SendEOM()
		}
	})
}
