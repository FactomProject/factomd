// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"io"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// MsgABlock implements the Message interface and represents a factom
// Admin Block message.  It is used by client to download Admin Block.
type MsgABlock struct {
	ABlk *AdminBlock
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgABlock) BtcEncode(w io.Writer, pver uint32) error {

	bytes, err := msg.ABlk.MarshalBinary()
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
// This is part of the Message interface implementation.
func (msg *MsgABlock) BtcDecode(r io.Reader, pver uint32) error {

	bytes, err := readVarBytes(r, pver, uint32(MaxBlockMsgPayload), CmdABlock)
	if err != nil {
		return err
	}

	msg.ABlk = new(AdminBlock)
	err = msg.ABlk.UnmarshalBinary(bytes)
	if err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgABlock) Command() string {
	return CmdABlock
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgABlock) MaxPayloadLength(pver uint32) uint32 {
	return MaxBlockMsgPayload
}

// NewMsgABlock returns a new bitcoin inv message that conforms to the Message
// interface.  See MsgInv for details.
func NewMsgABlock() *MsgABlock {
	return &MsgABlock{}
}

var _ interfaces.IMsg = (*MsgABlock)(nil)

func (m *MsgABlock) Process(uint32, interfaces.IState) {}

func (m *MsgABlock) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgABlock) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgABlock) Type() int {
	return -1
}

func (m *MsgABlock) Int() int {
	return -1
}

func (m *MsgABlock) Bytes() []byte {
	return nil
}

func (m *MsgABlock) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgABlock) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgABlock) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgABlock) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgABlock) String() string {
	return ""
}

func (m *MsgABlock) DBHeight() int {
	return 0
}

func (m *MsgABlock) ChainID() []byte {
	return nil
}

func (m *MsgABlock) ListHeight() int {
	return 0
}

func (m *MsgABlock) SerialHash() []byte {
	return nil
}

func (m *MsgABlock) Signature() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *MsgABlock) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgABlock) Leader(state interfaces.IState) bool {
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
func (m *MsgABlock) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgABlock) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgABlock) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgABlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgABlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgABlock) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
