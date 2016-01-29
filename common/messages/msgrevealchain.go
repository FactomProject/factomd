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

// MsgRevealChain implements the MsgRevealChain interface and represents a factom
// Reveal-Chain message.  It is used by client to reveal the chain.
type MsgRevealChain struct {
	FirstEntry interfaces.IEBEntry
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the MsgRevealChain interface implementation.
func (msg *MsgRevealChain) BtcEncode(w io.Writer, pver uint32) error {

	//FirstEntry
	bytes, err := msg.FirstEntry.MarshalBinary()
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
// This is part of the MsgRevealChain interface implementation.
func (msg *MsgRevealChain) BtcDecode(r io.Reader, pver uint32) error {
	//FirstEntry
	bytes, err := readVarBytes(r, pver, MaxAppMsgPayload, CmdRevealChain)
	if err != nil {
		return err
	}

	msg.FirstEntry = new(Entry)
	err = msg.FirstEntry.UnmarshalBinary(bytes)
	if err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the MsgRevealChain interface implementation.
func (msg *MsgRevealChain) Command() string {
	return CmdRevealChain
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the MsgRevealChain interface implementation.
func (msg *MsgRevealChain) MaxPayloadLength(pver uint32) uint32 {
	return MaxAppMsgPayload
}

// NewMsgInv returns a new bitcoin inv message that conforms to the MsgRevealChain
// interface.  See MsgInv for details.
func NewMsgRevealChain() *MsgRevealChain {
	return &MsgRevealChain{}
}

// Create a sha hash from the message binary (output of BtcEncode)
func (msg *MsgRevealChain) Sha() (interfaces.IHash, error) {

	buf := bytes.NewBuffer(nil)
	msg.BtcEncode(buf, ProtocolVersion)
	var sha interfaces.IHash
	_ = sha.SetBytes(Sha256(buf.Bytes()))

	return sha, nil
}

var _ interfaces.IMsg = (*MsgRevealChain)(nil)

func (m *MsgRevealChain) Process(uint32, interfaces.IState) {}

func (m *MsgRevealChain) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgRevealChain) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgRevealChain) Type() int {
	return -1
}

func (m *MsgRevealChain) Int() int {
	return -1
}

func (m *MsgRevealChain) Bytes() []byte {
	return nil
}

func (m *MsgRevealChain) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgRevealChain) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgRevealChain) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgRevealChain) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgRevealChain) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgRevealChain is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgRevealChain is valid
func (m *MsgRevealChain) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgRevealChain) Leader(state interfaces.IState) bool {
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
func (m *MsgRevealChain) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgRevealChain) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgRevealChain) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgRevealChain) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgRevealChain) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgRevealChain) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
