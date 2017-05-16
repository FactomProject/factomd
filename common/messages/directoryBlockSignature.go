// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type DirectoryBlockSignature struct {
	MessageBase
	Timestamp interfaces.Timestamp
	DBHeight  uint32
	//DirectoryBlockKeyMR   interfaces.IHash
	DirectoryBlockHeader  interfaces.IDirectoryBlockHeader
	ServerIdentityChainID interfaces.IHash

	Signature interfaces.IFullSignature

	// Signature that goes into the admin block
	// Signature of directory block header
	DBSignature interfaces.IFullSignature
	SysHeight   uint32
	SysHash     interfaces.IHash

	//Not marshalled
	Matches   bool
	Processed bool
	hash      interfaces.IHash
}

var _ interfaces.IMsg = (*DirectoryBlockSignature)(nil)
var _ Signable = (*DirectoryBlockSignature)(nil)

func (a *DirectoryBlockSignature) IsSameAs(b *DirectoryBlockSignature) bool {
	if b == nil {
		return false
	}

	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}
	if a.DBHeight != b.DBHeight {
		return false
	}

	if a.DirectoryBlockHeader == nil && b.DirectoryBlockHeader != nil {
		return false
	}
	if a.DirectoryBlockHeader != nil {
		if a.DirectoryBlockHeader.GetPrevFullHash().IsSameAs(b.DirectoryBlockHeader.GetPrevFullHash()) == false {
			return false
		}
	}

	if a.ServerIdentityChainID == nil && b.ServerIdentityChainID != nil {
		return false
	}
	if a.ServerIdentityChainID != nil {
		if a.ServerIdentityChainID.IsSameAs(b.ServerIdentityChainID) == false {
			return false
		}
	}

	if a.Signature == nil && b.Signature != nil {
		return false
	}
	if a.Signature != nil {
		if a.Signature.IsSameAs(b.Signature) == false {
			return false
		}
	}

	return true
}

func (e *DirectoryBlockSignature) Process(dbheight uint32, state interfaces.IState) bool {
	return state.ProcessDBSig(dbheight, e)
}

func (m *DirectoryBlockSignature) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *DirectoryBlockSignature) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *DirectoryBlockSignature) GetMsgHash() interfaces.IHash {
	data, _ := m.MarshalForSignature()
	if data == nil {
		return nil
	}
	m.MsgHash = primitives.Sha(data)

	return m.MsgHash
}

func (m *DirectoryBlockSignature) GetTimestamp() interfaces.Timestamp {
	if m.Timestamp == nil {
		m.Timestamp = new(primitives.Timestamp)
	}
	return m.Timestamp
}

func (m *DirectoryBlockSignature) Type() byte {
	return constants.DIRECTORY_BLOCK_SIGNATURE_MSG
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *DirectoryBlockSignature) Validate(state interfaces.IState) int {

	if m.IsValid() {
		return 1
	}

	if m.DBHeight <= state.GetHighestSavedBlk() {
		state.AddStatus(fmt.Sprintf("DirectoryBlockSignature: Fail dbstate ht: %v < dbht: %v  %s", m.DBHeight, state.GetHighestSavedBlk(), m.String()))
		return -1
	}

	found, _ := state.GetVirtualServers(m.DBHeight, 9, m.ServerIdentityChainID)

	if found == false {
		state.AddStatus(fmt.Sprintf("DirectoryBlockSignature: Fail dbht: %v Server not found %x %s",
			state.GetLLeaderHeight(),
			m.ServerIdentityChainID.Bytes()[3:5],
			m.String()))
		return 0
	}

	if m.IsLocal() {
		m.SetValid()
		return 1
	}

	isVer, err := m.VerifySignature()
	if err != nil || !isVer {
		state.AddStatus(fmt.Sprintf("DirectoryBlockSignature: Fail to Verify Sig dbht: %v %s", state.GetLLeaderHeight(), m.String()))
		// if there is an error during signature verification
		// or if the signature is invalid
		// the message is considered invalid
		return 0
	}

	marshalledMsg, _ := m.MarshalForSignature()
	authorityLevel, err := state.VerifyAuthoritySignature(marshalledMsg, m.Signature.GetSignature(), m.DBHeight)
	if err != nil || authorityLevel < 1 {
		//This authority is not a Fed Server (it's either an Audit or not an Authority at all)
		state.AddStatus(fmt.Sprintf("DirectoryBlockSignature: Fail to Verify Sig (not from a Fed Server) dbht: %v %s", state.GetLLeaderHeight(), m.String()))
		return 0
	}

	m.SetValid()
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *DirectoryBlockSignature) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
func (m *DirectoryBlockSignature) LeaderExecute(state interfaces.IState) {
	state.LeaderExecuteDBSig(m)
}

func (m *DirectoryBlockSignature) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteMsg(m)
}

func (m *DirectoryBlockSignature) Sign(key interfaces.Signer) error {
	// Signature that goes into admin block
	err := m.SignHeader(key)
	if err != nil {
		return err
	}

	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *DirectoryBlockSignature) SignHeader(key interfaces.Signer) error {
	header, err := m.DirectoryBlockHeader.MarshalBinary()
	if err != nil {
		return err
	}
	m.DBSignature = key.Sign(header)
	return nil
}

func (m *DirectoryBlockSignature) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *DirectoryBlockSignature) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

func (m *DirectoryBlockSignature) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)

	t, err := buf.PopByte()
	if err != nil {
		return nil, err
	}
	if t != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}

	// TimeStamp
	m.Timestamp = new(primitives.Timestamp)
	err = buf.PopBinaryMarshallable(m.Timestamp)
	if err != nil {
		return nil, err
	}
	m.SysHeight, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}
	m.SysHash = primitives.NewZeroHash()
	err = buf.PopBinaryMarshallable(m.SysHash)
	if err != nil {
		return nil, err
	}

	m.DBHeight, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}
	t, err = buf.PopByte()
	if err != nil {
		return nil, err
	}
	m.VMIndex = int(t)

	m.DirectoryBlockHeader = directoryBlock.NewDBlockHeader()
	err = buf.PopBinaryMarshallable(m.DirectoryBlockHeader)
	if err != nil {
		return nil, err
	}

	m.ServerIdentityChainID = primitives.NewZeroHash()
	err = buf.PopBinaryMarshallable(m.ServerIdentityChainID)
	if err != nil {
		return nil, err
	}

	//if len(newData) > 0 {
	m.DBSignature = new(primitives.Signature)
	err = buf.PopBinaryMarshallable(m.DBSignature)
	if err != nil {
		return nil, err
	}
	//}

	if buf.Len() > 0 {
		m.Signature = new(primitives.Signature)
		err = buf.PopBinaryMarshallable(m.Signature)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (m *DirectoryBlockSignature) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *DirectoryBlockSignature) MarshalForSignature() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	err := buf.PushByte(m.Type())
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.GetTimestamp())
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(m.SysHeight)
	if err != nil {
		return nil, err
	}
	if m.SysHash == nil {
		m.SysHash = primitives.NewZeroHash()
	}
	err = buf.PushBinaryMarshallable(m.SysHash)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(m.DBHeight)
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(byte(m.VMIndex))
	if err != nil {
		return nil, err
	}
	if m.DirectoryBlockHeader == nil {
		m.DirectoryBlockHeader = directoryBlock.NewDBlockHeader()
	}
	err = buf.PushBinaryMarshallable(m.DirectoryBlockHeader)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.ServerIdentityChainID)
	if err != nil {
		return nil, err
	}

	if m.DBSignature != nil {
		err = buf.PushBinaryMarshallable(m.DBSignature)
		if err != nil {
			return nil, err
		}
	} else {
		blankSig := make([]byte, constants.SIGNATURE_LENGTH)
		err = buf.Push(blankSig)
		if err != nil {
			return nil, err
		}
		err = buf.Push(blankSig[:32])
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (m *DirectoryBlockSignature) MarshalBinary() ([]byte, error) {
	h, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	buf := primitives.NewBuffer(h)
	sig := m.GetSignature()

	if sig != nil {
		err = buf.PushBinaryMarshallable(sig)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (m *DirectoryBlockSignature) String() string {
	return fmt.Sprintf("%6s-VM%3d:          DBHt:%5d -- Signer[:3]=%x PrevDBKeyMR[:3]=%x hash[:3]=%x",
		"DBSig",
		m.VMIndex,
		m.DBHeight,
		m.ServerIdentityChainID.Bytes()[2:6],
		m.DirectoryBlockHeader.GetPrevKeyMR().Bytes()[:3],
		m.GetHash().Bytes()[:3])

}

func (e *DirectoryBlockSignature) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *DirectoryBlockSignature) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}
