// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

//A placeholder structure for messages
type FullServerFault struct {
	MessageBase
	Timestamp interfaces.Timestamp

	// The following 4 fields represent the "Core" of the message
	// This should match the Core of ServerFault messages
	ServerID      interfaces.IHash
	AuditServerID interfaces.IHash
	VMIndex       byte
	DBHeight      uint32
	Height        uint32

	SignatureList SigList

	Signature interfaces.IFullSignature

	//Not marshalled
	hash interfaces.IHash
}

type SigList struct {
	Length uint32
	List   []interfaces.IFullSignature
}

var _ interfaces.IMsg = (*FullServerFault)(nil)
var _ Signable = (*FullServerFault)(nil)

func (m *FullServerFault) Process(uint32, interfaces.IState) bool { return true }

func (m *FullServerFault) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *FullServerFault) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *FullServerFault) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *FullServerFault) GetCoreHash() interfaces.IHash {
	data, err := m.MarshalCore()
	if err != nil {
		return nil
	}
	return primitives.Sha(data)
}

func (m *FullServerFault) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *FullServerFault) Type() byte {
	return constants.FULL_SERVER_FAULT_MSG
}

func (m *FullServerFault) MarshalCore() (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error marshalling Server Fault Core: %v", r)
		}
	}()

	var buf primitives.Buffer

	if d, err := m.ServerID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}
	if d, err := m.AuditServerID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	buf.WriteByte(m.VMIndex)
	binary.Write(&buf, binary.BigEndian, uint32(m.DBHeight))
	binary.Write(&buf, binary.BigEndian, uint32(m.Height))

	return buf.DeepCopyBytes(), nil
}

func (m *FullServerFault) MarshalForSignature() (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error marshalling Invalid Server Fault: %v", r)
		}
	}()

	var buf primitives.Buffer

	buf.Write([]byte{m.Type()})
	if d, err := m.Timestamp.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}
	if d, err := m.ServerID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}
	if d, err := m.AuditServerID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	buf.WriteByte(m.VMIndex)
	binary.Write(&buf, binary.BigEndian, uint32(m.DBHeight))
	binary.Write(&buf, binary.BigEndian, uint32(m.Height))

	if d, err := m.SignatureList.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	return buf.DeepCopyBytes(), nil
}

func (sl *SigList) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer

	binary.Write(&buf, binary.BigEndian, uint32(sl.Length))

	for _, individualSig := range sl.List {
		if d, err := individualSig.MarshalBinary(); err != nil {
			return nil, err
		} else {
			buf.Write(d)
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (sl *SigList) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling SigList in Full Server Fault: %v", r)
		}
	}()
	newData = data
	sl.Length, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	for i := sl.Length; i > 0; i-- {
		tempSig := new(primitives.Signature)
		newData, err = tempSig.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
		sl.List = append(sl.List, tempSig)
	}
	return newData, nil
}

func (m *FullServerFault) MarshalBinary() (data []byte, err error) {
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

func (m *FullServerFault) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling With Signatures Invalid Server Fault: %v", r)
		}
	}()
	newData = data
	if newData[0] != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}
	newData = newData[1:]

	m.Timestamp = new(primitives.Timestamp)
	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	if m.ServerID == nil {
		m.ServerID = primitives.NewZeroHash()
	}
	newData, err = m.ServerID.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	if m.AuditServerID == nil {
		m.AuditServerID = primitives.NewZeroHash()
	}
	newData, err = m.AuditServerID.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.VMIndex, newData = newData[0], newData[1:]
	m.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	m.Height, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	newData, err = m.SignatureList.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	if len(newData) > 0 {
		m.Signature = new(primitives.Signature)
		newData, err = m.Signature.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
	}

	return newData, nil
}

func (m *FullServerFault) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *FullServerFault) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *FullServerFault) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

func (m *FullServerFault) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *FullServerFault) String() string {
	return fmt.Sprintf("%6s-VM%3d (%v) AuditID: %v PL:%5d DBHt:%5d -- hash[:3]=%x\n SigList: %+v",
		"FullSFault",
		m.VMIndex,
		m.ServerID.String()[:10],
		m.AuditServerID.String()[:10],
		m.Height,
		m.DBHeight,
		m.GetHash().Bytes()[:3],
		m.SignatureList)
}

func (m *FullServerFault) GetDBHeight() uint32 {
	return m.DBHeight
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *FullServerFault) Validate(state interfaces.IState) int {
	// Check main signature
	bytes, err := m.MarshalForSignature()
	if err != nil {
		return -1
	}
	sig := m.Signature.GetSignature()
	sfSigned, err := state.VerifyAuthoritySignature(bytes, sig, m.DBHeight)
	if err != nil {
		return -1
	}
	if sfSigned < 1 {
		return -1
	}
	cb, err := m.MarshalCore()
	if err != nil {
		return -1
	}
	validSigCount := 0
	for _, fedSig := range m.SignatureList.List {
		check, err := state.VerifyAuthoritySignature(cb, fedSig.GetSignature(), m.DBHeight)
		if err == nil && check == 1 {
			validSigCount++
		}
		if validSigCount > len(state.GetFedServers(m.DBHeight))/2 {
			//fmt.Println("JUSTIN", state.GetFactomNodeName(), "GOT VALIDD FF:", m.ServerID.String()[:10], "AUD:", m.AuditServerID.String()[:10])
			return 1
		}
	}
	//fmt.Println("JUSTIN NEVS", state.GetFactomNodeName(), validSigCount, m.AuditServerID.String()[:10], m.ServerID.String()[:10])
	return -1 // didn't see enough valid sigs
}

func (m *FullServerFault) ComputeVMIndex(state interfaces.IState) {

}

// Execute the leader functions of the given message
func (m *FullServerFault) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *FullServerFault) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteFullFault(m)
}

func (e *FullServerFault) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *FullServerFault) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *FullServerFault) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (a *FullServerFault) IsSameAs(b *FullServerFault) bool {
	if b == nil {
		return false
	}
	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}

	if a.Signature == nil && b.Signature != nil {
		return false
	}
	if a.Signature != nil {
		if a.Signature.IsSameAs(b.Signature) == false {
			return false
		}
	}
	if !a.ServerID.IsSameAs(b.ServerID) {
		return false
	}
	if !a.AuditServerID.IsSameAs(b.AuditServerID) {
		return false
	}
	//TODO: expand

	return true
}

func (a *FullServerFault) ToAdminBlockEntry() *adminBlock.ServerFault {
	sf := new(adminBlock.ServerFault)

	sf.Timestamp = a.Timestamp
	sf.ServerID = a.ServerID
	sf.AuditServerID = a.AuditServerID
	sf.VMIndex = a.VMIndex
	sf.DBHeight = a.DBHeight
	sf.Height = a.Height

	sf.SignatureList.Length = a.SignatureList.Length
	sf.SignatureList.List = a.SignatureList.List

	return sf
}

//*******************************************************************************
// Build Function
//*******************************************************************************

func NewFullServerFault(faultMessage *ServerFault, sigList []interfaces.IFullSignature) *FullServerFault {
	sf := new(FullServerFault)
	sf.Timestamp = faultMessage.Timestamp
	sf.VMIndex = faultMessage.VMIndex
	sf.DBHeight = faultMessage.DBHeight
	sf.Height = faultMessage.Height
	sf.ServerID = faultMessage.ServerID
	sf.AuditServerID = faultMessage.AuditServerID

	numSigs := len(sigList)
	var allSigs []interfaces.IFullSignature
	for _, sig := range sigList {
		allSigs = append(allSigs, sig)
	}

	sl := new(SigList)
	sl.Length = uint32(numSigs)
	sl.List = allSigs

	sf.SignatureList = *sl
	return sf
}
