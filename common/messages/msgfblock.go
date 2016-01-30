// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"io"

	"github.com/FactomProject/factomd/common/factoid/block"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// factoid block
type MsgFBlock struct {
	FBlck interfaces.IFBlock
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the MsgFBlock interface implementation.
func (msg *MsgFBlock) BtcEncode(w io.Writer, pver uint32) error {

	bytes, err := msg.FBlck.MarshalBinary()
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
// This is part of the MsgFBlock interface implementation.
func (msg *MsgFBlock) BtcDecode(r io.Reader, pver uint32) error {

	bytes, err := readVarBytes(r, pver, uint32(MaxBlockMsgPayload), CmdFBlock)
	if err != nil {
		return err
	}

	msg.FBlck = new(block.FBlock)
	err = msg.FBlck.UnmarshalBinary(bytes)
	if err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the MsgFBlock interface implementation.
func (msg *MsgFBlock) Command() string {
	return CmdFBlock
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the MsgFBlock interface implementation.
func (msg *MsgFBlock) MaxPayloadLength(pver uint32) uint32 {
	return MaxBlockMsgPayload
}

// NewMsgABlock returns a new bitcoin inv message that conforms to the MsgFBlock
// interface.  See MsgInv for details.
func NewMsgFBlock() *MsgFBlock {
	return &MsgFBlock{}
}

var _ interfaces.IMsg = (*MsgFBlock)(nil)

func (m *MsgFBlock) Process(uint32, interfaces.IState) {}

func (m *MsgFBlock) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgFBlock) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgFBlock) Type() int {
	return -1
}

func (m *MsgFBlock) Int() int {
	return -1
}

func (m *MsgFBlock) Bytes() []byte {
	return nil
}

func (m *MsgFBlock) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgFBlock) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgFBlock) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgFBlock) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgFBlock) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgFBlock is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgFBlock is valid
func (m *MsgFBlock) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgFBlock) Leader(state interfaces.IState) bool {
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
func (m *MsgFBlock) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgFBlock) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgFBlock) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgFBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgFBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgFBlock) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
