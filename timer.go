// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/util"
	"time"
)

func Timer(state interfaces.IState) {
	cfg := state.GetCfg().(*util.FactomdConfig)

	billion := int64(1000000000)
	period := int64(cfg.App.DirectoryBlockInSeconds) * billion
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

			// End of the last period, and this is a server, send messages that
			// close off the minute.
			if state.GetServerState() == 1 {
				eom := messages.NewEOM(state, i)
				state.LeaderInMsgQueue() <- eom
				//state.NetworkOutMsgQueue() <- eom
			}

			fmt.Printf("\r%19s: DBlock %v minute %v %s", 
					   "Timer", 
						state.GetDBHeight(), 
						i,
						(string)((([]byte)("-\\|/-\\|/-="))[i]))
		}
	}

}
