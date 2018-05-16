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

type foo struct {
	where map[[32]byte]string
	what  map[[32]byte]interfaces.IMsg
}

//
var outs map[string]*foo = make(map[string]*foo)
var places map[string]interfaces.IMsg = make(map[string]interfaces.IMsg)

var sends, unique, duplicate int

func (m *MessageBase) SendOut(s interfaces.IState, msg interfaces.IMsg) {
	// Are we ever modifying a message?
	if m.ResendCnt > 4 { // If the first send fails, we need to try again
		// TODO: Maybe have it not resend unless x time passed?
		return
	}
	if msg.GetNoResend() {
		return
	}
	//// Don't resend if we are behind
	//if m.ResendCnt > 1 && s.GetHighestKnownBlock()-s.GetHighestSavedBlk() > 4 {
	//	s.LogMessage("NetworkOutputs", "Drop too busy", msg)
	//	return
	//}
	m.ResendCnt++
	sends++

	//if s.DebugExec() { // debug code
	//	whereAmIString := atomic.WhereAmIString(1)
	//	hash := msg.GetRepeatHash().Fixed()
	//	mu.Lock()
	//	f, ok := outs[s.GetFactomNodeName()]
	//	if !ok {
	//		f = new(foo)
	//		f.where = make(map[[32]byte]string)
	//		f.what = make(map[[32]byte]interfaces.IMsg)
	//		outs[s.GetFactomNodeName()] = f
	//	}
	//	where, ok := f.where[hash]
	//	mu.Unlock()
	//
	//	if ok {
	//		duplicate++
	//		mu.Lock()
	//		orig := f.what[hash]
	//		mu.Unlock()
	//		s.LogPrintf("NetworkOutputs", "Duplicate Send of R-%x (%d sends, %d duplicates, %d unique)", msg.GetRepeatHash().Bytes()[:4], sends, duplicate, unique)
	//		s.LogPrintf("NetworkOutputs", "Original: %p: %s", orig, where)
	//		s.LogPrintf("NetworkOutputs", "This:     %p: %s", msg, whereAmIString)
	//		s.LogMessage("NetworkOutputs", "Orig Message:", msg)
	//		s.LogMessage("NetworkOutputs", "This Message:", msg)
	//
	//		//mu.Lock()
	//		//which, ok := places[whereAmIString]
	//		//mu.Unlock()
	//		//if ok {
	//		//	s.LogPrintf("NetworkOutputs", "message <%s> is original from %s\n", which.String(), whereAmIString)
	//		//}
	//	} else {
	//		unique++
	//		mu.Lock()
	//		f.where[hash] = whereAmIString
	//		f.what[hash] = msg
	//		///minute 9[whereAmIString] = msg
	//		mu.Unlock()
	//	}
	//}

	s.LogMessage("NetworkOutputs", "Enqueue", msg)
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
	now := s.GetTimestamp().GetTimeMilli()
	if m.resend == 0 {
		m.resend = now
		return false
	}
	if now-m.resend > 2000 && s.NetworkOutMsgQueue().Length() < s.NetworkOutMsgQueue().Cap()*99/100 {
		m.resend = now
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
	return false, errors.New("Signarue is invalid")
}

func SignSignable(s interfaces.Signable, key interfaces.Signer) (interfaces.IFullSignature, error) {
	toSign, err := s.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	sig := key.Sign(toSign)
	return sig, nil
}
