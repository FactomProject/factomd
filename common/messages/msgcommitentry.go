// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"io"

	. "github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// MsgCommitEntry implements the MsgCommitEntry interface and represents a factom
// Commit-Entry message.  It is used by client to commit the entry before
// revealing it.
type MsgCommitEntry struct {
	CommitEntry *CommitEntry
}

// NewMsgCommitEntry returns a new Commit Entry message that conforms to the
// MsgCommitEntry interface.
func NewMsgCommitEntry() *MsgCommitEntry {
	m := new(MsgCommitEntry)
	m.CommitEntry = NewCommitEntry()
	return m
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the MsgCommitEntry interface implementation.
func (msg *MsgCommitEntry) BtcEncode(w io.Writer, pver uint32) error {
	bytes, err := msg.CommitEntry.MarshalBinary()
	if err != nil {
		return err
	}

	if err := writeVarBytes(w, pver, bytes); err != nil {
		return err
	}

	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the MsgCommitEntry interface implementation.
func (msg *MsgCommitEntry) BtcDecode(r io.Reader, pver uint32) error {
	bytes, err := readVarBytes(r, pver, uint32(CommitEntrySize),
		CmdEntry)
	if err != nil {
		return err
	}

	msg.CommitEntry = NewCommitEntry()
	err = msg.CommitEntry.UnmarshalBinary(bytes)
	if err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the MsgCommitEntry interface implementation.
func (msg *MsgCommitEntry) Command() string {
	return CmdCommitEntry
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the MsgCommitEntry interface implementation.
func (msg *MsgCommitEntry) MaxPayloadLength(pver uint32) uint32 {
	return MaxAppMsgPayload
}

// Check whether the msg can pass the message level validations
// such as timestamp, signiture and etc
func (msg *MsgCommitEntry) IsValid() bool {
	return msg.CommitEntry.IsValid()
}

// Create a sha hash from the message binary (output of BtcEncode)
func (msg *MsgCommitEntry) Sha() (interfaces.IHash, error) {

	buf := bytes.NewBuffer(nil)
	msg.BtcEncode(buf, ProtocolVersion)
	var sha interfaces.IHash
	_ = sha.SetBytes(Sha256(buf.Bytes()))

	return sha, nil
}

var _ interfaces.IMsg = (*MsgCommitEntry)(nil)

func (m *MsgCommitEntry) Process(uint32, interfaces.IState) {}

func (m *MsgCommitEntry) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgCommitEntry) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgCommitEntry) Type() int {
	return -1
}

func (m *MsgCommitEntry) Int() int {
	return -1
}

func (m *MsgCommitEntry) Bytes() []byte {
	return nil
}

func (m *MsgCommitEntry) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgCommitEntry) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgCommitEntry) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgCommitEntry) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgCommitEntry) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgCommitEntry is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgCommitEntry is valid
func (m *MsgCommitEntry) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgCommitEntry) Leader(state interfaces.IState) bool {
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
func (m *MsgCommitEntry) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgCommitEntry) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgCommitEntry) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgCommitEntry) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgCommitEntry) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgCommitEntry) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
