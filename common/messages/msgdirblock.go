// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"io"

	. "github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// MaxBlocksPerMsg is the maximum number of blocks allowed per message.
const MaxBlocksPerMsg = 500

// MaxBlockPayload is the maximum bytes a block message can be in bytes.
const MaxBlockPayload = 1000000 // Not actually 1MB which would be 1024 * 1024

// MsgDirBlock implements the MsgDirBlock interface and represents a factom
// DBlock message.  It is used by client to reveal the entry.
type MsgDirBlock struct {
	DBlk *DirectoryBlock
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the MsgDirBlock interface implementation.
func (msg *MsgDirBlock) BtcEncode(w io.Writer, pver uint32) error {

	bytes, err := msg.DBlk.MarshalBinary()
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
// This is part of the MsgDirBlock interface implementation.
func (msg *MsgDirBlock) BtcDecode(r io.Reader, pver uint32) error {
	//Entry
	bytes, err := readVarBytes(r, pver, uint32(MaxBlockMsgPayload), CmdRevealEntry)
	if err != nil {
		return err
	}

	msg.DBlk = new(DirectoryBlock)
	err = msg.DBlk.UnmarshalBinary(bytes)
	if err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the MsgDirBlock interface implementation.
func (msg *MsgDirBlock) Command() string {
	return CmdDirBlock
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the MsgDirBlock interface implementation.
func (msg *MsgDirBlock) MaxPayloadLength(pver uint32) uint32 {
	return MaxBlockMsgPayload
}

// NewMsgDirBlock returns a new bitcoin inv message that conforms to the MsgDirBlock
// interface.  See MsgInv for details.
func NewMsgDirBlock() *MsgDirBlock {
	return &MsgDirBlock{}
}

var _ interfaces.IMsg = (*MsgDirBlock)(nil)

func (m *MsgDirBlock) Process(dbheight uint32, state interfaces.IState) {
	//	Code to process this block
}

func (m *MsgDirBlock) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgDirBlock) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgDirBlock) Type() int {
	return -1
}

func (m *MsgDirBlock) Int() int {
	return -1
}

func (m *MsgDirBlock) Bytes() []byte {
	return nil
}

func (m *MsgDirBlock) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgDirBlock) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgDirBlock) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgDirBlock) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgDirBlock) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgDirBlock is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgDirBlock is valid
func (m *MsgDirBlock) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgDirBlock) Leader(state interfaces.IState) bool {
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
func (m *MsgDirBlock) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgDirBlock) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgDirBlock) FollowerExecute(state interfaces.IState) error {
	return nil
}

func (e *MsgDirBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgDirBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgDirBlock) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
