// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	s "github.com/FactomProject/factomd/state"
	"math/rand"
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

	delta := rand.Int63() % 10

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
			for len(state.InMsgQueue()) > 5000 {
				time.Sleep(100 * time.Millisecond)
			}

			// Delay some number of milliseconds.
			time.Sleep(time.Duration(delta) * time.Millisecond)

			state.TickerQueue() <- i

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
