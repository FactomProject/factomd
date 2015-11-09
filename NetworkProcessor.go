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

var _ = log.Printf
var _ = fmt.Print

func NetworkProcessor(state interfaces.IState) {

netloop:
	for {

		// This loop looks at the input queues and the invalid queues and
		// Handles messages as they come in.   If nothing is read, it sleeps
		// for 500 milliseconds.  Note you only sleep if both queues test
		// to be empty.

		select {
		case msg, ok := <-state.NetworkInMsgQueue():
			if ok {
				//log.Printf("%20s %s\n", "In Network:", msg.String())
				state.InMsgQueue() <- msg
				continue netloop
			}
		default:
		}

		select {
		case msg, ok := <-state.NetworkOutMsgQueue():
			if ok {
				var _ = msg
				//log.Printf("%20s %s\n", "Network Broadcast:", msg.String())
				// Ignore for now
				continue netloop
			}
		default:
		}

		select {
		case msg, ok := <-state.NetworkInvalidMsgQueue():
			if ok {
				var _ = msg
				//log.Printf("%20s %s\n", "Network Invalid Msg:", msg.String())
				// Ignore for now
				continue netloop
			}
		default:
		}
		time.Sleep(time.Duration(500) * time.Millisecond)
	}

}
