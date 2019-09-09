package adminBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// AddEfficiency is an admin block entry which contains a server identity and an efficiency for that server. The efficiency is specified in
// hundredths of a percent. So 0% would be specified as 0, 52.34% is specified as 5234, and 100% is specified as 10000.
type AddEfficiency struct {
	AdminIDType     uint32           `json:"adminidtype"` // the type of action in this admin block entry: uint32(TYPE_ADD_FACTOID_EFFICIENCY)
	IdentityChainID interfaces.IHash // the server identity whose efficiency will be updated
	Efficiency      uint16           // the efficiency this server will run at from 0 to 10000 (0% to 100.00%)
}

var _ interfaces.IABEntry = (*AddEfficiency)(nil)
var _ interfaces.BinaryMarshallable = (*AddEfficiency)(nil)

// Init initializes any nil hashes to the zero hash and sets the object type
func (e *AddEfficiency) Init() {
	e.AdminIDType = uint32(e.Type())
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}
}

// IsSameAs returns true iff the input object is identical to this object
func (a *AddEfficiency) IsSameAs(b *AddEfficiency) bool {
	if a.Type() != b.Type() {
		return false
	}

	if !a.IdentityChainID.IsSameAs(b.IdentityChainID) {
		return false
	}

	if a.Efficiency != b.Efficiency {
		return false
	}

	return true
}

// SortedIdentity returns the server identity associated with the efficiency change
func (e *AddEfficiency) SortedIdentity() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AddEfficiency.SortedIdentity() saw an interface that was nil")
		}
	}()

	return e.IdentityChainID
}

// String returns the AddEfficiency string
func (e *AddEfficiency) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %20s -- %17s %8x %12s %d",
		"AddAuditServer",
		"IdentityChainID", e.IdentityChainID.Bytes()[3:6],
		"Efficiency", e.Efficiency))
	return (string)(out.DeepCopyBytes())
}

// UpdateState updates factomd's state with information on the new efficiency
func (e *AddEfficiency) UpdateState(state interfaces.IState) error {
	e.Init()
	//state.AddAuditServer(c.DBHeight, c.IdentityChainID)
	state.UpdateAuthorityFromABEntry(e)

	return nil
}

// NewAddEfficiency creates a new AddEfficiency from the inputs. Efficiencies above 10000 (100%)
// are truncated to 10000
func NewAddEfficiency(chainID interfaces.IHash, efficiency uint16) (e *AddEfficiency) {
	e = new(AddEfficiency)
	e.Init()
	e.IdentityChainID = chainID
	if efficiency > 10000 {
		efficiency = 10000
	}
	e.Efficiency = efficiency
	return
}

// Type returns the hardcoded TYPE_ADD_FACTOID_EFFICIENCY
func (e *AddEfficiency) Type() byte {
	return constants.TYPE_ADD_FACTOID_EFFICIENCY
}

// MarshalBinary marshals the object
func (e *AddEfficiency) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "AddEfficiency.MarshalBinary err:%v", *pe)
		}
	}(&err)
	e.Init()
	var buf primitives.Buffer

	err = buf.PushByte(e.Type())
	if err != nil {
		return nil, err
	}

	// Need the size of the body
	var bodybuf primitives.Buffer
	err = bodybuf.PushIHash(e.IdentityChainID)
	if err != nil {
		return nil, err
	}

	err = bodybuf.PushUInt16(e.Efficiency)
	if err != nil {
		return nil, err
	}
	// end body

	err = buf.PushVarInt(uint64(bodybuf.Len()))
	if err != nil {
		return nil, err
	}

	err = buf.Push(bodybuf.Bytes())
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input data into this object
func (e *AddEfficiency) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	e.Init()

	t, err := buf.PopByte()
	if err != nil {
		return nil, err
	}

	if t != e.Type() {
		return nil, fmt.Errorf("Invalid Entry type")
	}

	bodyLimit := uint64(buf.Len())
	bodySize, err := buf.PopVarInt()
	if err != nil {
		return nil, err
	}
	if bodySize > bodyLimit {
		return nil, fmt.Errorf(
			"Error: AddEfficiency.UnmarshalBinary: body size %d is larger "+
				"than binary size %d. (uint underflow?)",
			bodySize, bodyLimit,
		)

	}

	body := make([]byte, bodySize)
	n, err := buf.Read(body)
	if err != nil {
		return nil, err
	}

	if uint64(n) != bodySize {
		return nil, fmt.Errorf("Expected to read %d bytes, but got %d", bodySize, n)
	}

	bodyBuf := primitives.NewBuffer(body)

	if uint64(n) != bodySize {
		return nil, fmt.Errorf("Unable to unmarshal body")
	}

	e.IdentityChainID, err = bodyBuf.PopIHash()
	if err != nil {
		return nil, err
	}

	e.Efficiency, err = bodyBuf.PopUInt16()
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinary unmarshals the input data into this object
func (e *AddEfficiency) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *AddEfficiency) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *AddEfficiency) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

// IsInterpretable always returns false
func (e *AddEfficiency) IsInterpretable() bool {
	return false
}

// Interpret always returns the empty string ""
func (e *AddEfficiency) Interpret() string {
	return ""
}

// Hash marshals the object and takes its hash
func (e *AddEfficiency) Hash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AddEfficiency.Hash() saw an interface that was nil")
		}
	}()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
