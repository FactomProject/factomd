package adminBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// AddFederatedServerBitcoinAnchorKey is an admin block entry adds a Bitcoin public key hash to the federated server
type AddFederatedServerBitcoinAnchorKey struct {
	AdminIDType     uint32                 `json:"adminidtype"`     // the type of action in this admin block entry: uint32(TYPE_ADD_BTC_ANCHOR_KEY)
	IdentityChainID interfaces.IHash       `json:"identitychainid"` // the server identity chain id affected by this action
	KeyPriority     byte                   `json:"keypriority"`     //
	KeyType         byte                   `json:"keytype"`         // 0=P2PKH 1=P2SH, "Pay 2 PublicKey Hash" or "Pay 2 Script Hash"
	ECDSAPublicKey  primitives.ByteSlice20 `json:"ecdsapublickey"`  // the bitcoin public key
}

var _ interfaces.IABEntry = (*AddFederatedServerBitcoinAnchorKey)(nil)
var _ interfaces.BinaryMarshallable = (*AddFederatedServerBitcoinAnchorKey)(nil)

// Init initializes any nil hashes to the zero hash and sets the object's type
func (e *AddFederatedServerBitcoinAnchorKey) Init() {
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}
	e.AdminIDType = uint32(e.Type())
}

// SortedIdentity returns the server identity chain id
func (e *AddFederatedServerBitcoinAnchorKey) SortedIdentity() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AddFederatedServerBitcoinAnchorKey.SortedIdentity() saw an interface that was nil")
		}
	}()

	return e.IdentityChainID
}

// String returns this objects string
func (e *AddFederatedServerBitcoinAnchorKey) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %35s -- %17s %8x %12s %8x %12s %8x %12s %8s",
		"AddFederatedServerBitcoinAnchorKey",
		"IdentityChainID", e.IdentityChainID.Bytes()[3:6],
		"KeyPriority", e.KeyPriority,
		"KeyType", e.KeyType,
		"ECDSAPublicKey", e.ECDSAPublicKey.String()[:8]))
	return (string)(out.DeepCopyBytes())
}

// UpdateState updates the factomd state with information about this object
func (e *AddFederatedServerBitcoinAnchorKey) UpdateState(state interfaces.IState) error {
	e.Init()
	state.UpdateAuthorityFromABEntry(e)
	return nil
}

// NewAddFederatedServerBitcoinAnchorKey returns a new object containing the input data
func NewAddFederatedServerBitcoinAnchorKey(identityChainID interfaces.IHash, keyPriority byte, keyType byte, ecdsaPublicKey primitives.ByteSlice20) (e *AddFederatedServerBitcoinAnchorKey) {
	e = new(AddFederatedServerBitcoinAnchorKey)
	e.IdentityChainID = identityChainID
	e.KeyPriority = keyPriority
	e.KeyType = keyType
	e.ECDSAPublicKey = ecdsaPublicKey
	return
}

// Type returns the hardcoded TYPE_ADD_BTC_ANCHOR_KEY
func (e *AddFederatedServerBitcoinAnchorKey) Type() byte {
	return constants.TYPE_ADD_BTC_ANCHOR_KEY
}

// MarshalBinary marshals this object
func (e *AddFederatedServerBitcoinAnchorKey) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "AddFederatedServerBitcoinAnchorKey.MarshalBinary err:%v", *pe)
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
	err = buf.PushByte(e.KeyType)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(&e.ECDSAPublicKey)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input data into this object
func (e *AddFederatedServerBitcoinAnchorKey) UnmarshalBinaryData(data []byte) ([]byte, error) {
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
	if err != nil {
		return nil, err
	}
	e.KeyType, err = buf.PopByte()
	if err != nil {
		return nil, err
	}
	if e.KeyType != 0 && e.KeyType != 1 {
		return nil, fmt.Errorf("Invalid KeyType, found %d", e.KeyType)
	}
	err = buf.PopBinaryMarshallable(&e.ECDSAPublicKey)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinary unmarshals the input data into this object
func (e *AddFederatedServerBitcoinAnchorKey) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *AddFederatedServerBitcoinAnchorKey) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *AddFederatedServerBitcoinAnchorKey) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

// IsInterpretable always returns false
func (e *AddFederatedServerBitcoinAnchorKey) IsInterpretable() bool {
	return false
}

// Interpret always returns the empty string ""
func (e *AddFederatedServerBitcoinAnchorKey) Interpret() string {
	return ""
}

// Hash marshals the object and computes its hash
func (e *AddFederatedServerBitcoinAnchorKey) Hash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AddFederatedServerBitcoinAnchorKey.Hash() saw an interface that was nil")
		}
	}()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
