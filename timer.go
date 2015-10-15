// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/util"
	"time"
)

func Timer(state interfaces.IState) {
	cfg := state.Cfg().(*util.FactomdConfig)

	billion     := int64(1000000000)
	period      := int64(cfg.App.DirectoryBlockInSeconds)*billion
	tenthPeriod := period/10
	
	now := time.Now().UnixNano()	// Time in billionths of a second
	
	wait := period - (now % period)

	next  := now + wait + tenthPeriod
	
	fmt.Println("Time:",time.Now())
	fmt.Println("Waiting", wait/billion, "seconds to the top of the period")
	time.Sleep(time.Duration(wait))
	fmt.Println("Starting Timer!",time.Now())

	for {
		for i := 0; i < 10; i++ {
			fmt.Println("Period",i+1,"--",time.Now())
			
			now = time.Now().UnixNano() 
			wait := next - now
			next += tenthPeriod
			time.Sleep(time.Duration(wait))
		}
	}
	
}