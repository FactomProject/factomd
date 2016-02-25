// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"math/rand"
	"github.com/FactomProject/factomd/log"
	"time"
)

var _ = log.Printf
var _ = fmt.Print

func NetworkProcessorNet(fnode *FactomNode) {

	for {
		
		// Put any broadcasts from our peers into our BroadcastIn queue
		for _,peer := range fnode.Peers {
			loop: for {
				select {
				case msg, ok := <- peer.BroadcastIn:
					if ok {
						//fnode.State.Println("In Comming!! ",msg)
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
				//fnode.State.Println("Msg Origin: ",msg.GetOrigin()," ",msg)
				fnode.State.InMsgQueue() <- msg
			}
		case msg, ok := <-fnode.State.NetworkOutMsgQueue():
			if ok {
				// We don't care about the result, but we do want to log that we have
				// seen this message before, because we might have generated the message
				// ourselves.
				IsTSValid_(msg.GetMsgHash().Fixed(),
						   int64(msg.GetTimestamp())/1000,
						   int64(fnode.State.GetTimestamp())/1000)
				if msg.IsPeer2peer() {
					
					fnode.State.Print("P2P Msg", msg)
					
					p := msg.GetOrigin()-1
					if p < 0 {
						p = rand.Int()%len(fnode.Peers)
					}
						
					fnode.Peers[p].BroadcastOut <- msg
					
				}else{
					for _, peer := range fnode.Peers {
						peer.BroadcastOut <- msg
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

