// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"io"

	. "github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// MsgRevealEntry implements the MsgRevealEntry interface and represents a factom
// Reveal-Entry message.  It is used by client to reveal the entry.
type MsgRevealEntry struct {
	Entry interfaces.IEBEntry
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the MsgRevealEntry interface implementation.
func (msg *MsgRevealEntry) BtcEncode(w io.Writer, pver uint32) error {

	//Entry
	bytes, err := msg.Entry.MarshalBinary()
	if err != nil {
		return err
	}

	err = writeVarBytes(w, pver, bytes)
	if err != nil {
		return err
	}

	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the MsgRevealEntry interface implementation.
func (msg *MsgRevealEntry) BtcDecode(r io.Reader, pver uint32) error {
	//Entry
	bytes, err := readVarBytes(r, pver, uint32(MaxAppMsgPayload), CmdRevealEntry)
	if err != nil {
		return err
	}

	msg.Entry = new(Entry)
	err = msg.Entry.UnmarshalBinary(bytes)
	if err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the MsgRevealEntry interface implementation.
func (msg *MsgRevealEntry) Command() string {
	return CmdRevealEntry
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the MsgRevealEntry interface implementation.
func (msg *MsgRevealEntry) MaxPayloadLength(pver uint32) uint32 {
	return MaxAppMsgPayload
}

// NewMsgInv returns a new bitcoin inv message that conforms to the MsgRevealEntry
// interface.  See MsgInv for details.
func NewMsgRevealEntry() *MsgRevealEntry {
	return &MsgRevealEntry{}
}

// Create a sha hash from the message binary (output of BtcEncode)
func (msg *MsgRevealEntry) Sha() (interfaces.IHash, error) {

	buf := bytes.NewBuffer(nil)
	msg.BtcEncode(buf, ProtocolVersion)
	var sha interfaces.IHash
	_ = sha.SetBytes(Sha256(buf.Bytes()))

	return sha, nil
}

// Check whether the msg can pass the message level validations
func (msg *MsgRevealEntry) IsValid() bool {
	return true
	//return msg.Entry.IsValid()
}

var _ interfaces.IMsg = (*MsgRevealEntry)(nil)

func (m *MsgRevealEntry) Process(uint32, interfaces.IState) {}

func (m *MsgRevealEntry) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgRevealEntry) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgRevealEntry) Type() int {
	return -1
}

func (m *MsgRevealEntry) Int() int {
	return -1
}

func (m *MsgRevealEntry) Bytes() []byte {
	return nil
}

func (m *MsgRevealEntry) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgRevealEntry) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgRevealEntry) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgRevealEntry) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgRevealEntry) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgRevealEntry is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgRevealEntry is valid
func (m *MsgRevealEntry) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgRevealEntry) Leader(state interfaces.IState) bool {
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
func (m *MsgRevealEntry) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgRevealEntry) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgRevealEntry) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgRevealEntry) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgRevealEntry) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgRevealEntry) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
