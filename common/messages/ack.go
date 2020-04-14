// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages/msgbase"
	"github.com/FactomProject/factomd/common/primitives"

	llog "github.com/FactomProject/factomd/log"
	log "github.com/sirupsen/logrus"
)

// packageLogger is the general logger for all message related logs. You can add additional fields,
// or create more context loggers off of this
var packageLogger = log.WithFields(log.Fields{"package": "messages"})

// Ack is a general acknowledge message
type Ack struct {
	msgbase.MessageBase
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
	hash         interfaces.IHash
	authvalid    bool
	Response     bool // A response to a missing data request
	BalanceHash  interfaces.IHash
	marshalCache []byte
}

var _ interfaces.IMsg = (*Ack)(nil)
var _ interfaces.Signable = (*Ack)(nil)

// AckBalanceHash is a global variable which determines whether balance hashes are passed around. Default is
// true, but can be overwritten on network startup
var AckBalanceHash = true

// GetRepeatHash returns the hash of the message, which is the sha256 of the marshaled object excluding its signature
func (m *Ack) GetRepeatHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "Ack.GetRepeatHash") }()

	return m.GetMsgHash()
}

// GetHash returns the hash of the underlying message.
func (m *Ack) GetHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "Ack.GetHash") }()

	return m.MessageHash
}

// GetMsgHash returns the hash of the message, which is the sha256 of the marshaled object excluding its signature
func (m *Ack) GetMsgHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "Ack.GetMsgHash") }()

	if m.MsgHash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

// Type returns this message type ACK_MSG
func (m *Ack) Type() byte {
	return constants.ACK_MSG
}

// GetTimestamp returns this object's timestamp
func (m *Ack) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp.Clone()
}

// VerifySignature verifies the message is signed properly
func (m *Ack) VerifySignature() (bool, error) {
	return msgbase.VerifyMessage(m)
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *Ack) Validate(s interfaces.IState) int {
	// If too old, it isn't valid.
	if m.DBHeight < s.GetLLeaderHeight() {
		s.LogMessage("executeMsg", "drop, from past", m)
		return -1
	}

	// Update the highest known ack to start requesting
	// DBState blocks if necessary
	if s.GetHighestAck() < m.DBHeight {
		s.SetHighestAck(m.DBHeight)
	}

	// drop future acks that are far in the future and not near the block we expect to build as soon as the boot finishes.
	if m.DBHeight-s.GetLLeaderHeight() > 5 && m.DBHeight != s.GetHighestKnownBlock() && m.DBHeight != s.GetHighestKnownBlock()+1 {
		s.LogMessage("executeMsg", "drop, from far future", m)
		return -1
	}

	if m.DBHeight > s.GetLLeaderHeight() {
		return s.HoldForHeight(m.DBHeight, 0, m) // release the ACKs at the start of minute 0 of their block
	}

	// Only new acks are valid. Of course, the VMIndex has to be valid too.
	msg, _ := s.GetMsg(m.VMIndex, int(m.DBHeight), int(m.Height))
	if msg != nil {
		if !msg.GetMsgHash().IsSameAs(m.GetHash()) {
			s.LogMessage("executeMsg", "Ack slot taken", m)
			s.LogMessage("executeMsg", "found:", msg)
		} else {
			s.LogPrintf("executeMsg", "duplicate at %d/%d/%d", int(m.DBHeight), m.VMIndex, int(m.Height))
		}
		return -1
	}

	if !m.authvalid {
		// Check signature
		bytes, err := m.MarshalForSignature()
		if err != nil {
			s.LogPrintf("executeMsg", "Validate Marshal Failed %v", err)
			//fmt.Println("Err is not nil on Ack sig check: ", err)
			return -1
		}
		ackSigned, err := s.FastVerifyAuthoritySignature(bytes, m.Signature, m.DBHeight)

		//ackSigned, err := m.VerifySignature()
		if err != nil {
			s.LogPrintf("executeMsg", "VerifyAuthoritySignature Failed %v", err)
			// Don't return fail here because the message might be a future message and thus become valid in the future.
		}

		if ackSigned <= 0 {
			if m.DBHeight == s.GetLLeaderHeight() && m.Minute < byte(s.GetCurrentMinute()) {
				s.LogPrintf("executeMsg", "Hold, Not signed by a leader")
				return 0
			}
			s.LogPrintf("executeMsg", "Drop, Not signed by a leader")
			return -1
		}
	}

	m.authvalid = true
	s.LogMessage("executeMsg", "Valid", m)
	return 1
}

// ComputeVMIndex is a no op function
func (m *Ack) ComputeVMIndex(state interfaces.IState) {
}

// LeaderExecute executes the leader functions of the given message
// Leader, follower, do the same thing.
func (m *Ack) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

// FollowerExecute executes the follower functions of the given message
// Leader, follower, do the same thing.
func (m *Ack) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteAck(m)
}

// Process panics if called. Acknowledgements do not go into the process list.
func (m *Ack) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Ack object should never have its Process() method called")
}

// JSONByte returns the json encoded byte array
func (m *Ack) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(m)
}

// JSONString returns the json encoded string
func (m *Ack) JSONString() (string, error) {
	return primitives.EncodeJSONString(m)
}

// Sign signs this message with the input key
func (m *Ack) Sign(key interfaces.Signer) error {
	signature, err := msgbase.SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

// GetSignature returns the stored signature
func (m *Ack) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

// UnmarshalBinaryData unmarshals the input data into this object
func (m *Ack) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
			llog.LogPrintf("recovery", "Error unmarshalling: %v", r)
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

	b, newData := newData[0], newData[1:]

	if b > 0 {
		m.Signature = new(primitives.Signature)
		newData, err = m.Signature.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
	}

	m.marshalCache = append(m.marshalCache, data[:len(data)-len(newData)]...)

	return
}

// UnmarshalBinary unmarshals the input data into this object
func (m *Ack) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

// MarshalForSignature marshals this object excluding the signature
func (m *Ack) MarshalForSignature() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "Ack.MarshalForSignature err:%v", *pe)
		}
	}(&err)
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

// MarshalBinary marshals the object
func (m *Ack) MarshalBinary() (data []byte, err error) {

	if m.marshalCache != nil {
		return m.marshalCache, nil
	}

	resp, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	sig := m.GetSignature()

	if sig != nil {
		resp = append(resp, 1)
		sigBytes, err := sig.MarshalBinary()
		if err != nil {
			return nil, err
		}
		return append(resp, sigBytes...), nil
	}
	resp = append(resp, 0)

	return resp, nil
}

// String returns this message as a string
func (m *Ack) String() string {
	return fmt.Sprintf("%6s-%27s -- Leader[%x] hash[%x]",
		"ACK",
		fmt.Sprintf("DBh/VMh/h %d/%d/%d       ", m.DBHeight, m.VMIndex, m.Height),
		m.LeaderChainID.Bytes()[3:6],
		m.GetHash().Bytes()[:3])

}

// LogFields returns a map structure of important fields of the Ack
func (m *Ack) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "ack", "dbheight": m.DBHeight, "vm": m.VMIndex,
		"vmheight": m.Height, "server": m.LeaderChainID.String(),
		"hash": m.GetHash().String()}
}

// IsSameAs returns true iff the input object is identical to this object
func (m *Ack) IsSameAs(b *Ack) bool {
	if m == nil && b == nil {
		return true
	}

	if m == nil {
		return false
	}

	if m.VMIndex != b.VMIndex {
		return false
	}

	if m.Minute != b.Minute {
		return false
	}

	if m.DBHeight != b.DBHeight {
		return false
	}

	if m.Height != b.Height {
		return false
	}

	if m.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}

	if m.Salt != b.Salt {
		return false
	}

	if m.SaltNumber != b.SaltNumber {
		return false
	}

	if m.GetFullMsgHash().IsSameAs(b.GetFullMsgHash()) == false {
		return false
	}

	if m.MessageHash.IsSameAs(b.MessageHash) == false {
		return false
	}

	if m.SerialHash.IsSameAs(b.SerialHash) == false {
		return false
	}

	if m.DataAreaSize != b.DataAreaSize {
		return false
	}

	if m.Signature != nil {
		if m.Signature.IsSameAs(b.Signature) == false {
			return false
		}
	}

	if m.LeaderChainID == nil && b.LeaderChainID != nil {
		return false
	}
	if m.LeaderChainID != nil {
		if m.LeaderChainID.IsSameAs(b.LeaderChainID) == false {
			return false
		}
	}

	if m.BalanceHash == nil && b.BalanceHash != nil {
		return false
	}

	if m.BalanceHash == nil && b.BalanceHash != nil {
		return false
	}

	if m.BalanceHash == nil && b.BalanceHash == nil {
		return true
	}

	if m.BalanceHash.Fixed() != b.BalanceHash.Fixed() {
		return false
	}

	return true
}
