// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/util"
	"time"
)

type timer struct {
	minute			int			// Really a 10th of the period... Usually a minute	
}


func Timer(state interfaces.IState) {
	cfg := state.GetCfg().(*util.FactomdConfig)

	billion     := int64(1000000000)
	period      := int64(cfg.App.DirectoryBlockInSeconds)*billion
	tenthPeriod := period/10
	
	now := time.Now().UnixNano()	// Time in billionths of a second
	
	wait := tenthPeriod - (now % tenthPeriod)

	next  := now + wait + tenthPeriod
	
	fmt.Println("Time:",time.Now())
	fmt.Println("Waiting", wait/billion, "seconds to the top of the period")
	time.Sleep(time.Duration(wait))
	fmt.Println("Starting Timer!",time.Now())
	for {
		for i := 0; i < 10; i++ {
			eom := messages.NewEOM(state,i)
			state.InMsgQueue() <- eom
			now = time.Now().UnixNano() 
			wait := next - now
			next += tenthPeriod
			time.Sleep(time.Duration(wait))
		}
	}
	
}