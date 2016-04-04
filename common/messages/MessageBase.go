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

	// Cash of the hash of a message
	MsgHash interfaces.IHash
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
