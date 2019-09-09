package adminBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// AddFederatedServerSigningKey is an admin block entry that adds a public signing key to a federated server
type AddFederatedServerSigningKey struct {
	AdminIDType     uint32               `json:"adminidtype"`     // the type of action in this admin block entry: uint32(TYPE_ADD_FED_SERVER_KEY)
	IdentityChainID interfaces.IHash     `json:"identitychainid"` // the federated server identity chain
	KeyPriority     byte                 `json:"keypriority"`     //
	PublicKey       primitives.PublicKey `json:"publickey"`       // the public key being added
	DBHeight        uint32               `json:"dbheight"`        // the directory block height this action activates
}

var _ interfaces.IABEntry = (*AddFederatedServerSigningKey)(nil)
var _ interfaces.BinaryMarshallable = (*AddFederatedServerSigningKey)(nil)

// Init initializes any nil hashs to the zero hash and sets the object type
func (e *AddFederatedServerSigningKey) Init() {
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}
	e.AdminIDType = uint32(e.Type())
}

// UpdateState updates factomd's state to include information from this object
func (e *AddFederatedServerSigningKey) UpdateState(state interfaces.IState) error {
	e.Init()
	state.UpdateAuthorityFromABEntry(e)
	return nil
}

// SortedIdentity returns the server identity chain
func (e *AddFederatedServerSigningKey) SortedIdentity() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AddFederatedServerSigningKey.SortedIdentity() saw an interface that was nil")
		}
	}()

	return e.IdentityChainID
}

// String returns the objects string
func (e *AddFederatedServerSigningKey) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %35s -- %17s %8x %12s %8x %12s %8s %12s %d",
		"AddFederatedServerSigningKey",
		"IdentityChainID", e.IdentityChainID.Bytes()[3:6],
		"KeyPriority", e.KeyPriority,
		"PublicKey", e.PublicKey.String()[:8],
		"DBHeight", e.DBHeight))
	return (string)(out.DeepCopyBytes())
}

// NewAddFederatedServerSigningKey creates a new AddFederatedServerSigningKey with the given inputs
func NewAddFederatedServerSigningKey(identityChainID interfaces.IHash, keyPriority byte, publicKey primitives.PublicKey, height uint32) (e *AddFederatedServerSigningKey) {
	e = new(AddFederatedServerSigningKey)
	e.IdentityChainID = identityChainID
	e.KeyPriority = keyPriority
	e.PublicKey = publicKey
	e.DBHeight = height
	return
}

// Type returns the hardcoded TYPE_ADD_FED_SERVER_KEY
func (e *AddFederatedServerSigningKey) Type() byte {
	return constants.TYPE_ADD_FED_SERVER_KEY
}

// MarshalBinary marshals the object
func (e *AddFederatedServerSigningKey) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "AddFederatedServerSigningKey.MarshalBinary err:%v", *pe)
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
	err = buf.PushByte(e.KeyPriority)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(&e.PublicKey)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(e.DBHeight)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input data into this object
func (e *AddFederatedServerSigningKey) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	b, err := buf.PopByte()
	if b != e.Type() {
		return nil, fmt.Errorf("Invalid Entry type")
	}

	e.IdentityChainID = new(primitives.Hash)
	err = buf.PopBinaryMarshallable(e.IdentityChainID)
	if err != nil {
		return nil, err
	}
	e.KeyPriority, err = buf.PopByte()
	err = buf.PopBinaryMarshallable(&e.PublicKey)
	if err != nil {
		return nil, err
	}
	e.DBHeight, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinary unmarshals the input data into this object
func (e *AddFederatedServerSigningKey) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *AddFederatedServerSigningKey) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *AddFederatedServerSigningKey) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

// IsInterpretable always returns false
func (e *AddFederatedServerSigningKey) IsInterpretable() bool {
	return false
}

// Interpret always returns the empty string ""
func (e *AddFederatedServerSigningKey) Interpret() string {
	return ""
}

// Hash marshals the object and computes its hash
func (e *AddFederatedServerSigningKey) Hash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AddFederatedServerSigningKey.Hash() saw an interface that was nil")
		}
	}()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
