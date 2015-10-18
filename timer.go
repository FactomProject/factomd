// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/util"
	"time"
)

type timer struct {
	int minute		// Really a 10th of the period... Usually a minute	
}

var _ interfaces.IMsg = (*timer)(nil)

func (timer) Type() int {
	return constants.EOM_MSG
}

func (t *timer) Payload() {
	payload := append([]byte, byte(t.minute))
	return payload
}

func (t *timer) Validate(state IState) int {
	

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
			IMsg
			state.In	
			now = time.Now().UnixNano() 
			wait := next - now
			next += tenthPeriod
			time.Sleep(time.Duration(wait))
		}
	}
	
}