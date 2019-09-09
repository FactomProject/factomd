package adminBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// AddReplaceMatryoshkaHash is a DEPRECATED admin block entry that adds a Matryoshka Hash to the specified server.
// According to 'Who', the Matryoshka hash is not used any longer, although its unclear if it may make an appearance
// in older sections of the block chain.
type AddReplaceMatryoshkaHash struct {
	AdminIDType     uint32           `json:"adminidtype"` // the type of action in this admin block entry: uint32(TYPE_ADD_MATRYOSHKA)
	IdentityChainID interfaces.IHash `json:"identitychainid"`
	MHash           interfaces.IHash `json:"mhash"`
}

var _ interfaces.Printable = (*AddReplaceMatryoshkaHash)(nil)
var _ interfaces.BinaryMarshallable = (*AddReplaceMatryoshkaHash)(nil)
var _ interfaces.IABEntry = (*AddReplaceMatryoshkaHash)(nil)

// Init initializes any nil hashes to the zero hash and sets its type
func (e *AddReplaceMatryoshkaHash) Init() {
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}
	if e.MHash == nil {
		e.MHash = primitives.NewZeroHash()
	}
	e.AdminIDType = uint32(e.Type())
}

// String returns this objects string
func (e *AddReplaceMatryoshkaHash) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %35s -- %17s %8x %12s %8s",
		"AddReplaceMatryoshkaHash",
		"IdentityChainID", e.IdentityChainID.Bytes()[3:6],
		"MHash", e.MHash.String()[:8]))
	return (string)(out.DeepCopyBytes())
}

// Type returns the hardcoded TYPE_ADD_MATRYOSHKA
func (e *AddReplaceMatryoshkaHash) Type() byte {
	return constants.TYPE_ADD_MATRYOSHKA
}

// UpdateState updates factomd's state to include information from this object
func (e *AddReplaceMatryoshkaHash) UpdateState(state interfaces.IState) error {
	e.Init()
	state.UpdateAuthorityFromABEntry(e)
	return nil
}

// NewAddReplaceMatryoshkaHash creates a new AddReplaceMatryoshkaHash with the input values
func NewAddReplaceMatryoshkaHash(identityChainID interfaces.IHash, mHash interfaces.IHash) *AddReplaceMatryoshkaHash {
	e := new(AddReplaceMatryoshkaHash)
	e.IdentityChainID = identityChainID
	e.MHash = mHash
	return e
}

// SortedIdentity returns the server identity chain associated with this action
func (e *AddReplaceMatryoshkaHash) SortedIdentity() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AddReplaceMatryoshkaHash.SortedIdentity() saw an interface that was nil")
		}
	}()

	return e.IdentityChainID
}

// MarshalBinary marshals this object
func (e *AddReplaceMatryoshkaHash) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "AddReplaceMatryoshkaHash.MarshalBinary err:%v", *pe)
		}
	}(&err)
	e.Init()
	var buf primitives.Buffer

	err = buf.PushByte(e.Type())
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.IdentityChainID)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.MHash)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input data into this object
func (e *AddReplaceMatryoshkaHash) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	buf := primitives.NewBuffer(data)
	b, err := buf.PopByte()
	if err != nil {
		return nil, err
	}
	if b != e.Type() {
		return nil, fmt.Errorf("Invalid Entry type")
	}

	e.IdentityChainID = new(primitives.Hash)
	err = buf.PopBinaryMarshallable(e.IdentityChainID)
	if err != nil {
		return nil, err
	}
	e.MHash = new(primitives.Hash)
	err = buf.PopBinaryMarshallable(e.MHash)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinary unmarshals the input data into this object
func (e *AddReplaceMatryoshkaHash) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *AddReplaceMatryoshkaHash) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *AddReplaceMatryoshkaHash) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

// IsInterpretable always returns false
func (e *AddReplaceMatryoshkaHash) IsInterpretable() bool {
	return false
}

// Interpret always returns the empty string ""
func (e *AddReplaceMatryoshkaHash) Interpret() string {
	return ""
}

// Hash marshals this object and computes its hash
func (e *AddReplaceMatryoshkaHash) Hash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AddReplaceMatryoshkaHash.Hash() saw an interface that was nil")
		}
	}()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
