// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/log"
	"time"
)

var _ = log.Printf
var _ = fmt.Print

func NetworkProcessorNet(fnode *FactomNode) {

	for {
		
		// Put any broadcasts from our peers into our BroadcastIn queue
		for i,peer := range fnode.Peers {
			loop: for {
				select {
				case msg, ok := <- peer.BroadcastIn:
					fnode.State.Println("In Comming!! ",msg)
					if ok {
						msg.SetOrigin(i+1)		// Remember this came from outside!
						fnode.State.NetworkInMsgQueue() <- msg 
					}
				default:
					break loop
				}
			}
		}

		select {
		case msg, ok := <-fnode.State.NetworkInMsgQueue():
			if ok {
				if fnode.State.PrintType(msg.Type()) {
					
				}
				fnode.State.Println("Msg Origin: ",msg.GetOrigin()," ",msg)
				fnode.State.InMsgQueue() <- msg
			}
		case msg, ok := <-fnode.State.NetworkOutMsgQueue():
			if ok {
				for i, peer := range fnode.Peers {
					if msg.GetOrigin() != i+1 {
						peer.BroadcastOut <- msg
						fnode.State.Println("Replying Back:",msg.GetOrigin()," ", i+1, " ", msg.String())
					}
				}
			}
		case msg, ok := <-fnode.State.NetworkInvalidMsgQueue():
			if ok {
				var _ = msg
				if fnode.State.PrintType(msg.Type()) {
					
				}
			}
		default:
			time.Sleep(time.Duration(5) * time.Millisecond)
		}
	}

}

