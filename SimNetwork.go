// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/FactomProject/factomd/btcd"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/go-spew/spew"
	//"reflect"
	"time"
)

var _ = spew.Sdump
var _ = log.Printf
var _ = fmt.Print

func SimNetwork(states ...interfaces.IState) {

	state := states[0]

netloop:
	for {

		// This loop looks at the input queues and the invalid queues and
		// Handles messages as they come in.   If nothing is read, it sleeps
		// for 500 milliseconds.  Note you only sleep if both queues test
		// to be empty.

		select {
		case msg, ok := <-state.NetworkInMsgQueue():
			if ok {
				//log.Printf("NetworkIn: %s\n", spew.Sdump(msg))
				if state.PrintType(msg.Type()) {
					//log.Printf("Ignored: NetworkIn: %s\n", msg.String())
				}
				state.InMsgQueue() <- msg
				continue netloop
			}
		default:
		}

		select {
		case msg, ok := <-state.NetworkOutMsgQueue():
			if ok {
				var _ = msg
				//log.Printf("NetworkOut: %s\n", msg.String())
				if state.PrintType(msg.Type()) {
					//log.Printf("Ignored: NetworkOut: %s\n", msg.String())
				}
				switch msg.(type) {
				case *messages.EOM:
					msgeom := msg.(*messages.EOM)
					server := state.GetServer().(*btcd.Server)
					server.BroadcastMessage(msgeom)

				default:
				}

				continue netloop
			}
		default:
		}

		select {
		case msg, ok := <-state.NetworkInvalidMsgQueue():
			if ok {
				var _ = msg
				if state.PrintType(msg.Type()) {
					//log.Printf("%20s %s\n", "Invalid:", msg.String())
				}
				continue netloop
			}
		default:
		}
		time.Sleep(time.Duration(500) * time.Millisecond)
	}

}
