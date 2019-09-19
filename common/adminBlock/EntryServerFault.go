package adminBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// ServerFault is an admin block entry which contains all the information and messages from an election, which includes the removal and
// promotion of a server from and to the federated set. Note: this is not currently used, but may be revisited again in the future
type ServerFault struct {
	AdminIDType uint32               `json:"adminidtype"` //  the type of action in this admin block entry: uint32(TYPE_SERVER_FAULT)
	Timestamp   interfaces.Timestamp `json:"timestamp"`
	// The following 5 fields represent the "Core" of the message
	// This should match the Core of ServerFault messages
	ServerID      interfaces.IHash `json:"serverid"`
	AuditServerID interfaces.IHash `json:"auditserverid"`
	VMIndex       byte             `json:"vmindex"`
	DBHeight      uint32           `json:"dbheight"` // Directory block height where the server fault occurred
	Height        uint32           `json:"height"`   // The entry in the process list where the server fault occurred

	SignatureList SigList `json:"signaturelist"`
}

// Init initializes all nil hashes to the zero hash, sets the type, and sets the timestamp to zero
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

// SigList contains an array of signatures
type SigList struct {
	Length uint32
	List   []interfaces.IFullSignature
}

var _ interfaces.BinaryMarshallable = (*SigList)(nil)

// MarshalBinary marshals the object
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

// UnmarshalBinaryData unmarshals the input data into this object
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

// UnmarshalBinary unmarshals the input data into this object
func (sl *SigList) UnmarshalBinary(data []byte) error {
	_, err := sl.UnmarshalBinaryData(data)
	return err
}

// UpdateState updates the factomd state to include information related to this action
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

// MarshalCore marshals the core of the object, excluding the type, timestamp, and signature list
func (e *ServerFault) MarshalCore() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ServerFault.MarshalCore err:%v", *pe)
		}
	}(&err)
	e.Init()
	var buf primitives.Buffer

	err = buf.PushBinaryMarshallable(e.ServerID)
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(e.AuditServerID)
	if err != nil {
		return nil, err
	}

	err = buf.PushByte(e.VMIndex)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(e.DBHeight)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(e.Height)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// MarshalBinary marshals the object
func (e *ServerFault) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ServerFault.MarshalBinary err:%v", *pe)
		}
	}(&err)
	e.Init()
	var buf primitives.Buffer

	err = buf.PushByte(e.Type())
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(e.Timestamp)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.ServerID)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.AuditServerID)
	if err != nil {
		return nil, err
	}

	err = buf.PushByte(e.VMIndex)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(e.DBHeight)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(e.Height)
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(&e.SignatureList)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input data into this object
func (e *ServerFault) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	b, err := buf.PopByte()
	if err != nil {
		return nil, err
	}
	if b != e.Type() {
		return nil, fmt.Errorf("Invalid Entry type")
	}

	e.Timestamp = new(primitives.Timestamp)
	err = buf.PopBinaryMarshallable(e.Timestamp)
	if err != nil {
		return nil, err
	}

	if e.ServerID == nil {
		e.ServerID = primitives.NewZeroHash()
	}
	err = buf.PopBinaryMarshallable(e.ServerID)
	if err != nil {
		return nil, err
	}

	if e.AuditServerID == nil {
		e.AuditServerID = primitives.NewZeroHash()
	}
	err = buf.PopBinaryMarshallable(e.AuditServerID)
	if err != nil {
		return nil, err
	}

	e.VMIndex, err = buf.PopByte()
	e.DBHeight, err = buf.PopUInt32()
	e.Height, err = buf.PopUInt32()

	err = buf.PopBinaryMarshallable(&e.SignatureList)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinary unmarshals the input data into this object
func (e *ServerFault) UnmarshalBinary(data []byte) error {
	_, err := e.UnmarshalBinaryData(data)
	return err
}

// JSONByte returns the json encoded byte array
func (e *ServerFault) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *ServerFault) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

// IsInterpretable always returns false
func (e *ServerFault) IsInterpretable() bool {
	return false
}

// Interpret always returns the empty string ""
func (e *ServerFault) Interpret() string {
	return ""
}

// Hash marshals the object and computes its hash
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

// String returns the object as a string
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

// Type returns the hardcoded TYPE_SERVER_FAULT
func (e *ServerFault) Type() byte {
	return constants.TYPE_SERVER_FAULT
}

// Compare returns -1 if this object has an earlier timestamp than the input, +1 if a later timestamp. If both
// timestamps are equal, returns -1 if this object has a smaller VMIndex, +1 if a larger VMIndex, and finally 0 if equal
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
