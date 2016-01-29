// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"io"
)

// MsgVerAck defines a bitcoin verack message which is used for a peer to
// acknowledge a version message (MsgVersion) after it has used the information
// to negotiate parameters.  It implements the Message interface.
//
// This message has no payload.
type MsgVerAck struct{}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgVerAck) BtcDecode(r io.Reader, pver uint32) error {
	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgVerAck) BtcEncode(w io.Writer, pver uint32) error {
	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgVerAck) Command() string {
	return CmdVerAck
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgVerAck) MaxPayloadLength(pver uint32) uint32 {
	return 0
}

// NewMsgVerAck returns a new bitcoin verack message that conforms to the
// Message interface.
func NewMsgVerAck() *MsgVerAck {
	return &MsgVerAck{}
}

var _ interfaces.IMsg = (*MsgVerAck)(nil)

func (m *MsgVerAck) Process(uint32, interfaces.IState) {}

func (m *MsgVerAck) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgVerAck) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgVerAck) Type() int {
	return -1
}

func (m *MsgVerAck) Int() int {
	return -1
}

func (m *MsgVerAck) Bytes() []byte {
	return nil
}

func (m *MsgVerAck) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgVerAck) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgVerAck) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgVerAck) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgVerAck) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgVerAck is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgVerAck is valid
func (m *MsgVerAck) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgVerAck) Leader(state interfaces.IState) bool {
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
func (m *MsgVerAck) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgVerAck) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgVerAck) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgVerAck) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgVerAck) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgVerAck) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
