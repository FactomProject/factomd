// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package msgbase

import (
	"errors"
	"fmt"
	"sync"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/util/atomic"
)

type MessageBase struct {
	FullMsgHash interfaces.IHash

	Origin        int    // Set and examined on a server, not marshalled with the message
	NetworkOrigin string // Hash of the network peer/connection where the message is from
	Peer2Peer     bool   // The nature of this message type, not marshalled with the message
	LocalOnly     bool   // This message is only a local message, is not broadcast and may skip verification

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

	Stalled     bool // This message is currently stalled
	MarkInvalid bool
	Sigvalid    bool
}

var mu sync.Mutex // lock for debug struct

// keep the last N messages sent by each node
type msgHistory struct {
	mu      sync.Mutex // lock for debug struct
	where   map[[32]byte]string
	history *([1024][32]byte) // Last 1k messages hashes logged
	h       int               // head of history
	msgmap  map[[32]byte]interfaces.IMsg
}

func (f *msgHistory) addmsg(hash [32]byte, msg interfaces.IMsg, where string) {
	f.mu.Lock()
	if f.history == nil {
		f.history = new([1024][32]byte)
	}
	if f.msgmap == nil {
		f.msgmap = make(map[[32]byte]interfaces.IMsg)
		f.where = make(map[[32]byte]string)
	}
	remove := f.history[f.h] // get the oldest message
	delete(f.msgmap, remove)
	delete(f.where, remove)
	f.history[f.h] = hash
	f.where[hash] = where
	f.msgmap[hash] = msg
	f.h = (f.h + 1) % cap(f.history) // move the head
	f.mu.Unlock()
}

func (f *msgHistory) getmsg(hash [32]byte) (what interfaces.IMsg, where string, ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.msgmap == nil {
		f.msgmap = make(map[[32]byte]interfaces.IMsg)
		f.where = make(map[[32]byte]string)
	}
	what, ok = f.msgmap[hash]
	if ok {
		where = f.where[hash]
	}
	return what, where, ok
}

// keep one history per node (should move this to state and save the lookup?)
var outs map[string]*msgHistory = make(map[string]*msgHistory)

// counts are global to all nodes
var sends, unique, duplicate int

var logname = "duplicateSend" //"NetworkOutputs" to put then in the common place

func checkForDuplicateSend(s interfaces.IState, msg interfaces.IMsg, whereAmI string) {
	mu.Lock()
	f, ok := outs[s.GetFactomNodeName()]
	if !ok {
		f = new(msgHistory)
		outs[s.GetFactomNodeName()] = f
	}
	mu.Unlock()
	hash := msg.GetRepeatHash().Fixed()
	what, where, ok := f.getmsg(hash)
	if ok {
		duplicate++
		s.LogPrintf(logname, "Duplicate Send of R-%x (%d sends, %d duplicates, %d unique)", msg.GetRepeatHash().Bytes()[:4], sends, duplicate, unique)
		s.LogPrintf(logname, "Original: %p: %s", what, where)
		s.LogPrintf(logname, "This:     %p: %s", msg, whereAmI)
		s.LogMessage(logname, "Orig Message:", what)
		s.LogMessage(logname, "This Message:", msg)

	} else {
		unique++
		f.addmsg(hash, msg, whereAmI)
	}
}

func (m *MessageBase) SendOut(s interfaces.IState, msg interfaces.IMsg) {
	// Are we ever modifying a message?
	if m.ResendCnt > 4 { // If the first send fails, we need to try again
		return
	}
	now := s.GetTimestamp().GetTimeMilli()

	comment := fmt.Sprintf("Enqueue %v %v", m.ResendCnt, now-m.resend)
	s.LogMessage("NetworkOutputsCall", comment, msg)

	if m.ResendCnt > 1 { // If the first send fails, we need to try again
		//block := s.GetHighestKnownBlock()
		//blk := s.GetHighestSavedBlk()
		//if block-blk > 4 {
		//	return // don't resend when we are behind by more than a block
		//}
		if now-m.resend < 2000 {
			//			s.LogPrintf("NetworkOutputsCall", "too soon")
			return
		}
		if s.NetworkOutMsgQueue().Length() > s.NetworkOutMsgQueue().Cap()*99/100 {
			//			s.LogPrintf("NetworkOutputsCall", "too full
			return
		}

	}

	m1, m2 := fmt.Sprintf("%p", m), fmt.Sprintf("%p", msg)

	if m1 != m2 {
		panic("mismatch")
	}

	m.ResendCnt++
	m.resend = now
	sends++

	// debug code start ............
	if s.DebugExec() /* && s.CheckFileName(logname)*/ { // if debug is on and this logfile is enabled
		checkForDuplicateSend(s, msg, atomic.WhereAmIString(1))
	}
	// debug code end ............
	s.LogMessage("NetworkOutputs", comment, msg)
	s.NetworkOutMsgQueue().Enqueue(msg)
}

func (m *MessageBase) GetResendCnt() int {
	return m.ResendCnt
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

// Try and Resend.  Return true if we should keep the message, false if we should give up.
func (m *MessageBase) Resend(s interfaces.IState) (rtn bool) {
	return true
	if m.ResendCnt > 4 { // Only send four times ...
		return false
	}
	now := s.GetTimestamp().GetTimeMilli()
	if m.resend == 0 {
		m.resend = now
		return false
	}
	if now-m.resend > 2000 && s.NetworkOutMsgQueue().Length() < s.NetworkOutMsgQueue().Cap()*99/100 {
		return true
	}
	return false
}

// Try and Resend.  Return true if we should keep the message, false if we should give up.
func (m *MessageBase) Expire(s interfaces.IState) (rtn bool) {
	now := s.GetTimestamp().GetTimeMilli()
	if m.expire == 0 {
		m.expire = now
	}
	if now-m.expire > 60*60*1000 { // Keep messages for some length before giving up.
		rtn = true
	}
	return
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

func (m *MessageBase) SetOrigin(o int) { // Origin is one based but peers is 0 based.
	m.Origin = o
}

func (m *MessageBase) GetNetworkOrigin() string {
	return m.NetworkOrigin
}

func (m *MessageBase) SetNetworkOrigin(o string) {
	m.NetworkOrigin = o
}

// Returns true if this is a response to a peer to peer request.
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

func VerifyMessage(s interfaces.Signable) (bool, error) {
	if s.IsValid() {
		return true, nil
	}
	toSign, err := s.MarshalForSignature()
	if err != nil {
		return false, err
	}
	sig := s.GetSignature()
	if sig == nil {
		return false, fmt.Errorf("%s", "Message signature is nil")
	}
	if sig.Verify(toSign) {
		s.SetValid()
		return true, nil
	}
	return false, errors.New("Signature is invalid")
}

func SignSignable(s interfaces.Signable, key interfaces.Signer) (interfaces.IFullSignature, error) {
	toSign, err := s.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	sig := key.Sign(toSign)
	return sig, nil
}
