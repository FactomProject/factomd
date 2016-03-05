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
	nodeCnt     int
	
	start		interfaces.Timestamp
	msgCnt		int
	msgPerSec   int
	
	// The last period (msg rate over the last period, so msg changes can be seen)
	period		int
	startp		interfaces.Timestamp
	msgCntp		int
	msgPerSecp	int
}

func (m *MsgLog) init(nodecnt int) {
	m.nodeCnt = nodecnt
	if nodecnt == 0 {
		m.nodeCnt = 1
	}
}

func (m *MsgLog) add2(fnode *FactomNode, dest string , where string, valid bool, msg interfaces.IMsg) {
	m.sem.Lock()
	defer m.sem.Unlock()
	
	now := fnode.State.GetTimestamp() / 1000
	if m.start == 0 {
		m.start = fnode.State.GetTimestamp() / 1000
		m.last = m.start		// last is start
		m.period = 2
		m.startp = m.start
	}
		
	nm := new(msglist)
	nm.fnode = fnode
	nm.name = fnode.State.FactomNodeName
	nm.dest = dest
	nm.valid = valid
	nm.where = where
	nm.msg = msg
	m.MsgList = append(m.MsgList, nm)
	
	if now-m.start > 1  {
		m.msgPerSec = (m.msgCnt+len(m.MsgList)) / int(now-m.start) / m.nodeCnt
	}
	if int(now-m.startp) >= m.period {
		m.msgPerSecp = (m.msgCntp+len(m.MsgList)) / int(now-m.startp) / m.nodeCnt
		m.msgCntp = 0
		m.startp = now	// Reset timer
	}
	
	// If it has been 2 seconds, and we are printing, then print
	if now-m.last > 2 && fnode.State.GetOut() {
		m.prtMsgs(fnode.State)
		m.last = now
		m.msgCnt += len(m.MsgList)	// Keep my counts
		m.msgCntp += len(m.MsgList)
		m.MsgList = m.MsgList[0:0]  // Once printed, clear the list
		// If it has been 4 seconds and we are NOT printing, then toss.
		// This gives us a second to get to print.
	}else if now-m.last > 4 {
		m.msgCnt += len(m.MsgList)	// Keep my counts
		m.msgCntp += len(m.MsgList)
		m.MsgList = m.MsgList[0:0]	// Clear the record.
	}
}

func (m *MsgLog) prtMsgs(state interfaces.IState) {

	state.Println("\n***************************************************")
	state.Println(fmt.Sprintf("*** %42s ****", "State: "+state.String()))
	
	for _, e := range m.MsgList {
		if e.valid {
			if e.fnode.State.GetOut() { 
				state.Print(fmt.Sprintf("**** %8s -> %8s %10s %5v     **** %s\n", e.name, e.dest, e.where, e.valid, e.msg.String()))
			}
		}
	}
	state.Println(fmt.Sprintf("*** %42s **** ",fmt.Sprintf("Length: %d    Msgs/sec: T %d P %d",len(m.MsgList),m.msgPerSec,m.msgPerSecp)))
	state.Println("***************************************************\n")
}

func NetworkProcessorNet(fnode *FactomNode) {
	
	for {

		// Put any broadcasts from our peers into our BroadcastIn queue
		for i, peer := range fnode.Peers {
			for {
				msg, err := peer.Recieve()
				if err != nil && msg != nil {
					msg.SetOrigin(i+1)
					if fnode.State.Replay.IsTSValid_(msg.GetMsgHash().Fixed(),
						int64(msg.GetTimestamp())/1000,
						int64(fnode.State.GetTimestamp())/1000) {
						//fnode.State.Println("In Comming!! ",msg)
						fnode.MLog.add2(fnode, peer.GetName(), "PeerIn", true, msg)
						fnode.State.NetworkInMsgQueue() <- msg
					} else {
						//fnode.MLog.add2(fnode, peer.name, "PeerIn", false, msg)
					}
				} else {
					break
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
						fmt.Println("Waiting for the Network")
						time.Sleep(1 * time.Second)
						fnode.State.NetworkOutMsgQueue() <- msg
						break
					}
					if p < 0 {
						p = rand.Int() % len(fnode.Peers)
					}

					fnode.MLog.add2(fnode, fnode.Peers[p].GetName(), "P2P out", true, msg)
					fnode.Peers[p].Send(msg)

				} else {
					p := msg.GetOrigin() - 1
					for i, peer := range fnode.Peers {
						// Don't resend to the node that sent it to you.
						if i != p {
							bco := fmt.Sprintf("%s/%d/%d","BCast",p,i)
							fnode.MLog.add2(fnode, peer.GetName(), bco, true, msg)
							peer.Send(msg)
						}
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
