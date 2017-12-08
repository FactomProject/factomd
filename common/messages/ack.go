// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"encoding/binary"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	log "github.com/sirupsen/logrus"
)

//General acknowledge message
type Ack struct {
	MessageBase
	Timestamp   interfaces.Timestamp // Timestamp of Ack by Leader
	Salt        [8]byte              // Eight bytes of the salt
	SaltNumber  uint32               // Secret Number used to detect multiple servers with the same ID
	MessageHash interfaces.IHash     // Hash of message acknowledged
	DBHeight    uint32               // Directory Block Height that owns this ack
	Height      uint32               // Height of this ack in this process list
	SerialHash  interfaces.IHash     // Serial hash including previous ack

	DataAreaSize uint64 // Size of the Data Area
	DataArea     []byte // Data Area

	Signature interfaces.IFullSignature
	//Not marshalled
	hash        interfaces.IHash
	authvalid   bool
	Response    bool // A response to a missing data request
	BalanceHash interfaces.IHash
}

var _ interfaces.IMsg = (*Ack)(nil)
var _ Signable = (*Ack)(nil)
var AckBalanceHash = true

func (m *Ack) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

// We have to return the haswh of the underlying message.
func (m *Ack) GetHash() interfaces.IHash {
	return m.MessageHash
}

func (m *Ack) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *Ack) Type() byte {
	return constants.ACK_MSG
}

func (m *Ack) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *Ack) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *Ack) Validate(state interfaces.IState) int {
	// If too old, it isn't valid.
	if m.DBHeight <= state.GetHighestSavedBlk() {
		return -1
	}

	// Only new acks are valid. Of course, the VMIndex has to be valid too.
	msg, _ := state.GetMsg(m.VMIndex, int(m.DBHeight), int(m.Height))
	if msg != nil {
		return -1
	}

	if !m.authvalid {
		// Check signature
		bytes, err := m.MarshalForSignature()
		if err != nil {
			//fmt.Println("Err is not nil on Ack sig check: ", err)
			return -1
		}
		sig := m.Signature.GetSignature()
		ackSigned, err := state.VerifyAuthoritySignature(bytes, sig, m.DBHeight)

		//ackSigned, err := m.VerifySignature()
		if err != nil {
			//fmt.Println("Err is not nil on Ack sig check: ", err)
			return -1
		}
		if ackSigned <= 0 {
			return -1
		}
	}

	m.authvalid = true
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *Ack) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *Ack) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *Ack) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteAck(m)
}

// Acknowledgements do not go into the process list.
func (e *Ack) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

func (e *Ack) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *Ack) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *Ack) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *Ack) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *Ack) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data
	if newData[0] != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}
	newData = newData[1:]

	m.VMIndex, newData = int(newData[0]), newData[1:]

	m.Timestamp = new(primitives.Timestamp)
	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	copy(m.Salt[:], newData[:8])
	newData = newData[8:]

	m.SaltNumber, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	m.MessageHash = new(primitives.Hash)
	newData, err = m.MessageHash.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	newData, err = m.GetFullMsgHash().UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.LeaderChainID = new(primitives.Hash)
	newData, err = m.LeaderChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	m.Height, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	m.Minute, newData = newData[0], newData[1:]

	if m.SerialHash == nil {
		m.SerialHash = primitives.NewHash(constants.ZERO_HASH)
	}
	newData, err = m.SerialHash.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	if AckBalanceHash {
		m.DataAreaSize, newData = primitives.DecodeVarInt(newData)
		if m.DataAreaSize > 0 {
			das := newData[:int(m.DataAreaSize)]

			lenb := uint64(0)
			for len(das) > 0 {
				typeb := das[0]
				lenb, das = primitives.DecodeVarInt(das[1:])
				switch typeb {
				case 1:
					m.BalanceHash = primitives.NewHash(das[:32])
				}
				das = das[lenb:]
			}
			m.DataArea = append(m.DataArea[:0], newData[:m.DataAreaSize]...)
			newData = newData[int(m.DataAreaSize):]
		}
	}

	if len(newData) > 0 {
		m.Signature = new(primitives.Signature)
		newData, err = m.Signature.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
	}
	return
}

func (m *Ack) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *Ack) MarshalForSignature() ([]byte, error) {
	var buf primitives.Buffer

	binary.Write(&buf, binary.BigEndian, m.Type())
	binary.Write(&buf, binary.BigEndian, byte(m.VMIndex))

	t := m.GetTimestamp()
	data, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	buf.Write(m.Salt[:8])
	binary.Write(&buf, binary.BigEndian, m.SaltNumber)

	data, err = m.MessageHash.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.GetFullMsgHash().MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.LeaderChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, m.DBHeight)
	binary.Write(&buf, binary.BigEndian, m.Height)
	binary.Write(&buf, binary.BigEndian, m.Minute)

	data, err = m.SerialHash.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	if AckBalanceHash {
		if m.BalanceHash == nil {
			primitives.EncodeVarInt(&buf, 0)
			m.DataArea = nil
		} else {

			// Figure out all the data we are going to write out.
			var area primitives.Buffer
			area.WriteByte(1)
			primitives.EncodeVarInt(&area, 32)
			area.Write(m.BalanceHash.Bytes())

			// Write out the size of said data, and then the data.
			m.DataAreaSize = uint64(len(area.Bytes()))
			primitives.EncodeVarInt(&buf, m.DataAreaSize)
			buf.Write(area.Bytes())
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (m *Ack) MarshalBinary() (data []byte, err error) {
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

func (m *Ack) String() string {
	return fmt.Sprintf("%6s-VM%3d: PL:%5d DBHt:%5d -- Leader[:3]=%x hash[:3]=%x",
		"ACK",
		m.VMIndex,
		m.Height,
		m.DBHeight,
		m.LeaderChainID.Bytes()[:3],
		m.GetHash().Bytes()[:3])

}

func (m *Ack) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "ack", "dbheight": m.DBHeight, "vm": m.VMIndex,
		"vmheight": m.Height, "server": m.LeaderChainID.String(),
		"hash": m.GetHash().String()}
}

func (a *Ack) IsSameAs(b *Ack) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil {
		return false
	}

	if a.VMIndex != b.VMIndex {
		return false
	}

	if a.Minute != b.Minute {
		return false
	}

	if a.DBHeight != b.DBHeight {
		return false
	}

	if a.Height != b.Height {
		return false
	}

	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}

	if a.Salt != b.Salt {
		return false
	}

	if a.SaltNumber != b.SaltNumber {
		return false
	}

	if a.GetFullMsgHash().IsSameAs(b.GetFullMsgHash()) == false {
		return false
	}

	if a.MessageHash.IsSameAs(b.MessageHash) == false {
		return false
	}

	if a.SerialHash.IsSameAs(b.SerialHash) == false {
		return false
	}

	if a.DataAreaSize != b.DataAreaSize {
		return false
	}

	if a.Signature != nil {
		if a.Signature.IsSameAs(b.Signature) == false {
			return false
		}
	}

	if a.LeaderChainID == nil && b.LeaderChainID != nil {
		return false
	}
	if a.LeaderChainID != nil {
		if a.LeaderChainID.IsSameAs(b.LeaderChainID) == false {
			return false
		}
	}

	if a.BalanceHash == nil && b.BalanceHash != nil {
		return false
	}

	if b.BalanceHash == nil && a.BalanceHash != nil {
		return false
	}

	if a.BalanceHash == nil && b.BalanceHash == nil {
		return true
	}

	if a.BalanceHash.Fixed() != b.BalanceHash.Fixed() {
		return false
	}

	return true
}
