// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/log"
	"math/rand"
	"time"
)

var _ = log.Printf
var _ = fmt.Print

func NetworkProcessorNet(fnode *FactomNode) {

	for {
		// Put any broadcasts from our peers into our BroadcastIn queue
		for i, peer := range fnode.Peers {
			for {
				msg, err := peer.Recieve()
				if err == nil && msg != nil {
					msg.SetOrigin(i + 1)
					if fnode.State.Replay.IsTSValid_(msg.GetMsgHash().Fixed(),
						int64(msg.GetTimestamp())/1000,
						int64(fnode.State.GetTimestamp())/1000) {
						//fnode.State.Println("In Comming!! ",msg)
						nme := fmt.Sprintf("%s %d", "PeerIn", i+1)
						fnode.MLog.add2(fnode, peer.GetNameTo(), nme, true, msg)

						fnode.State.InMsgQueue() <- msg

					} else {
						fnode.MLog.add2(fnode, peer.GetNameTo(), "PeerIn", false, msg)
					}
				} else {
					if err != nil {
						fmt.Println(fnode.State.GetFactomNodeName(), err)
					}
					break
				}
			}
		}

		select {
		case msg, ok := <-fnode.State.NetworkOutMsgQueue():
			if ok {
				// We don't care about the result, but we do want to log that we have
				// seen this message before, because we might have generated the message
				// ourselves.
				fnode.State.Replay.IsTSValid_(msg.GetMsgHash().Fixed(),
					int64(msg.GetTimestamp())/1000,
					int64(fnode.State.GetTimestamp())/1000)

				if msg.IsPeer2peer() {
					p := msg.GetOrigin() - 1
					if len(fnode.Peers) == 0 {
						// No peers yet, put back in queue
						fmt.Print(" Waiting ")
						time.Sleep(1 * time.Second)
						fnode.State.NetworkOutMsgQueue() <- msg
						break
					}
					if p < 0 {
						p = rand.Int() % len(fnode.Peers)
					}
					
					fnode.MLog.add2(fnode, fnode.Peers[p].GetNameTo(), "P2P out", true, msg)
					fnode.Peers[p].Send(msg)

				} else {
					p := msg.GetOrigin() - 1
					for i, peer := range fnode.Peers {
						// Don't resend to the node that sent it to you.
						if i != p || true {
							bco := fmt.Sprintf("%s/%d/%d", "BCast", p, i)
							fnode.MLog.add2(fnode, peer.GetNameTo(), bco, true, msg)
							peer.Send(msg)
						}
					}
				}
			}
		case msg, ok := <-fnode.State.NetworkInvalidMsgQueue():
			if ok {
				//				fnode.State.Println("\n&&&&&&&&&&&&&&&&&&&&&&&&&&&&&& Bad Message %%%%%%%%%%%%%%%%%%%%%%%%")
				var _ = msg
				if fnode.State.PrintType(msg.Type()) {

				}
			}
		default:
			time.Sleep(time.Duration(10) * time.Millisecond)
		}
	}

}
