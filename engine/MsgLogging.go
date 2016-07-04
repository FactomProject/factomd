// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"sync"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
)

var _ = log.Printf
var _ = fmt.Print

type msglist struct {
	fnode *FactomNode
	out   bool // True if this is an output.
	name  string
	peer  string
	where string // NetIn, In, Netout
	valid bool   // Valid or not
	msg   interfaces.IMsg
}

type MsgLog struct {
	Enable  bool
	sem     sync.Mutex
	MsgList []*msglist
	last    interfaces.Timestamp
	all     bool
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

func (m *MsgLog) init(enable bool, nodecnt int) {
	m.Enable = enable
	m.nodeCnt = nodecnt
	if nodecnt == 0 {
		m.nodeCnt = 1
	}

	m.last = new(primitives.Timestamp)
	m.start = new(primitives.Timestamp)
	m.startp = new(primitives.Timestamp)
}

func (m *MsgLog) add2(fnode *FactomNode, out bool, peer string, where string, valid bool, msg interfaces.IMsg) {

	m.sem.Lock()
	defer m.sem.Unlock()
	now := fnode.State.GetTimestamp().GetTimeSeconds()
	if m.start.GetTimeSeconds() == 0 {
		m.start = fnode.State.GetTimestamp()
		m.last = m.start // last is start
		m.period = 2
		m.startp = m.start
	}

	nm := new(msglist)
	nm.fnode = fnode
	nm.out = out
	nm.name = fnode.State.FactomNodeName
	nm.peer = peer
	nm.valid = valid
	nm.where = where
	nm.msg = msg
	m.MsgList = append(m.MsgList, nm)

	interval := int(now - m.start.GetTimeSeconds())
	if interval == 0 || m.nodeCnt == 0 {
		return
	}

	if now-m.start.GetTimeSeconds() > 1 {
		m.msgPerSec = (m.msgCnt + len(m.MsgList)) / interval / m.nodeCnt
	}
	if int(now-m.startp.GetTimeSeconds()) >= m.period {
		m.msgPerSecp = (m.msgCntp + len(m.MsgList)) / interval / m.nodeCnt
		m.msgCntp = 0
		m.startp.SetTimeSeconds(now) // Reset timer
	}
	// If it has been 4 seconds and we are NOT printing, then toss.
	// This gives us a second to get to print.
	if now-m.last.GetTimeSeconds() > 100 {
		m.msgCnt += len(m.MsgList) // Keep my counts
		m.msgCntp += len(m.MsgList)
		m.MsgList = m.MsgList[0:0] // Clear the record.
		m.last.SetTimeSeconds(now)
	}

}

func (m *MsgLog) PrtMsgs(state interfaces.IState) {
	m.sem.Lock()
	defer m.sem.Unlock()

	if len(m.MsgList) == 0 {
		return
	}

	fmt.Println(state.String())
	fmt.Println("\n-----------------------------------------------------")

	for _, e := range m.MsgList {
		if e.valid {

			dirstr := "->"
			if !e.out {
				dirstr = "<-"
			}

			fmt.Print(fmt.Sprintf("**** %8s %2s %8s %10s %5v     **** %s\n", e.name, dirstr, e.peer, e.where, e.valid, e.msg.String()))

		}
	}
	m.last.SetTimestamp(state.GetTimestamp())
	m.msgCnt += len(m.MsgList) // Keep my counts
	m.msgCntp += len(m.MsgList)
	m.MsgList = m.MsgList[0:0] // Once printed, clear the list

	fmt.Println(fmt.Sprintf("*** %42s **** ", fmt.Sprintf("Length: %d    Msgs/sec: T %d P %d", len(m.MsgList), m.msgPerSec, m.msgPerSecp)))
	fmt.Println("\n-----------------------------------------------------")
}
