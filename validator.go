// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/log"
)

var _ = fmt.Print
var _ = log.Printf

func Validator(state interfaces.IState) {

	for {
		msg := <-state.InMsgQueue() // Get message from the input queue
		
		if state.IgnoreType(msg.Type()) {
			log.Printf("%20s %s", "Validator:", msg.String())
		}
		
		switch msg.Validate(state) { // Validate the message.
		case 1: // Process if valid
			state.NetworkOutMsgQueue() <- msg 
			if state.IgnoreType(msg.Type()) {
				log.Printf(" Valid\n")
			}
			if msg.Leader(state) {
				state.LeaderInMsgQueue() <- msg
			} else if msg.Follower(state) {
				state.FollowerInMsgQueue() <- msg
			} else {
				log.Printf(" Message ignored\n")
			}
		case 0: // Hold for later if unknown.
			if state.IgnoreType(msg.Type()) {
				log.Printf(" Hold\n")
			}
		default: 
			if state.IgnoreType(msg.Type()) {
				log.Printf(" Invalid\n")
			}
			state.NetworkInvalidMsgQueue() <- msg
		}
	}

}
