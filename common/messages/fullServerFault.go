// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

//A placeholder structure for messages
type FullServerFault struct {
	MessageBase
	Timestamp interfaces.Timestamp

	ClearFault bool

	// The following 5 fields represent the "Core" of the message
	// This should match the Core of ServerFault messages
	ServerID      interfaces.IHash
	AuditServerID interfaces.IHash
	DBHeight      uint32 // The DBHeight of the Fault
	VMIndex       byte   // The VM that has faulted
	Height        uint32 // The Height of the VM at the time of the fault
	SystemHeight  uint32 // The order of this Full Fault relative to all Full Faults. (Height in System List)

	SignatureList SigList
	SSerialHash   interfaces.IHash // Serial hash of the previous Full Fault Messages

	Signature interfaces.IFullSignature

	//Not marshalled
	alreadyValidated bool
	alreadyProcessed bool
	hash             interfaces.IHash
	//Local FaultState information (not marshalled)
	AmINegotiator bool
	MyVoteTallied bool
	LocalVoteMap  map[[32]byte]interfaces.IFullSignature
	PledgeDone    bool
	LastMatch     int64
}

var _ interfaces.IMsg = (*FullServerFault)(nil)
var _ Signable = (*FullServerFault)(nil)

func (m *FullServerFault) GetAmINegotiator() bool {
	return m.AmINegotiator
}

func (m *FullServerFault) SetAmINegotiator(b bool) {
	m.AmINegotiator = b
}

func (m *FullServerFault) GetMyVoteTallied() bool {
	return m.MyVoteTallied
}

func (m *FullServerFault) SetMyVoteTallied(b bool) {
	m.MyVoteTallied = b
}

func (m *FullServerFault) GetPledgeDone() bool {
	return m.PledgeDone
}

func (m *FullServerFault) SetPledgeDone(b bool) {
	m.PledgeDone = b
}

func (m *FullServerFault) GetLastMatch() int64 {
	return m.LastMatch
}

func (m *FullServerFault) SetLastMatch(b int64) {
	m.LastMatch = b
}

func (m *FullServerFault) IsNil() bool {
	if m == nil {
		return true
	}
	if m.ServerID.IsZero() {
		return true
	}
	if m.AuditServerID == nil || m.AuditServerID.IsZero() {
		return true
	}
	return false
}

func (m *FullServerFault) AddFaultVote(issuerID [32]byte, sig interfaces.IFullSignature) {
	if m.IsNil() {
		return
	}
	if m.LocalVoteMap == nil {
		m.LocalVoteMap = make(map[[32]byte]interfaces.IFullSignature)
	}

	m.LocalVoteMap[issuerID] = sig
}

func (m *FullServerFault) Priority(state interfaces.IState) (priority int64) {
	now := state.GetTimestamp()

	// After 20 seconds, a negotiation's priority is now zero.
	if now.GetTimeSeconds()-m.Timestamp.GetTimeSeconds() > 20 {
		return 0
	}

	// oldest timestamp is highest priority
	priority = math.MaxInt64 - m.Timestamp.GetTime().UnixNano()
	// Mask off lowest byte
	priority = (priority | 0xFF) ^ 0xFF
	// Add VMIndex
	priority = priority + int64(m.VMIndex)
	return
}

// Return the serial height for this Full Fault message.  Can return nil if there is
// no process list at this dbheight, or if we are missing a preceeding Full Fault message.
func (m *FullServerFault) GetSerialHash() interfaces.IHash {
	if m.SSerialHash == nil {
		sh, err := primitives.CreateHash(m.SSerialHash, m.GetCoreHash())
		if err != nil {
			panic(err.Error())
		}
		m.SSerialHash = sh
	}
	return m.SSerialHash
}

func (m *FullServerFault) Process(dbheight uint32, state interfaces.IState) bool {
	return state.ProcessFullServerFault(dbheight, m)
}

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

func (m *FullServerFault) MarshalCore() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	err := buf.PushBinaryMarshallable(m.ServerID)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.AuditServerID)
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(m.VMIndex)
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
	err = buf.PushUInt32(m.SystemHeight)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.Timestamp)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (m *FullServerFault) MarshalForSF() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	err := buf.PushBinaryMarshallable(m.ServerID)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.AuditServerID)
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(m.VMIndex)
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
	err = buf.PushUInt32(m.SystemHeight)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.Timestamp)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (m *FullServerFault) MarshalForSignature() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	err := buf.PushByte(m.Type())
	if err != nil {
		return nil, err
	}
	err = buf.PushBool(m.ClearFault)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.ServerID)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.AuditServerID)
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(m.VMIndex)
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
	err = buf.PushUInt32(m.SystemHeight)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.Timestamp)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.SSerialHash)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(&m.SignatureList)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (m *FullServerFault) MarshalBinary() ([]byte, error) {
	h, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	buf := primitives.NewBuffer(h)

	sig := m.GetSignature()
	if sig != nil {
		buf.PushBinaryMarshallable(sig)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

// Make this stuff easier to read.
func Unmarshall(thing interfaces.BinaryMarshallable, err error, data []byte) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	newdata, err := thing.UnmarshalBinaryData(data)
	return newdata, err
}

//
//                               UnmarshalBinaryData for FullServerFault
//
func (m *FullServerFault) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)

	t, err := buf.PopByte()
	if err != nil {
		return nil, err
	}
	if t != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}

	m.ClearFault, err = buf.PopBool()
	if err != nil {
		return nil, err
	}

	if m.ServerID == nil {
		m.ServerID = primitives.NewZeroHash()
	}
	err = buf.PopBinaryMarshallable(m.ServerID)
	if err != nil {
		return nil, err
	}

	if m.AuditServerID == nil {
		m.AuditServerID = primitives.NewZeroHash()
	}
	err = buf.PopBinaryMarshallable(m.AuditServerID)
	if err != nil {
		return nil, err
	}

	m.VMIndex, err = buf.PopByte()
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
	m.SystemHeight, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	m.Timestamp = new(primitives.Timestamp)
	err = buf.PopBinaryMarshallable(m.Timestamp)
	if err != nil {
		return nil, err
	}

	m.SSerialHash = primitives.NewZeroHash()
	err = buf.PopBinaryMarshallable(m.SSerialHash)
	if err != nil {
		return nil, err
	}

	err = buf.PopBinaryMarshallable(&m.SignatureList)
	if err != nil {
		return nil, err
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
	if m == nil {
		return "-nil-"
	}
	return fmt.Sprintf("%6s-vm%02d[%d] (%v) AuditID: %v DBHt:%5d SysHt:%3d Clr:%t -- hash[:3]=%x Sig Cnt: %d TS:%d",
		"FullSFault",
		m.VMIndex,
		m.Height,
		m.ServerID.String()[4:10],
		m.AuditServerID.String()[4:10],
		m.DBHeight,
		m.SystemHeight,
		m.ClearFault,
		m.GetHash().Bytes()[:3],
		len(m.SignatureList.List),
		m.Timestamp.GetTimeSeconds())
}

func (m *FullServerFault) StringWithSigCnt(s interfaces.IState) string {
	if m == nil {
		return "-nil-"
	}
	return fmt.Sprintf(
		"%6s-vm%02d[%d] (%v) AuditID: %v DBHt:%5d SysHt:%3d Clr:%t -- hash[:3]=%x Valid Sigs: %d TS:%d",
		"FullSFault",
		m.VMIndex,
		m.Height,
		m.ServerID.String()[4:10],
		m.AuditServerID.String()[4:10],
		m.DBHeight,
		m.SystemHeight,
		m.ClearFault,
		m.GetHash().Bytes()[:3],
		m.SigTally(s),
		m.Timestamp.GetTimeSeconds())
}

func (m *FullServerFault) GetDBHeight() uint32 {
	return m.DBHeight
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *FullServerFault) Validate(state interfaces.IState) int {
	// Ignore old faults
	if m.DBHeight <= state.GetHighestSavedBlk() {
		return -1
	}

	if m.alreadyValidated {
		return 1
	}

	if m.DBHeight < state.GetLLeaderHeight() {
		return -1
	}

	if m.ServerID.IsZero() || m.AuditServerID.IsZero() {
		state.AddStatus("FULL FAULT Validate Fake Fault.  Ignore")
		return -1
	}

	// Check main signature
	bytes, err := m.MarshalForSignature()
	if err != nil {
		return -1
	}
	//sig := m.Signature.GetSignature()

	//sfSigned, err := state.VerifyAuthoritySignature(bytes, sig, m.DBHeight)
	sfSigned, err := state.FastVerifyAuthoritySignature(bytes, m.Signature, m.DBHeight)
	if err != nil {
		return -1
	}
	if sfSigned < 1 {
		return -1
	}
	_, err = m.MarshalCore()
	if err != nil {
		return -1
	}

	m.alreadyValidated = true
	return 1
}

func (m *FullServerFault) SetAlreadyProcessed() {
	m.alreadyProcessed = true
}

func (m *FullServerFault) GetAlreadyProcessed() bool {
	return m.alreadyProcessed
}

func (m *FullServerFault) HasEnoughSigs(state interfaces.IState) bool {
	sigTally := m.SigTally(state)
	if sigTally > len(state.GetFedServers(m.DBHeight))/2 {
		return true
	}
	return false
}

func (m *FullServerFault) SigTally(state interfaces.IState) int {
	validSigCount := 0
	// Check main signature
	bytes, err := m.MarshalForSignature()
	if err != nil {
		return 0
	}
	sig := m.Signature.GetSignature()
	sfSigned, err := state.VerifyAuthoritySignature(bytes, sig, m.DBHeight)
	if err != nil {
		return 0
	}
	if sfSigned < 1 {
		return 0
	}
	cb, err := m.MarshalCore()
	if err != nil {
		return 0
	}
	for _, fedSig := range m.SignatureList.List {
		check, err := state.VerifyAuthoritySignature(cb, fedSig.GetSignature(), m.DBHeight)
		if err == nil && check == 1 {
			validSigCount++
		}
	}

	return validSigCount
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

type SigList struct {
	Length uint32
	List   []interfaces.IFullSignature
}

var _ interfaces.BinaryMarshallable = (*SigList)(nil)

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

func (sl *SigList) UnmarshalBinary(data []byte) error {
	_, err := sl.UnmarshalBinaryData(data)
	return err
}

//*******************************************************************************
// Build Function
//*******************************************************************************

func NewFullServerFault(Previous *FullServerFault, faultMessage *ServerFault, sigList []interfaces.IFullSignature, sysHeight int) *FullServerFault {
	sf := new(FullServerFault)
	sf.ClearFault = false
	sf.Timestamp = faultMessage.Timestamp
	sf.VMIndex = faultMessage.VMIndex
	sf.DBHeight = faultMessage.DBHeight
	sf.Height = faultMessage.Height
	sf.ServerID = faultMessage.ServerID
	sf.AuditServerID = faultMessage.AuditServerID
	sf.SystemHeight = uint32(sysHeight)

	if Previous != nil {
		sf.SSerialHash = Previous.GetSerialHash()
	} else {
		core, err := sf.MarshalCore()
		if err != nil {
			panic("Failed to Marshal Core of a Full Server Fault")
		}
		sf.SSerialHash = primitives.Sha(core)
	}

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
