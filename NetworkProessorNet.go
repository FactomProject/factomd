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
	fnode *FactomNode
	name  string
	dest  string
	where string // NetIn, In, Netout
	valid bool   // Valid or not
	msg   interfaces.IMsg
}

type MsgLog struct {	
	sem 		sync.Mutex
	MsgList 	[]*msglist
	last 		interfaces.Timestamp
}

func (m *MsgLog) init() {
	//m.sem = &sync.Mutex{}
}

func (m *MsgLog) add2(fnode *FactomNode, dest string , where string, valid bool, msg interfaces.IMsg) {
	m.sem.Lock()
	defer m.sem.Unlock()
	nm := new(msglist)
	nm.fnode = fnode
	nm.name = fnode.State.FactomNodeName
	nm.dest = dest
	nm.valid = valid
	nm.where = where
	nm.msg = msg
	m.MsgList = append(m.MsgList, nm)
	
	now := fnode.State.GetTimestamp() / 1000
	
	if now-m.last > 6 && fnode.State.GetOut() {
		m.prtMsgs(fnode.State)
		m.last = now
	}
	
}

func (m *MsgLog) prtMsgs(state interfaces.IState) {

	state.Println("\n***************************************************")
	state.Println(fmt.Sprintf("*** State: %35s ****", state.String()))
	
	state.Println("*** Length ",len(m.MsgList),"                         ***")
	for _, e := range m.MsgList {
		if e.valid {
			if e.fnode.State.GetOut() { 
				state.Print(fmt.Sprintf("*** %8s -> %8s %10s %5v      **** %s\n", e.name, e.dest, e.where, e.valid, e.msg.String()))
			}
		}
	}
	state.Println("***************************************************\n")
	m.MsgList = m.MsgList[0:0]  // Once printed, clear the list
}

func NetworkProcessorNet(fnode *FactomNode) {

	for {

		// Put any broadcasts from our peers into our BroadcastIn queue
		for i, peer := range fnode.Peers {
		loop:
			for {
				select {
				case msg, ok := <-peer.BroadcastIn:
					if ok {
						msg.SetOrigin(i+1)
						if fnode.State.Replay.IsTSValid_(msg.GetMsgHash().Fixed(),
							int64(msg.GetTimestamp())/1000,
							int64(fnode.State.GetTimestamp())/1000) {
							//fnode.State.Println("In Comming!! ",msg)
							fnode.MLog.add2(fnode, peer.name, "PeerIn", true, msg)
							fnode.State.NetworkInMsgQueue() <- msg
						} else {
							//fnode.MLog.add2(fnode, peer.name, "PeerIn", false, msg)
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
				fnode.MLog.add2(fnode, "--", "InMsgQ", true, msg)
				//fnode.State.Println("Msg Origin: ",msg.GetOrigin()," ",msg)
				fnode.State.InMsgQueue() <- msg
			} else if msg != nil {
				fnode.MLog.add2(fnode, "--", "InMsgQ", false, msg)
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

					fnode.MLog.add2(fnode, fnode.Peers[p].name, "P2P out", true, msg)
					fnode.Peers[p].BroadcastOut <- msg

				} else {
					for _, peer := range fnode.Peers {
						fnode.MLog.add2(fnode, peer.name, "BCast out", true, msg)
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
