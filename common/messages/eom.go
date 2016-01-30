// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
)

var _ = log.Printf

type EOM struct {
	Timestamp interfaces.Timestamp
	Minute    byte

	DirectoryBlockHeight uint32
	ServerIndex          int
	Signature            interfaces.IFullSignature

	//Not marshalled
	hash interfaces.IHash
}

//var _ interfaces.IConfirmation = (*EOM)(nil)
var _ Signable = (*EOM)(nil)

func (e *EOM) Process(dbheight uint32, state interfaces.IState) {

	state.GetFactoidState(dbheight).EndOfPeriod(int(e.Minute))

	ecblk := state.GetEntryCreditBlock(dbheight)
	ecbody := ecblk.GetBody()
	mn := entryCreditBlock.NewMinuteNumber2(e.Minute)

	ecbody.AddEntry(mn)

	if e.Minute == 9 {

		// TODO: This code needs to be reviewed... It works here, but we are effectively
		// executing "leader" code in the compainion "follower" goroutine...
		// Maybe that's okay?
		if state.LeaderFor(e.Bytes()) {
			// What really needs to happen is that we look to make sure all
			// EOM messages have been recieved.  If this is the LAST message,
			// and we have ALL EOM messages from all servers, then we
			// create a DirectoryBlockSignature (if we are the leader) and
			// send it out to the network.
			DBM := NewDirectoryBlockSignature()
			if state.GetDirectoryBlock(dbheight-1) == nil {
				DBM.DirectoryBlockKeyMR = primitives.NewHash(constants.ZERO_HASH)
			} else {
				DBM.DirectoryBlockKeyMR = state.GetDirectoryBlock(dbheight - 1).GetKeyMR()
			}
			DBM.Sign(state)

			ack, err := NewAck(state, DBM.GetHash())
			if err != nil {
				fmt.Println("Ack Error")
				return
			}

			state.NetworkOutMsgQueue() <- ack
			state.NetworkOutMsgQueue() <- DBM
			state.InMsgQueue() <- ack
			state.InMsgQueue() <- DBM
		}
	}
}

func (m *EOM) GetHash() interfaces.IHash {
	if m.hash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			panic(fmt.Sprintf("Error in EOM.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *EOM) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *EOM) Int() int {
	return int(m.Minute)
}

func (m *EOM) Bytes() []byte {
	var ret []byte
	return append(ret, m.Minute)
}

func (m *EOM) Type() int {
	return constants.EOM_MSG
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *EOM) Validate(dbheight uint32, state interfaces.IState) int {
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *EOM) Leader(state interfaces.IState) bool {
	return state.LeaderFor(m.Bytes()) // TODO: This has to be fixed!
}

// Execute the leader functions of the given message
func (m *EOM) LeaderExecute(state interfaces.IState) error {
	b := m.GetHash()

	ack, err := NewAck(state, b)
	if err != nil {
		fmt.Println("Ack Error")
		return err
	}

	// Leader Execute creates an acknowledgement and the EOM
	state.NetworkOutMsgQueue() <- ack
	state.FollowerInMsgQueue() <- ack // Send the Ack to follower

	state.NetworkOutMsgQueue() <- m // Send the Message;  It works better if
	state.FollowerInMsgQueue() <- m // the msg follows the Ack
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *EOM) Follower(interfaces.IState) bool {
	return true
}

func (m *EOM) FollowerExecute(state interfaces.IState) error {
	_, err := state.MatchAckFollowerExecute(m)

	return err
}

func (e *EOM) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *EOM) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *EOM) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (m *EOM) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *EOM) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *EOM) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

func (m *EOM) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data[1:]

	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.Minute, newData = newData[0], newData[1:]

	if m.Minute < 0 || m.Minute >= 10 {
		return nil, fmt.Errorf("Minute number is out of range")
	}

	m.DirectoryBlockHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	m.ServerIndex = int(newData[0])
	newData = newData[1:]

	if len(newData) > 0 {
		sig := new(primitives.Signature)
		newData, err = sig.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
		m.Signature = sig
	}

	return data, nil
}

func (m *EOM) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *EOM) MarshalForSignature() (data []byte, err error) {
	var buf bytes.Buffer
	buf.Write([]byte{byte(m.Type())})
	if d, err := m.Timestamp.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}
	binary.Write(&buf, binary.BigEndian, m.Minute)
	binary.Write(&buf, binary.BigEndian, m.DirectoryBlockHeight)
	binary.Write(&buf, binary.BigEndian, uint8(m.ServerIndex))
	return buf.Bytes(), nil
}

func (m *EOM) MarshalBinary() (data []byte, err error) {
	resp, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	sig := m.GetSignature()

	if sig != nil {
		sigBytes, err := sig.MarshalBinary()
		if err != nil {
			return nil, err
		}
		return append(resp, sigBytes...), nil
	}
	return resp, nil
}

func (m *EOM) String() string {
	return fmt.Sprintf("%s: %d, DBHeight %d: %s", "EOM", m.Minute+1, m.DirectoryBlockHeight, m.GetHash().String())
}

// EOM methods that conform to the Message interface.

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *EOM) BtcEncode(w io.Writer, pver uint32) error {
	bytes, err := msg.MarshalBinary()
	if err != nil {
		return err
	}

	if err := writeVarBytes(w, pver, bytes); err != nil {
		return err
	}

	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *EOM) BtcDecode(r io.Reader, pver uint32) error {
	bytes, err := readVarBytes(r, pver, uint32(8+1+4+32+32+64), CmdEOM)
	if err != nil {
		return err
	}

	if err := msg.UnmarshalBinary(bytes); err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *EOM) Command() string {
	return CmdEOM
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *EOM) MaxPayloadLength(pver uint32) uint32 {
	return MaxAppMsgPayload
}

// Check whether the msg can pass the message level validations
// such as timestamp, signiture and etc
func (msg *EOM) IsValid() bool {
	//return msg.EOM.IsValid()
	return true
}

// Create a sha hash from the message binary (output of BtcEncode)
func (msg *EOM) Sha() (interfaces.IHash, error) {

	buf := bytes.NewBuffer(nil)
	msg.BtcEncode(buf, ProtocolVersion)
	var sha interfaces.IHash
	_ = sha.SetBytes(Sha256(buf.Bytes()))

	return sha, nil
}

/**********************************************************************
 * Support
 **********************************************************************/

func NewEOM(state interfaces.IState, minute int) interfaces.IMsg {
	// The construction of the EOM message needs information from the state of
	// the server to create the proper serial hashes and such.  Right now
	// I am ignoring all of that.
	eom := new(EOM)
	eom.Minute = byte(minute)
	eom.ServerIndex = state.GetServerIndex(state.GetDBHeight())

	return eom
}
