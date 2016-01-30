// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"encoding/binary"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"io"
)

// MsgPong implements the Message interface and represents a bitcoin pong
// message which is used primarily to confirm that a connection is still valid
// in response to a bitcoin ping message (MsgPong).
type MsgPong struct {
	// Unique value associated with message that is used to identify
	// specific ping message.
	Nonce uint64
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgPong) BtcDecode(r io.Reader, pver uint32) error {
	err := readElement(r, &msg.Nonce)
	if err != nil {
		return err
	}

	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgPong) BtcEncode(w io.Writer, pver uint32) error {
	err := writeElement(w, msg.Nonce)
	if err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgPong) Command() string {
	return CmdPong
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgPong) MaxPayloadLength(pver uint32) uint32 {
	plen := uint32(0)
	return plen
}

// NewMsgPong returns a new bitcoin pong message that conforms to the Message
// interface.  See MsgPong for details.
func NewMsgPong(nonce uint64) *MsgPong {
	return &MsgPong{
		Nonce: nonce,
	}
}

var _ interfaces.IMsg = (*MsgPong)(nil)

func (m *MsgPong) Process(uint32, interfaces.IState) {}

func (m *MsgPong) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgPong) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgPong) Type() int {
	return -1
}

func (m *MsgPong) Int() int {
	return -1
}

func (m *MsgPong) Bytes() []byte {
	return nil
}

func (m *MsgPong) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	m.Nonce, newdata = binary.BigEndian.Uint64(data[0:8]), data[8:]
	return newdata, nil
}

func (m *MsgPong) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgPong) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgPong) MarshalForSignature() (data []byte, err error) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, m.Nonce)
	return buf.Bytes(), nil
}

func (m *MsgPong) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgPong is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgPong is valid
func (m *MsgPong) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgPong) Leader(state interfaces.IState) bool {
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
func (m *MsgPong) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgPong) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgPong) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgPong) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgPong) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgPong) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
