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

// MsgPing implements the Message interface and represents a bitcoin ping
// message.
//
// The payload for this message just consists of a nonce used for identifying
// it later.
type MsgPing struct {
	// Unique value associated with message that is used to identify
	// specific ping message.
	Nonce uint64
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgPing) BtcDecode(r io.Reader, pver uint32) error {
	err := readElement(r, &msg.Nonce)
	if err != nil {
		return err
	}

	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgPing) BtcEncode(w io.Writer, pver uint32) error {
	err := writeElement(w, msg.Nonce)
	if err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgPing) Command() string {
	return CmdPing
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgPing) MaxPayloadLength(pver uint32) uint32 {
	plen := uint32(0)
	return plen
}

// NewMsgPing returns a new bitcoin ping message that conforms to the Message
// interface.  See MsgPing for details.
func NewMsgPing(nonce uint64) *MsgPing {
	return &MsgPing{
		Nonce: nonce,
	}
}

var _ interfaces.IMsg = (*MsgPing)(nil)

func (m *MsgPing) Process(uint32, interfaces.IState) {}

func (m *MsgPing) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgPing) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgPing) Type() int {
	return -1
}

func (m *MsgPing) Int() int {
	return -1
}

func (m *MsgPing) Bytes() []byte {
	return nil
}

func (m *MsgPing) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	m.Nonce, newdata = binary.BigEndian.Uint64(data[0:8]), data[8:]
	return newdata, nil
}

func (m *MsgPing) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgPing) MarshalBinary() (data []byte, err error) {
	var pver uint32
	buf := bytes.NewBuffer(make([]byte, 0, m.MaxPayloadLength(pver)))
	binary.Write(buf, binary.BigEndian, m.Nonce)
	return buf.Bytes(), nil
}

func (m *MsgPing) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgPing) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgPing is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgPing is valid
func (m *MsgPing) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgPing) Leader(state interfaces.IState) bool {
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
func (m *MsgPing) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgPing) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgPing) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgPing) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgPing) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgPing) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
