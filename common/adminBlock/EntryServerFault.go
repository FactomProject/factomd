package adminBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type ServerFault struct {
	AdminIDType uint32               `json:"adminidtype"`
	Timestamp   interfaces.Timestamp `json:"timestamp"`
	// The following 4 fields represent the "Core" of the message
	// This should match the Core of ServerFault messages
	ServerID      interfaces.IHash `json:"serverid"`
	AuditServerID interfaces.IHash `json:"auditserverid"`
	VMIndex       byte             `json:"vmindex"`
	DBHeight      uint32           `json:"dbheight"`
	Height        uint32           `json:"height"`

	SignatureList SigList `json:"signaturelist"`
}

func (e *ServerFault) Init() {
	if e.Timestamp == nil {
		e.Timestamp = primitives.NewTimestampFromMilliseconds(0)
	}
	if e.ServerID == nil {
		e.ServerID = primitives.NewZeroHash()
	}
	if e.AuditServerID == nil {
		e.AuditServerID = primitives.NewZeroHash()
	}
	e.AdminIDType = uint32(e.Type())
}

var _ interfaces.IABEntry = (*ServerFault)(nil)
var _ interfaces.BinaryMarshallable = (*ServerFault)(nil)

type SigList struct {
	Length uint32
	List   []interfaces.IFullSignature
}

var _ interfaces.BinaryMarshallable = (*SigList)(nil)

func (sl *SigList) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer

	buf.PushUInt32(sl.Length)

	for _, individualSig := range sl.List {
		if individualSig == nil {
			return nil, fmt.Errorf("Nil signature present")
		}
		err := buf.PushBinaryMarshallable(individualSig)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (sl *SigList) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)

	var err error
	sl.Length, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	for i := 0; i < int(sl.Length); i++ {
		tempSig := new(primitives.Signature)
		err = buf.PopBinaryMarshallable(tempSig)
		if err != nil {
			return nil, err
		}
		sl.List = append(sl.List, tempSig)
	}
	return buf.DeepCopyBytes(), nil
}

func (m *SigList) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (e *ServerFault) UpdateState(state interfaces.IState) error {
	e.Init()
	core, err := e.MarshalCore()
	if err != nil {
		return err
	}

	verifiedSignatures := 0
	for _, fullSig := range e.SignatureList.List {
		sig := fullSig.GetSignature()
		v, err := state.VerifyAuthoritySignature(core, sig, state.GetLeaderHeight())
		if err != nil {
			if err.Error() != "Signature Key Invalid or not Federated Server Key" {
				return err
			}
		}
		if v == 1 {
			verifiedSignatures++
		}
	}

	feds := state.GetFedServers(state.GetLeaderHeight())

	//50% threshold
	if verifiedSignatures <= len(feds)/2 {
		return fmt.Errorf(fmt.Sprintf("Quorum not reached for ServerFault.  Have %d sigs out of %d feds", verifiedSignatures, len(feds)))
	}

	//TODO: do
	/*
		state.AddFedServer(e.DBHeight, e.IdentityChainID)
		state.UpdateAuthorityFromABEntry(e)
	*/
	return nil
}

func (m *ServerFault) MarshalCore() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ServerFault.MarshalCore err:%v", *pe)
		}
	}(&err)
	m.Init()
	var buf primitives.Buffer

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

	return buf.DeepCopyBytes(), nil
}

func (m *ServerFault) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ServerFault.MarshalBinary err:%v", *pe)
		}
	}(&err)
	m.Init()
	var buf primitives.Buffer

	err = buf.PushByte(m.Type())
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(m.Timestamp)
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

	err = buf.PushBinaryMarshallable(&m.SignatureList)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (m *ServerFault) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	b, err := buf.PopByte()
	if err != nil {
		return nil, err
	}
	if b != m.Type() {
		return nil, fmt.Errorf("Invalid Entry type")
	}

	m.Timestamp = new(primitives.Timestamp)
	err = buf.PopBinaryMarshallable(m.Timestamp)
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
	m.DBHeight, err = buf.PopUInt32()
	m.Height, err = buf.PopUInt32()

	err = buf.PopBinaryMarshallable(&m.SignatureList)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (m *ServerFault) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (e *ServerFault) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

func (e *ServerFault) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

func (e *ServerFault) IsInterpretable() bool {
	return false
}

func (e *ServerFault) Interpret() string {
	return ""
}

func (e *ServerFault) Hash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("ServerFault.Hash() saw an interface that was nil")
		}
	}()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}

func (e *ServerFault) String() string {
	e.Init()
	str := fmt.Sprintf("    E: %35s -- DBheight %ds FedID %8x AuditServer %8x, #sigs %d, VMIndex %d",
		"EntryServerFault",
		e.DBHeight,
		e.ServerID.Bytes()[3:6],
		e.AuditServerID.Bytes()[3:6],
		len(e.SignatureList.List), e.VMIndex)
	return str
}

func (e *ServerFault) Type() byte {
	return constants.TYPE_SERVER_FAULT
}

func (e *ServerFault) Compare(b *ServerFault) int {
	if e.Timestamp.GetTimeMilliUInt64() < b.Timestamp.GetTimeMilliUInt64() {
		return -1
	}
	if e.Timestamp.GetTimeMilliUInt64() > b.Timestamp.GetTimeMilliUInt64() {
		return 1
	}
	if e.VMIndex < b.VMIndex {
		return -1
	}
	if e.VMIndex > b.VMIndex {
		return 1
	}
	return 0
}
