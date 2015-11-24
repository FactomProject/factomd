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
		msg := <-state.InMsgQueue()  // Get message from the input queue
		switch msg.Validate(state) { // Validate the message.
		case 1: // Process if valid
			if msg.Leader(state) {
				state.LeaderInMsgQueue() <- msg
			} else if msg.Follower(state) {
				state.FollowerInMsgQueue() <- msg
			}
		case 0: // Hold for later if unknown.
		case -1: // Drop if invalid.
		}
	}

}
