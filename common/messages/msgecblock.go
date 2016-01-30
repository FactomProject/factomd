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

// MsgECBlock implements the MsgECBlock interface and represents a factom ECBlock
// message.  It is used by client to download ECBlock.
type MsgECBlock struct {
	ECBlock *ECBlock
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the MsgECBlock interface implementation.
func (msg *MsgECBlock) BtcEncode(w io.Writer, pver uint32) error {
	bytes, err := msg.ECBlock.MarshalBinary()
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
// This is part of the MsgECBlock interface implementation.
func (msg *MsgECBlock) BtcDecode(r io.Reader, pver uint32) error {

	bytes, err := readVarBytes(r, pver, uint32(MaxBlockMsgPayload), CmdECBlock)
	if err != nil {
		return err
	}

	msg.ECBlock = new(ECBlock)
	err = msg.ECBlock.UnmarshalBinary(bytes)
	if err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the MsgECBlock interface implementation.
func (msg *MsgECBlock) Command() string {
	return CmdECBlock
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the MsgECBlock interface implementation.
func (msg *MsgECBlock) MaxPayloadLength(pver uint32) uint32 {
	return MaxBlockMsgPayload
}

// NewMsgECBlock returns a new bitcoin inv message that conforms to the MsgECBlock
// interface.  See MsgInv for details.
func NewMsgECBlock() *MsgECBlock {
	return &MsgECBlock{}
}

var _ interfaces.IMsg = (*MsgECBlock)(nil)

func (m *MsgECBlock) Process(uint32, interfaces.IState) {}

func (m *MsgECBlock) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgECBlock) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgECBlock) Type() int {
	return -1
}

func (m *MsgECBlock) Int() int {
	return -1
}

func (m *MsgECBlock) Bytes() []byte {
	return nil
}

func (m *MsgECBlock) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgECBlock) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgECBlock) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgECBlock) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgECBlock) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgECBlock is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgECBlock is valid
func (m *MsgECBlock) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgECBlock) Leader(state interfaces.IState) bool {
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
func (m *MsgECBlock) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgECBlock) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgECBlock) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgECBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgECBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgECBlock) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
