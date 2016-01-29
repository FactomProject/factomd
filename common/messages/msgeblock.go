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

// MsgEBlock implements the MsgEBlock interface and represents a factom
// EBlock message.  It is used by client to download the EBlock.
type MsgEBlock struct {
	EBlk *EBlock
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the MsgEBlock interface implementation.
func (msg *MsgEBlock) BtcEncode(w io.Writer, pver uint32) error {

	bytes, err := msg.EBlk.MarshalBinary()
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
// This is part of the MsgEBlock interface implementation.
func (msg *MsgEBlock) BtcDecode(r io.Reader, pver uint32) error {

	bytes, err := readVarBytes(r, pver, uint32(MaxBlockMsgPayload), CmdEBlock)
	if err != nil {
		return err
	}

	msg.EBlk = NewEBlock()
	err = msg.EBlk.UnmarshalBinary(bytes)
	if err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the MsgEBlock interface implementation.
func (msg *MsgEBlock) Command() string {
	return CmdEBlock
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the MsgEBlock interface implementation.
func (msg *MsgEBlock) MaxPayloadLength(pver uint32) uint32 {
	return MaxBlockMsgPayload
}

// NewMsgEBlock returns a new bitcoin inv message that conforms to the MsgEBlock
// interface.  See MsgInv for details.
func NewMsgEBlock() *MsgEBlock {
	return &MsgEBlock{}
}

var _ interfaces.IMsg = (*MsgEBlock)(nil)

func (m *MsgEBlock) Process(uint32, interfaces.IState) {}

func (m *MsgEBlock) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgEBlock) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgEBlock) Type() int {
	return -1
}

func (m *MsgEBlock) Int() int {
	return -1
}

func (m *MsgEBlock) Bytes() []byte {
	return nil
}

func (m *MsgEBlock) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgEBlock) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgEBlock) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgEBlock) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgEBlock) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgEBlock is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgEBlock is valid
func (m *MsgEBlock) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgEBlock) Leader(state interfaces.IState) bool {
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
func (m *MsgEBlock) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgEBlock) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgEBlock) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgEBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgEBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgEBlock) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
