package adminBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// RevealMatryoshkaHash is a type of entry in the admin block which reveals the Matryoshka Hash.
//
// Deprecated: the Matryoshka hash is not used any longer,see comments under AddReplaceMatryoshkaHash
type RevealMatryoshkaHash struct {
	AdminIDType     uint32           `json:"adminidtype"`     //  the type of action in this admin block entry: uint32(TYPE_REVEAL_MATRYOSHKA)
	IdentityChainID interfaces.IHash `json:"identitychainid"` // Server 32 byte identity chain id
	MHash           interfaces.IHash `json:"mhash"`           // the MatryoshkaHash
}

var _ interfaces.Printable = (*RevealMatryoshkaHash)(nil)
var _ interfaces.BinaryMarshallable = (*RevealMatryoshkaHash)(nil)
var _ interfaces.IABEntry = (*RevealMatryoshkaHash)(nil)

// Init initializes the internal hashes to the zero hash if they are nil
func (e *RevealMatryoshkaHash) Init() {
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}
	if e.MHash == nil {
		e.MHash = primitives.NewZeroHash()
	}
	e.AdminIDType = uint32(e.Type())
}

// Type returns a hardcoded TYPE_REVEAL_MATRYOSHKA
func (e *RevealMatryoshkaHash) Type() byte {
	return constants.TYPE_REVEAL_MATRYOSHKA
}

// NewRevealMatryoshkaHash creates a new RevealMatryoshkaHash with the input chain id and mHash
func NewRevealMatryoshkaHash(identityChainID interfaces.IHash, mHash interfaces.IHash) *RevealMatryoshkaHash {
	e := new(RevealMatryoshkaHash)
	e.IdentityChainID = identityChainID
	e.MHash = mHash
	return e
}

// UpdateState initializes internal hashes to the zero hash if they are nil but does not touch the input state
func (e *RevealMatryoshkaHash) UpdateState(state interfaces.IState) error {
	e.Init()
	return nil
}

// MarshalBinary marshals thee RevealMatryoshkaHash
func (e *RevealMatryoshkaHash) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "RevealMatryoshkaHash.MarshalBinary err:%v", *pe)
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

// UnmarshalBinaryData unmarshals the input data into this RevealMatryoshkaHash
func (e *RevealMatryoshkaHash) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
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

// UnmarshalBinary unmarshals the input data into this RevealMatryoshkaHash
func (e *RevealMatryoshkaHash) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte string
func (e *RevealMatryoshkaHash) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *RevealMatryoshkaHash) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

// String returns the string of this RevealMatryoshkaHash
func (e *RevealMatryoshkaHash) String() string {
	e.Init()
	str := fmt.Sprintf("    E: %35s -- %17s %8x %12s %x",
		"RevealMatryoshkaHash",
		"IdentityChainID", e.IdentityChainID.Bytes()[3:6],
		"Hash", e.MHash.Bytes()[:5])
	return str
}

// IsInterpretable always returns false
func (e *RevealMatryoshkaHash) IsInterpretable() bool {
	return false
}

// Interpret always returns the empty string ""
func (e *RevealMatryoshkaHash) Interpret() string {
	return ""
}

// Hash marshals this object and takes its hash
func (e *RevealMatryoshkaHash) Hash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("RevealMatryoshkaHash.Hash() saw an interface that was nil")
		}
	}()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
