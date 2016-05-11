// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type MessageBase struct {
	Origin    			int  // Set and examined on a server, not marshaled with the message
	Peer2Peer 			bool // The nature of this message type, not marshaled with the message
	LocalOnly 			bool // This message is only a local message, is not broadcasted and may skip verification

	Salt					interfaces.Timestamp	// Might be used to get past duplicate protection when messages are missing

	Stalled				bool						// Messages marked as stalled do not get transmitted out on the network.
	LeaderChainID 		interfaces.IHash
	MsgHash				interfaces.IHash 		// Cash of the hash of a message
	VMIndex				int              		// The Index of the VM responsible for this message.
	VMHash   			[]byte           		// Basis for selecting a VMIndex
	Minute            byte
	// Used by Leader code, but only Marshaled and Unmarshalled in Ack Messages
	// EOM messages, and DirectoryBlockSignature messages
}

func (m *MessageBase) GetStalled() bool {
	return m.Stalled
}

func (m *MessageBase) SetStalled(stalled bool) {
	m.Stalled = stalled
}


func (m *MessageBase) SaltReply(state interfaces.IState) {
	m.Salt = state.GetTimestamp()
}

func (m *MessageBase) GetOrigin() int {
	return m.Origin
}

func (m *MessageBase) SetOrigin(o int) {
	m.Origin = o
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