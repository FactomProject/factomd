// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// Acknowledgement Type
const (
	ACK_FACTOID_TX uint8 = iota
	END_MINUTE_1
	END_MINUTE_2
	END_MINUTE_3
	END_MINUTE_4
	END_MINUTE_5
	END_MINUTE_6
	END_MINUTE_7
	END_MINUTE_8
	END_MINUTE_9
	END_MINUTE_10
	ACK_REVEAL_ENTRY
	ACK_COMMIT_CHAIN
	ACK_REVEAL_CHAIN
	ACK_COMMIT_ENTRY

	FORCE_FACTOID_GENESIS_REBUILD
	INFO_CURRENT_HEIGHT // info message to the wire-side to indicate the current known block height;
)

type MsgAck struct {
	Height      uint32
	ChainID     interfaces.IHash
	Index       uint32
	Typ         byte
	Affirmation interfaces.IHash // affirmation value -- hash of the message/object in question
	SerialHash  [32]byte
	Signature   [64]byte
}

// Write out the MsgAck (excluding Signature) to binary.
func (msg *MsgAck) GetBinaryForSignature() (data []byte, err error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, msg.Height)
	if msg.ChainID != nil {
		data, err = msg.ChainID.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(data)
	}

	binary.Write(&buf, binary.BigEndian, msg.Index)

	buf.Write([]byte{msg.Typ})

	buf.Write(msg.Affirmation.Bytes())

	buf.Write(msg.SerialHash[:])

	return buf.Bytes(), err
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the MsgAck interface implementation.
func (msg *MsgAck) BtcDecode(r io.Reader, pver uint32) error {
	newData, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("MsgAck.BtcDecode reader is invalid")
	}

	if len(newData) != 169 {
		return fmt.Errorf("MsgAck.BtcDecode reader does not have right length: ", len(newData))
	}

	msg.Height, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	msg.ChainID = new(primitives.Hash)
	newData, _ = msg.ChainID.UnmarshalBinaryData(newData)

	msg.Index, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	msg.Typ, newData = newData[0], newData[1:]

	msg.Affirmation = primitives.NewHash(newData[0:32])
	newData = newData[32:]

	copy(msg.SerialHash[:], newData[0:32])
	newData = newData[32:]

	copy(msg.Signature[:], newData[0:63])

	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the MsgAck interface implementation.
func (msg *MsgAck) BtcEncode(w io.Writer, pver uint32) error {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, msg.Height)
	buf.Write(msg.ChainID.Bytes())

	binary.Write(&buf, binary.BigEndian, msg.Index)
	buf.Write([]byte{msg.Typ})
	buf.Write(msg.Affirmation.Bytes())
	buf.Write(msg.SerialHash[:])
	buf.Write(msg.Signature[:])

	w.Write(buf.Bytes())

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the MsgAck interface implementation.
func (msg *MsgAck) Command() string {
	return CmdAck
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the MsgAck interface implementation.
func (msg *MsgAck) MaxPayloadLength(pver uint32) uint32 {

	// 10K is too big of course, TODO: adjust
	return MaxAppMsgPayload
}

// NewMsgAcknowledgement returns a new bitcoin ping message that conforms to the MsgAck
// interface.  See MsgAck for details.
func NewMsgAcknowledgement(height uint32, index uint32, affirm interfaces.IHash, ackType byte) *MsgAck {

	if affirm == nil {
		affirm = new(primitives.Hash)
	}
	return &MsgAck{
		Height:      height,
		ChainID:     new(primitives.Hash), //TODO: get the correct chain id from processor
		Index:       index,
		Affirmation: affirm,
		Typ:         ackType,
	}
}

// Create a sha hash from the message binary (output of BtcEncode)
func (msg *MsgAck) Sha() (interfaces.IHash, error) {

	buf := bytes.NewBuffer(nil)
	msg.BtcEncode(buf, ProtocolVersion)
	var sha interfaces.IHash
	_ = sha.SetBytes(Sha256(buf.Bytes()))

	return sha, nil
}

var _ interfaces.IMsg = (*MsgAck)(nil)

func (m *MsgAck) Process(uint32, interfaces.IState) {}

func (m *MsgAck) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgAck) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgAck) Type() int {
	return -1
}

func (m *MsgAck) Int() int {
	return -1
}

func (m *MsgAck) Bytes() []byte {
	return nil
}

func (m *MsgAck) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgAck) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgAck) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgAck) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgAck) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgAck is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgAck is valid
func (m *MsgAck) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgAck) Leader(state interfaces.IState) bool {
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
func (m *MsgAck) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgAck) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgAck) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgAck) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgAck) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgAck) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
