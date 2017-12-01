// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type MessageBase struct {
	FullMsgHash interfaces.IHash

	Origin        int    // Set and examined on a server, not marshaled with the message
	NetworkOrigin string // Hash of the network peer/connection where the message is from
	Peer2Peer     bool   // The nature of this message type, not marshaled with the message
	LocalOnly     bool   // This message is only a local message, is not broadcasted and may skip verification

	NoResend  bool // Don't resend this message if true.
	ResendCnt int  // Put a limit on resends

	LeaderChainID interfaces.IHash
	MsgHash       interfaces.IHash // Cache of the hash of a message
	RepeatHash    interfaces.IHash // Cache of the hash of a message
	VMIndex       int              // The Index of the VM responsible for this message.
	VMHash        []byte           // Basis for selecting a VMIndex
	Minute        byte
	resend        int64 // Time to resend (milliseconds)
	expire        int64 // Time to expire (milliseconds)

	Ack interfaces.IMsg

	Stalled     bool // This message is currently stalled
	MarkInvalid bool
	Sigvalid    bool
}

func resend(state interfaces.IState, msg interfaces.IMsg, cnt int, delay int) {
	for i := 0; i < cnt; i++ {
		state.NetworkOutMsgQueue().Enqueue(msg)
		time.Sleep(time.Duration(delay) * time.Second)
	}
}

func (m *MessageBase) GetAck() interfaces.IMsg {
	return m.Ack
}

func (m *MessageBase) PutAck(ack interfaces.IMsg) {
	m.Ack = ack
}

func (m *MessageBase) SendOut(state interfaces.IState, msg interfaces.IMsg) {
	// Dont' resend if we are behind
	if m.ResendCnt > 0 {
		return
	}

	m.ResendCnt++

	switch msg.(interface{}).(type) {
	//case ServerFault:
	//	go resend(state, msg, 20, 1)
	case FullServerFault:
		go resend(state, msg, 2, 5)
	case ServerFault:
		go resend(state, msg, 2, 5)
	default:
		go resend(state, msg, 1, 0)
	}
}

func (m *MessageBase) GetNoResend() bool {
	return m.NoResend
}

func (m *MessageBase) SetNoResend(v bool) {
	m.NoResend = v
}

func (m *MessageBase) IsValid() bool {
	return m.Sigvalid
}

func (m *MessageBase) SetValid() {
	m.Sigvalid = true
}

// To suppress how many messages are sent to the NetworkInvalid Queue, we mark them, and only
// send them once.
func (m *MessageBase) MarkSentInvalid(b bool) {
	m.MarkInvalid = b
}

func (m *MessageBase) SentInvalid() bool {
	return m.MarkInvalid
}

const secs = 10 // How many seconds we are going to put between resends.

// Try and Resend.  Return true if we should keep the message, false if we should give up.
func (m *MessageBase) Resend(state interfaces.IState) (rtn bool) {
	now := state.GetTimestamp().GetTimeMilli()
	if m.resend == 0 {
		m.resend = now
		return false
	}
	if now-m.resend > secs*1000 { // now is in milliseconds.  x1000 makes it seconds
		m.ResendCnt += 1
		if state.NetworkOutMsgQueue().Length() < 1000 {
			m.resend = now
			return true
		}
	}
	return false
}

// Try and Resend.  Return false if we should keep the message, true if we should expire the message.
func (m *MessageBase) Expire(state interfaces.IState, msg interfaces.IMsg) (rtn bool) {

	// minutes is a local fucntion that let's us estimate the time we have spent
	// attempting to retry sending a message from the holding queue in minutes given
	// how many seconds we wait between Resend() actually resending a message.
	minutes := func(i int) int {
		return i * 60 / secs
	}

	// If a message has not validated and our holding queue has some amount of
	// pending messages, then toss non-validatable messages after 5 minutes.
	vf := msg.Validate(state)
	if state.HoldingLen() > 100 && vf == 0 && m.ResendCnt > minutes(5) {
		return true
	}

	// queue is backing up, hold for 2 min
	if state.HoldingLen() > 1500 && m.ResendCnt > minutes(2) {
		return true
	}

	// Okay, a little worried, hold for 10 min
	if state.HoldingLen() > 1000 && m.ResendCnt > minutes(10) {
		return true
	}

	// Not too worried, hold for 15
	if state.HoldingLen() > 500 && m.ResendCnt > minutes(15) {
		return true
	}

	// Just want to rush a bit, hold for 20
	if state.HoldingLen() > 200 && m.ResendCnt > minutes(20) {
		return true
	}

	// Not worried at all, hold for 60 minutes
	if m.ResendCnt > minutes(60) { // Wait an hour
		return true
	}

	return false
}

func (m *MessageBase) IsStalled() bool {
	return m.Stalled
}
func (m *MessageBase) SetStall(b bool) {
	m.Stalled = b
}

func (m *MessageBase) GetFullMsgHash() interfaces.IHash {
	if m.FullMsgHash == nil {
		m.FullMsgHash = primitives.NewZeroHash()
	}
	return m.FullMsgHash
}

func (m *MessageBase) SetFullMsgHash(hash interfaces.IHash) {
	m.GetFullMsgHash().SetBytes(hash.Bytes())
}

func (m *MessageBase) GetOrigin() int {
	return m.Origin
}

func (m *MessageBase) SetOrigin(o int) {
	m.Origin = o
}

func (m *MessageBase) GetNetworkOrigin() string {
	return m.NetworkOrigin
}

func (m *MessageBase) SetNetworkOrigin(o string) {
	m.NetworkOrigin = o
}

// Returns true if this is a response to a peer to peer
// request.
func (m *MessageBase) IsPeer2Peer() bool {
	return m.Peer2Peer
}

func (m *MessageBase) SetPeer2Peer(f bool) {
	m.Peer2Peer = f
}

func (m *MessageBase) IsLocal() bool {
	return m.LocalOnly
}

func (m *MessageBase) SetLocal(v bool) {
	m.LocalOnly = v
}

func (m *MessageBase) GetLeaderChainID() interfaces.IHash {
	if m.LeaderChainID == nil {
		m.LeaderChainID = primitives.NewZeroHash()
	}
	return m.LeaderChainID
}

func (m *MessageBase) SetLeaderChainID(hash interfaces.IHash) {
	m.LeaderChainID = hash
}

func (m *MessageBase) GetVMIndex() (index int) {
	index = m.VMIndex
	return
}

func (m *MessageBase) SetVMIndex(index int) {
	m.VMIndex = index
}

func (m *MessageBase) GetVMHash() []byte {
	return m.VMHash
}

func (m *MessageBase) SetVMHash(vmhash []byte) {
	m.VMHash = vmhash
}

func (m *MessageBase) GetMinute() byte {
	return m.Minute
}

func (m *MessageBase) SetMinute(minute byte) {
	m.Minute = minute
}
