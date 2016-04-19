// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	s "github.com/FactomProject/factomd/state"
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

	state.Print(fmt.Sprintf("Time: %v\r\n", time.Now()))
	time.Sleep(time.Duration(wait))
	for {
       
		found, index := state.GetFedServerIndexHash(state.GetLeaderHeight(), state.GetIdentityChainID())
        sent := false
		for i := 0; i < 10; i++ {
			now = time.Now().UnixNano()
			wait := next - now
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
            for found && len(state.InMsgQueue()) > 5000 {
				fmt.Println("Skip Period")
				time.Sleep(time.Duration(tenthPeriod))
			}
        	
        	// End of the last period, and this is a server, send messages that
			// close off the minute.
        
          
           //fmt.Println("         ", "found",found,"green",state.Green(), "sent",sent,"i", i,"dbheight",state.GetLeaderHeight())
           
           if found && state.Green() && (sent || i==0){
                sent = true
				eom := new(messages.EOM)
				eom.Minute = byte(i)
				eom.Timestamp = state.GetTimestamp()
				eom.ChainID = state.GetIdentityChainID()
				eom.ServerIndex = index
				eom.Sign(state)
		        eom.DBHeight = state.GetLeaderHeight()
                if i == 9 {
                    DBS := new(messages.DirectoryBlockSignature)
                    DBS.ServerIdentityChainID = state.GetIdentityChainID()
                    DBS.LocalOnly = true
                    DBS.DBHeight = state.GetLeaderHeight()
                    DBS.ServerIndex = uint32(index)
    				state.TimerMsgQueue() <- eom
                    state.TimerMsgQueue() <- DBS
                }else{
      				state.TimerMsgQueue() <- eom
                }
			}
		}
	}
}

func Throttle(state interfaces.IState) {
	time.Sleep(2 * time.Second)

	throttlePeriod := time.Duration(int64(state.GetDirectoryBlockInSeconds()) * 50000000)

	for {
		time.Sleep(throttlePeriod)
		state.Dethrottle()
	}

}

func PrintBusy(state interfaces.IState, i int) {

	s := state.(*s.State)

	if len(s.ShutdownChan) == 0 {
		state.Print(fmt.Sprintf("\r%19s: %s %s",
			"Timer",
			state.String(),
			(string)((([]byte)("-\\|/-\\|/-="))[i])))
	}

}
