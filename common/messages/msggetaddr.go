// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"io"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// MsgGetAddr implements the MsgGetAddr interface and represents a bitcoin
// getaddr message.  It is used to request a list of known active peers on the
// network from a peer to help identify potential nodes.  The list is returned
// via one or more addr messages (MsgAddr).
//
// This message has no payload.
type MsgGetAddr struct{}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the MsgGetAddr interface implementation.
func (msg *MsgGetAddr) BtcDecode(r io.Reader, pver uint32) error {
	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the MsgGetAddr interface implementation.
func (msg *MsgGetAddr) BtcEncode(w io.Writer, pver uint32) error {
	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the MsgGetAddr interface implementation.
func (msg *MsgGetAddr) Command() string {
	return CmdGetAddr
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the MsgGetAddr interface implementation.
func (msg *MsgGetAddr) MaxPayloadLength(pver uint32) uint32 {
	return 0
}

// NewMsgGetAddr returns a new bitcoin getaddr message that conforms to the
// MsgGetAddr interface.  See MsgGetAddr for details.
func NewMsgGetAddr() *MsgGetAddr {
	return &MsgGetAddr{}
}

var _ interfaces.IMsg = (*MsgGetAddr)(nil)

func (m *MsgGetAddr) Process(uint32, interfaces.IState) {}

func (m *MsgGetAddr) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgGetAddr) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgGetAddr) Type() int {
	return -1
}

func (m *MsgGetAddr) Int() int {
	return -1
}

func (m *MsgGetAddr) Bytes() []byte {
	return nil
}

func (m *MsgGetAddr) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgGetAddr) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgGetAddr) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgGetAddr) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgGetAddr) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgGetAddr is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgGetAddr is valid
func (m *MsgGetAddr) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgGetAddr) Leader(state interfaces.IState) bool {
	switch state.GetNetworkNumber() {
	case 0: // Main Network
		panic("Not implemented yet")
	case 1: // Test Network
		panic("Not implemented yet")
	case 2: // Local Network
		panic("Not implemented yet")
	default:
		panic("Not implemented yet")
	}

}

// Execute the leader functions of the given message
func (m *MsgGetAddr) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgGetAddr) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgGetAddr) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgGetAddr) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgGetAddr) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgGetAddr) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
