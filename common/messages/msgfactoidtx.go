// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"fmt"
	"io"

	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var _ = fmt.Printf

type IMsgFactoidTX interface {
	// Set the Transaction to be carried by this message.
	SetTransaction(interfaces.ITransaction)
	// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
	// This is part of the MsgFactoidTX interface implementation.
	BtcEncode(w io.Writer, pver uint32) error
	// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
	// This is part of the MsgFactoidTX interface implementation.
	BtcDecode(r io.Reader, pver uint32) error
	// Command returns the protocol command string for the message.  This is part
	// of the MsgFactoidTX interface implementation.
	Command() string
	// MaxPayloadLength returns the maximum length the payload can be for the
	// receiver.  This is part of the MsgFactoidTX interface implementation.
	MaxPayloadLength(pver uint32) uint32
	// NewMsgCommitEntry returns a new bitcoin Commit Entry message that conforms to
	// the MsgFactoidTX interface.
	NewMsgFactoidTX() IMsgFactoidTX
	// Check whether the msg can pass the message level validations
	// such as timestamp, signiture and etc
	IsValid() bool
	// Create a sha hash from the message binary (output of BtcEncode)
	Sha() (interfaces.IHash, error)
}

// MsgCommitEntry implements the MsgFactoidTX interface and represents a factom
// Commit-Entry message.  It is used by client to commit the entry before
// revealing it.
type MsgFactoidTX struct {
	IMsgFactoidTX
	Transaction interfaces.ITransaction
}

// Accessor to set the Transaction for a message.
func (msg *MsgFactoidTX) SetTransaction(transaction interfaces.ITransaction) {
	msg.Transaction = transaction
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the MsgFactoidTX interface implementation.
func (msg *MsgFactoidTX) BtcEncode(w io.Writer, pver uint32) error {

	data, err := msg.Transaction.MarshalBinary()
	if err != nil {
		return err
	}

	err = writeVarBytes(w, pver, data)
	if err != nil {
		return err
	}

	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the MsgFactoidTX interface implementation.
func (msg *MsgFactoidTX) BtcDecode(r io.Reader, pver uint32) error {

	data, err := readVarBytes(r, pver, uint32(MaxAppMsgPayload), CmdEBlock)
	if err != nil {
		return err
	}

	msg.Transaction = new(Transaction)
	err = msg.Transaction.UnmarshalBinary(data)
	if err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the MsgFactoidTX interface implementation.
func (msg *MsgFactoidTX) Command() string {
	return CmdFactoidTX
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the MsgFactoidTX interface implementation.
func (msg *MsgFactoidTX) MaxPayloadLength(pver uint32) uint32 {
	return MaxAppMsgPayload
}

// NewMsgCommitEntry returns a new bitcoin Commit Entry message that conforms to
// the MsgFactoidTX interface.
func NewMsgFactoidTX() IMsgFactoidTX {
	return &MsgFactoidTX{}
}

// Check whether the msg can pass the message level validations
// such as timestamp, signiture and etc
func (msg *MsgFactoidTX) IsValid() bool {
	err := msg.Transaction.Validate(1)
	if err != nil {
		return false
	}
	err = msg.Transaction.ValidateSignatures()
	if err != nil {
		return false
	}
	return true
}

// Create a sha hash from the message binary (output of BtcEncode)
func (msg *MsgFactoidTX) Sha() (interfaces.IHash, error) {

	buf := bytes.NewBuffer(nil)
	msg.BtcEncode(buf, ProtocolVersion)
	var sha interfaces.IHash
	_ = sha.SetBytes(Sha256(buf.Bytes()))

	return sha, nil
}

var _ interfaces.IMsg = (*MsgFactoidTX)(nil)

func (m *MsgFactoidTX) Process(uint32, interfaces.IState) {}

func (m *MsgFactoidTX) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgFactoidTX) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgFactoidTX) Type() int {
	return -1
}

func (m *MsgFactoidTX) Int() int {
	return -1
}

func (m *MsgFactoidTX) Bytes() []byte {
	return nil
}

func (m *MsgFactoidTX) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgFactoidTX) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgFactoidTX) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgFactoidTX) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgFactoidTX) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgFactoidTX is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgFactoidTX is valid
func (m *MsgFactoidTX) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgFactoidTX) Leader(state interfaces.IState) bool {
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
func (m *MsgFactoidTX) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgFactoidTX) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgFactoidTX) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgFactoidTX) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgFactoidTX) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgFactoidTX) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
