// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

type MessageBase struct {
	Origin    int  // Set and examined on a server, not marshaled with the message
	Peer2peer bool // The nature of this message type, not marshaled with the message
	LocalOnly bool // This message is only a local message, is not broadcasted and may skip verification

	LeaderChainID interfaces.IHash
	MsgHash       interfaces.IHash // Cash of the hash of a message
	VMIndex       int              // The Index of the VM responsible for this message.
	VMHash        []byte	       // Basis for selecting a VMIndex
	// Used by Leader code, but only Marshaled and Unmarshalled in Ack Messages
	// EOM messages, and DirectoryBlockSignature messages
}

func (m *MessageBase) GetOrigin() int {
	return m.Origin
}

func (m *MessageBase) SetOrigin(o int) {
	m.Origin = o
}

// Returns true if this is a response to a peer to peer
// request.
func (m *MessageBase) IsPeer2peer() bool {
	return m.Peer2peer
}

func (m *MessageBase) IsLocal() bool {
	return m.LocalOnly
}

func (m *MessageBase) SetLocal(v bool) {
	m.LocalOnly = v
}

func (m *MessageBase) GetLeaderChainID() interfaces.IHash {
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

func (m *MessageBase) GetVMHash() ([]byte) {
	return m.VMHash
}

func (m *MessageBase) SetVMHash(vmhash []byte) {
	m.VMHash = vmhash
}
