// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"sync"

	"github.com/FactomProject/factomd/common/interfaces"
)

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
	Last    interfaces.Timestamp
	all     bool
	nodeCnt int

	start     interfaces.Timestamp
	msgCnt    int
	msgPerSec int

	// The last period (msg rate over the last period, so msg changes can be seen)
	Period     int64
	Startp     interfaces.Timestamp
	MsgCntp    int
	MsgPerSecp int
}

func (m *MsgLog) Init(enable bool, nodecnt int) {
	m.Enable = enable
	m.nodeCnt = nodecnt
	if nodecnt == 0 {
		m.nodeCnt = 1
	}
}

func (m *MsgLog) Add2(fnode *FactomNode, out bool, peer string, where string, valid bool, msg interfaces.IMsg) {
	if !m.Enable {
		return
	}
	m.sem.Lock()
	defer m.sem.Unlock()
	now := fnode.State.GetTimestamp()
	if m.start == nil {
		m.start = fnode.State.GetTimestamp()
		m.Last = m.start // last is start
		m.Period = 2
		m.Startp = m.start
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

	interval := int(now.GetTimeMilli() - m.start.GetTimeMilli())
	if interval == 0 || m.nodeCnt == 0 {
		return
	}

	if now.GetTimeSeconds()-m.start.GetTimeSeconds() > 1 {
		m.msgPerSec = (m.msgCnt + len(m.MsgList)) / interval / m.nodeCnt
	}
	if now.GetTimeSeconds()-m.Startp.GetTimeSeconds() >= m.Period {
		m.MsgPerSecp = (m.MsgCntp + len(m.MsgList)) / interval / m.nodeCnt
		m.MsgCntp = 0
		m.Startp = now // Reset timer
	}
	// If it has been 4 seconds and we are NOT printing, then toss.
	// This gives us a second to get to print.
	if now.GetTimeSeconds()-m.Last.GetTimeSeconds() > 3 {
		m.msgCnt += len(m.MsgList) // Keep my counts
		m.MsgCntp += len(m.MsgList)
		m.MsgList = make([]*msglist, 0) // Clear the record.
		m.Last = now
	}

}

func (m *MsgLog) PrtMsgs(state interfaces.IState) {
	if !m.Enable {
		fmt.Println("Message log is not enabled. Run factomd with runtime log enabled.")
		return
	}
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
	now := state.GetTimestamp()
	m.Last = now
	m.msgCnt += len(m.MsgList) // Keep my counts
	m.MsgCntp += len(m.MsgList)
	m.MsgList = m.MsgList[0:0] // Once printed, clear the list

	fmt.Println(fmt.Sprintf("*** %42s **** ", fmt.Sprintf("Length: %d    Msgs/sec: T %d P %d", len(m.MsgList), m.msgPerSec, m.MsgPerSecp)))
	fmt.Println("\n-----------------------------------------------------")
}
