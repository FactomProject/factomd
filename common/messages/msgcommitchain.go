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

// MsgCommitChain implements the MsgCommitChain interface and represents a factom
// Commit-Chain message.  It is used by client to commit the chain before revealing it.
type MsgCommitChain struct {
	CommitChain *CommitChain
}

// NewMsgCommitChain returns a new Commit Chain message that conforms to the
// MsgCommitChain interface.  See MsgInv for details.
func NewMsgCommitChain() *MsgCommitChain {
	m := new(MsgCommitChain)
	m.CommitChain = new(CommitChain)
	return m
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the MsgCommitChain interface implementation.
func (msg *MsgCommitChain) BtcEncode(w io.Writer, pver uint32) error {
	bytes, err := msg.CommitChain.MarshalBinary()
	if err != nil {
		return err
	}

	if err := writeVarBytes(w, pver, bytes); err != nil {
		return err
	}

	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the MsgCommitChain interface implementation.
func (msg *MsgCommitChain) BtcDecode(r io.Reader, pver uint32) error {
	bytes, err := readVarBytes(r, pver, uint32(CommitChainSize+8),
		CmdEntry)
	if err != nil {
		return err
	}

	msg.CommitChain = new(CommitChain)
	if err := msg.CommitChain.UnmarshalBinary(bytes); err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the MsgCommitChain interface implementation.
func (msg *MsgCommitChain) Command() string {
	return CmdCommitChain
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the MsgCommitChain interface implementation.
func (msg *MsgCommitChain) MaxPayloadLength(pver uint32) uint32 {
	return MaxAppMsgPayload
}

// Check whether the msg can pass the message level validations
// such as timestamp, signiture and etc
func (msg *MsgCommitChain) IsValid() bool {
	//return msg.CommitChain.IsValid()
	return true
}

// Create a sha hash from the message binary (output of BtcEncode)
func (msg *MsgCommitChain) Sha() (interfaces.IHash, error) {

	buf := bytes.NewBuffer(nil)
	msg.BtcEncode(buf, ProtocolVersion)
	var sha interfaces.IHash
	_ = sha.SetBytes(Sha256(buf.Bytes()))

	return sha, nil
}

var _ interfaces.IMsg = (*MsgCommitChain)(nil)

func (m *MsgCommitChain) Process(uint32, interfaces.IState) {}

func (m *MsgCommitChain) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgCommitChain) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgCommitChain) Type() int {
	return -1
}

func (m *MsgCommitChain) Int() int {
	return -1
}

func (m *MsgCommitChain) Bytes() []byte {
	return nil
}

func (m *MsgCommitChain) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgCommitChain) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgCommitChain) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgCommitChain) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgCommitChain) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgCommitChain is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgCommitChain is valid
func (m *MsgCommitChain) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgCommitChain) Leader(state interfaces.IState) bool {
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
func (m *MsgCommitChain) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgCommitChain) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgCommitChain) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgCommitChain) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgCommitChain) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgCommitChain) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
