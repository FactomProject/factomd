// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
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
	_, err := state.GetMsg(m.VMIndex, int(m.DBHeight), int(m.Height))
	if err != nil {
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

func (m *Ack) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	t, err := buf.PopByte()
	if err != nil {
		return nil, err
	}
	if t != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}

	t, err = buf.PopByte()
	if err != nil {
		return nil, err
	}
	m.VMIndex = int(t)

	m.Timestamp = new(primitives.Timestamp)
	err = buf.PopBinaryMarshallable(m.Timestamp)
	if err != nil {
		return nil, err
	}

	bs, err := buf.PopLen(8)
	if err != nil {
		return nil, err
	}
	copy(m.Salt[:], bs)

	m.SaltNumber, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	m.MessageHash = new(primitives.Hash)
	err = buf.PopBinaryMarshallable(m.MessageHash)
	if err != nil {
		return nil, err
	}
	err = buf.PopBinaryMarshallable(m.GetFullMsgHash())
	if err != nil {
		return nil, err
	}

	m.LeaderChainID = new(primitives.Hash)
	err = buf.PopBinaryMarshallable(m.LeaderChainID)
	if err != nil {
		return nil, err
	}

	m.DBHeight, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}
	m.Height, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}
	m.Minute, err = buf.PopByte()
	if err != nil {
		return nil, err
	}

	if m.SerialHash == nil {
		m.SerialHash = primitives.NewZeroHash()
	}
	err = buf.PopBinaryMarshallable(m.SerialHash)
	if err != nil {
		return nil, err
	}

	if AckBalanceHash {
		m.DataAreaSize, err = buf.PopVarInt()
		if err != nil {
			return nil, err
		}
		if m.DataAreaSize > 0 {
			das, err := buf.PopLen(int(m.DataAreaSize))
			if err != nil {
				return nil, err
			}
			m.DataArea = make([]byte, len(das))
			copy(m.DataArea, das)

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
		}
	}

	if buf.Len() > 0 {
		m.Signature = new(primitives.Signature)
		err = buf.PopBinaryMarshallable(m.Signature)
		if err != nil {
			return nil, err
		}
	}
	return buf.DeepCopyBytes(), nil
}

func (m *Ack) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *Ack) MarshalForSignature() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	err := buf.PushByte(m.Type())
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(byte(m.VMIndex))
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(m.GetTimestamp())
	if err != nil {
		return nil, err
	}

	err = buf.Push(m.Salt[:8])
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(m.SaltNumber)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.MessageHash)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.GetFullMsgHash())
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.LeaderChainID)
	if err != nil {
		return nil, err
	}

	err = buf.PushUInt32(m.DBHeight)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(m.Height)
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(m.Minute)
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(m.SerialHash)
	if err != nil {
		return nil, err
	}

	if AckBalanceHash {
		if m.BalanceHash == nil {
			err = buf.PushVarInt(0)
			if err != nil {
				return nil, err
			}
			m.DataArea = nil
		} else {

			// Figure out all the data we are going to write out.
			var area primitives.Buffer
			area.WriteByte(1)
			primitives.EncodeVarInt(&area, 32)
			area.Write(m.BalanceHash.Bytes())

			// Write out the size of said data, and then the data.
			m.DataAreaSize = uint64(len(area.Bytes()))
			primitives.EncodeVarInt(buf, m.DataAreaSize)
			err = buf.Push(area.Bytes())
			if err != nil {
				return nil, err
			}
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (m *Ack) MarshalBinary() ([]byte, error) {
	resp, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	buf := primitives.NewBuffer(resp)

	sig := m.GetSignature()
	if sig != nil {
		err := buf.PushBinaryMarshallable(sig)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
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
