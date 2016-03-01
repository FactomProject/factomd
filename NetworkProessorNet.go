// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/log"
	"math/rand"
	"sync"
	"time"
)

var _ = log.Printf
var _ = fmt.Print

type msglist struct {
	name  string
	dest  string
	t     string // NetIn, In, Netout
	valid bool   // Valid or not
	msg   interfaces.IMsg
}

var sem *sync.Mutex
var MsgList []*msglist

func add2(name string, dest string, t string, valid bool, msg interfaces.IMsg) {
	if sem == nil {
		sem = new(sync.Mutex)
	}
	m := new(msglist)
	m.name = name
	m.dest = dest
	m.valid = valid
	m.t = t
	m.msg = msg
	sem.Lock()
	MsgList = append(MsgList, m)
	sem.Unlock()
}

func prt(state interfaces.IState) {
	sem.Lock()
	state.Println("\n***************************************************")
	state.Println(fmt.Sprintf("*** State: %35s ****", state.String()))
	for _, m := range MsgList {
		if m.valid {
			state.Print(fmt.Sprintf("*** %8s -> %8s %10s %5v      **** %s\n", m.name, m.dest, m.t, m.valid, m.msg.String()))
		}
	}
	state.Println("***************************************************\n")
	MsgList = MsgList[0:0]
	sem.Unlock()
}

func NetworkProcessorNet(fnode *FactomNode) {

	last := fnode.State.GetTimestamp() / 1000

	for {

		now := fnode.State.GetTimestamp() / 1000

		if now-last > 6 {
			prt(fnode.State)
			last = now
		}

		// Put any broadcasts from our peers into our BroadcastIn queue
		for _, peer := range fnode.Peers {
		loop:
			for {
				select {
				case msg, ok := <-peer.BroadcastIn:
					if ok {
						if fnode.State.Replay.IsTSValid_(msg.GetMsgHash().Fixed(),
							int64(msg.GetTimestamp())/1000,
							int64(fnode.State.GetTimestamp())/1000) {
							//fnode.State.Println("In Comming!! ",msg)
							add2(fnode.State.FactomNodeName, peer.name, "PeerIn", true, msg)
							fnode.State.NetworkInMsgQueue() <- msg
						} else {
							add2(fnode.State.FactomNodeName, peer.name, "PeerIn", false, msg)
						}
					}
				default:
					break loop
				}
			}
		}

		select {
		case msg, ok := <-fnode.State.NetworkInMsgQueue():
			if ok {
				add2(fnode.State.FactomNodeName, "--", "InMsgQ", true, msg)
				//fnode.State.Println("Msg Origin: ",msg.GetOrigin()," ",msg)
				fnode.State.InMsgQueue() <- msg
			} else if msg != nil {
				add2(fnode.State.FactomNodeName, "--", "InMsgQ", false, msg)
			}
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
						time.Sleep(1 * time.Second)
						fnode.State.NetworkOutMsgQueue() <- msg
						break
					}
					if p < 0 {
						p = rand.Int() % len(fnode.Peers)
					}

					add2(fnode.State.FactomNodeName, fnode.Peers[p].name, "P2P out", false, msg)
					fnode.Peers[p].BroadcastOut <- msg

				} else {
					for _, peer := range fnode.Peers {
						add2(fnode.State.FactomNodeName, peer.name, "BCast out", true, msg)
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
