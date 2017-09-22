package adminBlock

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// DB Signature Entry -------------------------
type AddFederatedServerBitcoinAnchorKey struct {
	IdentityChainID interfaces.IHash       `json:"identitychainid"`
	KeyPriority     byte                   `json:"keypriority"`
	KeyType         byte                   `json:"keytype"` //0=P2PKH 1=P2SH
	ECDSAPublicKey  primitives.ByteSlice20 `json:"ecdsapublickey"`
}

var _ interfaces.IABEntry = (*AddFederatedServerBitcoinAnchorKey)(nil)
var _ interfaces.BinaryMarshallable = (*AddFederatedServerBitcoinAnchorKey)(nil)

func (e *AddFederatedServerBitcoinAnchorKey) Init() {
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}
}

func (e *AddFederatedServerBitcoinAnchorKey) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %35s -- %17s %8x %12s %8x %12s %8x %12s %8s",
		"AddFederatedServerBitcoinAnchorKey",
		"IdentityChainID", e.IdentityChainID.Bytes()[3:5],
		"KeyPriority", e.KeyPriority,
		"KeyType", e.KeyType,
		"ECDSAPublicKey", e.ECDSAPublicKey.String()[:8]))
	return (string)(out.DeepCopyBytes())
}

func (c *AddFederatedServerBitcoinAnchorKey) UpdateState(state interfaces.IState) error {
	c.Init()
	state.UpdateAuthorityFromABEntry(c)
	return nil
}

// Create a new DB Signature Entry
func NewAddFederatedServerBitcoinAnchorKey(identityChainID interfaces.IHash, keyPriority byte, keyType byte, ecdsaPublicKey primitives.ByteSlice20) (e *AddFederatedServerBitcoinAnchorKey) {
	e = new(AddFederatedServerBitcoinAnchorKey)
	e.IdentityChainID = identityChainID
	e.KeyPriority = keyPriority
	e.KeyType = keyType
	e.ECDSAPublicKey = ecdsaPublicKey
	return
}

func (e *AddFederatedServerBitcoinAnchorKey) Type() byte {
	return constants.TYPE_ADD_BTC_ANCHOR_KEY
}

func (e *AddFederatedServerBitcoinAnchorKey) MarshalBinary() ([]byte, error) {
	e.Init()
	var buf primitives.Buffer

	err := buf.PushByte(e.Type())
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
		return nil, fmt.Errorf("Invalid KeyType")
	}
	err = buf.PopBinaryMarshallable(&e.ECDSAPublicKey)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *AddFederatedServerBitcoinAnchorKey) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *AddFederatedServerBitcoinAnchorKey) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *AddFederatedServerBitcoinAnchorKey) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *AddFederatedServerBitcoinAnchorKey) IsInterpretable() bool {
	return false
}

func (e *AddFederatedServerBitcoinAnchorKey) Interpret() string {
	return ""
}

func (e *AddFederatedServerBitcoinAnchorKey) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
