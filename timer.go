// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/log"
	"time"
)

func Timer(state interfaces.IState) {

	time.Sleep(1 * time.Second)
	
	billion := int64(1000000000)
	period := int64(state.GetDirectoryBlockInSeconds()) * billion
	tenthPeriod := period / 10

	now := time.Now().UnixNano() // Time in billionths of a second

	wait := tenthPeriod - (now % tenthPeriod)

	next := now + wait + tenthPeriod

	log.Printfln("Time: %v", time.Now())
	log.Printfln("Waiting %d seconds to the top of the period", wait/billion)
	time.Sleep(time.Duration(wait))
	log.Printfln("Starting Timer! %v\n", time.Now())
	for {
		for i := 0; i < 10; i++ {
			now = time.Now().UnixNano()
			wait := next - now
			next += tenthPeriod
			time.Sleep(time.Duration(wait))

			pls := fmt.Sprintf(" #%d servers: ", state.GetTotalServers())
			for i := 0; i < state.GetTotalServers(); i++ {
				pls = fmt.Sprintf("%s #%d:%d;", pls, i+1, 0, i)
			}

			state.Print(fmt.Sprintf("\r%19s: %s %s",
				"Timer",
				state.String(),
				(string)((([]byte)("-\\|/-\\|/-="))[i])))
			// End of the last period, and this is a server, send messages that
			// close off the minute.
			if state.GetServerState() == 1 {
				eom := state.NewEOM(i)
				state.InMsgQueue() <- eom
			}
		}
	}

}
