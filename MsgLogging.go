// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/log"
	"sync"
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
	sem     sync.Mutex
	MsgList []*msglist
	last    interfaces.Timestamp
	nodeCnt int

	start     interfaces.Timestamp
	msgCnt    int
	msgPerSec int

	// The last period (msg rate over the last period, so msg changes can be seen)
	period     int
	startp     interfaces.Timestamp
	msgCntp    int
	msgPerSecp int
}

func (m *MsgLog) init(nodecnt int) {
	m.nodeCnt = nodecnt
	if nodecnt == 0 {
		m.nodeCnt = 1
	}
}

func (m *MsgLog) add2(fnode *FactomNode, dest string, where string, valid bool, msg interfaces.IMsg) {
	m.sem.Lock()
	defer m.sem.Unlock()
	now := fnode.State.GetTimestamp() / 1000
	if m.start == 0 {
		m.start = fnode.State.GetTimestamp() / 1000
		m.last = m.start // last is start
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

	interval := int(now-m.start)
	if interval == 0 || m.nodeCnt == 0 { return }
	
	if now-m.start > 1 {
		m.msgPerSec = (m.msgCnt + len(m.MsgList)) / interval / m.nodeCnt
	}
	if int(now-m.startp) >= m.period {
		m.msgPerSecp = (m.msgCntp + len(m.MsgList)) / interval / m.nodeCnt
		m.msgCntp = 0
		m.startp = now // Reset timer
	}

	// If it has been 2 seconds, and we are printing, then print
	if now-m.last > 2 && fnode.State.GetOut() {
		m.prtMsgs(fnode.State)
		m.last = now
		m.msgCnt += len(m.MsgList) // Keep my counts
		m.msgCntp += len(m.MsgList)
		m.MsgList = m.MsgList[0:0] // Once printed, clear the list
		// If it has been 4 seconds and we are NOT printing, then toss.
		// This gives us a second to get to print.
	} else if now-m.last > 4 {
		m.msgCnt += len(m.MsgList) // Keep my counts
		m.msgCntp += len(m.MsgList)
		m.MsgList = m.MsgList[0:0] // Clear the record.
	}
}

func (m *MsgLog) prtMsgs(state interfaces.IState) {

	state.Println(state.String())
	state.Println("\n***************************************************")

	for _, e := range m.MsgList {
		if e.valid {
			if e.fnode.State.GetOut() {
				state.Print(fmt.Sprintf("**** %8s -> %8s %10s %5v     **** %s\n", e.name, e.dest, e.where, e.valid, e.msg.String()))
			}
		}
	}
	state.Println(fmt.Sprintf("*** %42s **** ", fmt.Sprintf("Length: %d    Msgs/sec: T %d P %d", len(m.MsgList), m.msgPerSec, m.msgPerSecp)))
	state.Println("***************************************************\n")
	
}
