package adminBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// AddFactoidAddress is an admin block entry containing a server identity and a factoid address where coinbase payouts will be sent for that
// server
type AddFactoidAddress struct {
	AdminIDType     uint32              `json:"adminidtype"`     // the type of action in this admin block entry: uint32(TYPE_ADD_FACTOID_ADDRESS)
	IdentityChainID interfaces.IHash    `json:"identitychainid"` // the server identity
	FactoidAddress  interfaces.IAddress `json:"factoidaddress"`  // the factoid address for this server
}

var _ interfaces.IABEntry = (*AddFactoidAddress)(nil)
var _ interfaces.BinaryMarshallable = (*AddFactoidAddress)(nil)

// Init initializes any nil hashes to the zero hash and sets the objects type
func (e *AddFactoidAddress) Init() {
	e.AdminIDType = uint32(e.Type())
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}

	if e.FactoidAddress == nil {
		e.FactoidAddress = &factoid.Address{*(primitives.NewZeroHash().(*primitives.Hash))}
	}
}

// IsSameAs returns true iff the input is identital to this object
func (e *AddFactoidAddress) IsSameAs(b *AddFactoidAddress) bool {
	if e.Type() != b.Type() {
		return false
	}

	if !e.IdentityChainID.IsSameAs(b.IdentityChainID) {
		return false
	}

	if !e.FactoidAddress.IsSameAs(b.FactoidAddress) {
		return false
	}

	return true
}

// String returns the AddFactoidAddress string
func (e *AddFactoidAddress) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %20s -- %17s %8x %12s %s",
		"AddAuditServer",
		"IdentityChainID", e.IdentityChainID.Bytes()[3:6],
		"Address", e.FactoidAddress.String()))
	return (string)(out.DeepCopyBytes())
}

// UpdateState updates the factomd state with the new entry information
func (e *AddFactoidAddress) UpdateState(state interfaces.IState) error {
	e.Init()
	//state.AddAuditServer(c.DBHeight, c.IdentityChainID)
	state.UpdateAuthorityFromABEntry(e)

	return nil
}

// NewAddFactoidAddress creates a new AddFactoidAddress
func NewAddFactoidAddress(chainID interfaces.IHash, add interfaces.IAddress) (e *AddFactoidAddress) {
	e = new(AddFactoidAddress)
	e.Init()
	e.IdentityChainID = chainID
	e.FactoidAddress = add
	return
}

// Type returns the hardcoded TYPE_ADD_FACTOID_ADDRESS
func (e *AddFactoidAddress) Type() byte {
	return constants.TYPE_ADD_FACTOID_ADDRESS
}

// SortedIdentity returns the server identity of for the AddFactoidAddress
func (e *AddFactoidAddress) SortedIdentity() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AddFactoidAddress.SortedIdentity() saw an interface that was nil")
		}
	}()

	return e.IdentityChainID
}

// MarshalBinary marshals the AddFactoidAddress
func (e *AddFactoidAddress) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "AddFactoidAddress.MarshalBinary err:%v", *pe)
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

	err = bodybuf.PushBinaryMarshallable(e.FactoidAddress)
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

// UnmarshalBinaryData unmarshals the input data into this AddFactoidAddress
func (e *AddFactoidAddress) UnmarshalBinaryData(data []byte) ([]byte, error) {
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
			"Error: AddFactoidAddress.UnmarshalBinary: body size %d is larger "+
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

	err = bodyBuf.PopBinaryMarshallable(e.FactoidAddress)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinary unmarshals the input data into this AddFactoidAddress
func (e *AddFactoidAddress) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *AddFactoidAddress) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *AddFactoidAddress) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

// IsInterpretable always returns false
func (e *AddFactoidAddress) IsInterpretable() bool {
	return false
}

// Interpret always returns the empty string ""
func (e *AddFactoidAddress) Interpret() string {
	return ""
}

// Hash marshals the object and takes its hash
func (e *AddFactoidAddress) Hash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AddFactoidAddress.Hash() saw an interface that was nil")
		}
	}()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
