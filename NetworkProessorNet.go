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

netloop:
	for {
		
		// Put any broadcasts from our peers into our BroadcastIn queue
		for i,peer := range fnode.Peers {
			loop: for {
				select {
				case msg, ok := <- peer.BroadcastIn:
					if ok {
						msg.SetOrigin(i+1)		// Remember this came from outside!
						fnode.State.NetworkInMsgQueue() <- msg 
					}
				default:
					break loop
				}
			}
		}

		loop2: for {
			select {
			case msg, ok := <-fnode.State.NetworkInMsgQueue():
				if ok {
					//log.Printf("NetworkIn: %s\n", spew.Sdump(msg))
					if fnode.State.PrintType(msg.Type()) {
						//log.Printf("Ignored: NetworkIn: %s\n", msg.String())
					}
					fnode.State.InMsgQueue() <- msg
					continue netloop
				}
			default:
				break loop2
			}
		}
		
		loop3: for {
			select {
			case msg, ok := <-fnode.State.NetworkOutMsgQueue():
				if ok {
					for i, peer := range fnode.Peers {
						if msg.GetOrigin() != i+1 {
							peer.BroadcastOut <- msg
							fmt.Println(msg.String())
						}
					}
				}
			default:
				break loop3
			}
		}
		
		loop4: for {
			select {
			case msg, ok := <-fnode.State.NetworkInvalidMsgQueue():
				if ok {
					var _ = msg
					if fnode.State.PrintType(msg.Type()) {
						
					}
					continue netloop
				}
			default:
				break loop4
			}
		}
		time.Sleep(time.Duration(500) * time.Millisecond)
	}

}

