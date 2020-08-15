// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package msgbase

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
)

type MessageBase struct {
	FullMsgHash interfaces.IHash

	Origin        int    // Set and examined on a server, not marshalled with the message
	NetworkOrigin string // Hash of the network peer/connection where the message is from
	Peer2Peer     bool   // The nature of this message type, not marshalled with the message
	LocalOnly     bool   // This message is only a local message, is not broadcast and may skip verification
	Network       bool   // If we got this message from the network, it is true.  Not marsheled.
	FullBroadcast bool   // This is used for messages with no missing message support e.g. election related messages

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

	// The time the message was received by our node
	// Use time.Time vs Timestamp so we don't have to deal with nils
	// Also the code that adds this timestamp already has the time.Time
	LocalReceived time.Time
}

func (m *MessageBase) GetReceivedTime() time.Time {
	return m.LocalReceived
}

func (m *MessageBase) SetReceivedTime(time time.Time) {
	m.LocalReceived = time
}

func (m *MessageBase) StringOfMsgBase() string {

	rval := fmt.Sprintf("origin %s(%d), LChain=%x resendCnt=%d", m.NetworkOrigin, m.Origin, m.LeaderChainID.Bytes()[3:6], m.ResendCnt)
	if m.LocalOnly {
		rval += " local"
	}
	if m.Peer2Peer {
		rval += " p2p"
	}
	if m.FullBroadcast {
		rval += " FullBroadcast"
	}
	if m.NoResend {
		rval += "noResend"
	}
	if m.MarkInvalid {
		rval += "MarkInvalid"
	}
	if m.Stalled {
		rval += "Stalled"
	}

	return rval
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
		s.LogPrintf(logname, "Duplicate Send of R-%x (%d sends, %d duplicates, %d unique)", msg.GetRepeatHash().Bytes()[:3], sends, duplicate, unique)
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
	if msg.GetRepeatHash() == nil || reflect.ValueOf(msg.GetRepeatHash()).IsNil() || msg.GetMsgHash() == nil || reflect.ValueOf(msg.GetMsgHash()).IsNil() { // Do not send pokemon messages
		return
	}

	if msg.GetNoResend() {
		return
	}
	// Local Messages are NOT broadcast out.  This is mostly the block signature
	// generated by the timer for the leaders which needs to be processed, but replaced
	// by an updated version when the block is ready.
	if msg.IsLocal() {
		return
	}

	// Dont resend until its time.  If resend==0 send immediately (this is the first time), or if it has been long enough.
	now := s.GetTimestamp()
	if m.resend != 0 && m.resend > now.GetTimeMilli() {
		return
	}

	// We only send at a slow rate, but keep doing it because in slow networks, we are pushing the message to the leader
	if m.ResendCnt > 3 { // If the first send fails, we need to try again.  Give up eventually.
		return
	}

	m.ResendCnt++
	sends++

	// Send once every so often.
	m.resend = now.GetTimeMilli() + 1*1000

	// debug code start ............
	if !msg.IsPeer2Peer() && s.DebugExec() && s.CheckFileName(logname) { // if debug is on and this logfile is enabled
		//		checkForDuplicateSend(s, msg, atomic.WhereAmIString(1)) // turn off duplicate send check for now
	}
	// debug code end ............
	s.LogMessage("NetworkOutputs", "Enqueue", msg)

	if s.IsRunLeader() { // true means - we are not in wait period
		s.NetworkOutMsgQueue().Enqueue(msg)
	} else {
		q := s.NetworkOutMsgQueue()
		if q.Length() < q.Cap() {
			q.Enqueue(msg)
		} else {
			popped := s.NetworkOutMsgQueue().BlockingDequeue()
			s.LogMessage("NetworkOutputs", "Popped & dropped", popped)
			q.Enqueue(msg)
		}
	}

	// Add this to the network replay filter so we don't bother processing any echos
	s.AddToReplayFilter(constants.NETWORK_REPLAY, msg.GetRepeatHash().Fixed(), msg.GetTimestamp(), now)
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

func (m *MessageBase) GetFullMsgHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "MessageBase.GetFullMsgHash") }()

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

func (m *MessageBase) IsNetwork() bool {
	return m.Network
}

func (m *MessageBase) SetNetwork(v bool) {
	m.Network = v
}

func (m *MessageBase) IsFullBroadcast() bool {
	return m.FullBroadcast
}

func (m *MessageBase) SetFullBroadcast(v bool) {
	m.FullBroadcast = v
}
func (m *MessageBase) GetLeaderChainID() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "MessageBase.GetLeaderChainID") }()

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

func (m *MessageBase) InvalidateSignatures() {
	m.Sigvalid = false
}
