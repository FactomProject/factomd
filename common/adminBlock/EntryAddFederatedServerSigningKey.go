package adminBlock

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// DB Signature Entry -------------------------
type AddFederatedServerSigningKey struct {
	IdentityChainID interfaces.IHash
	KeyPriority     byte
	PublicKey       primitives.PublicKey
}

var _ interfaces.IABEntry = (*AddFederatedServerSigningKey)(nil)
var _ interfaces.BinaryMarshallable = (*AddFederatedServerSigningKey)(nil)

func (c *AddFederatedServerSigningKey) UpdateState(state interfaces.IState) {
	state.UpdateAuthorityFromABEntry(c)
}

func (e *AddFederatedServerSigningKey) String() string {
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %35s -- %17s %8x %12s %8x %12s %8s\n",
		"AddFederatedServerSigningKey",
		"IdentityChainID", e.IdentityChainID.Bytes()[:4],
		"KeyPriority", e.KeyPriority,
		"PublicKey", e.PublicKey.String()[:8]))
	return (string)(out.DeepCopyBytes())
}

// Create a new DB Signature Entry
func NewAddFederatedServerSigningKey(identityChainID interfaces.IHash, keyPriority byte, publicKey primitives.PublicKey) (e *AddFederatedServerSigningKey) {
	e = new(AddFederatedServerSigningKey)
	e.IdentityChainID = identityChainID
	e.KeyPriority = keyPriority
	e.PublicKey = publicKey
	return
}

func (e *AddFederatedServerSigningKey) Type() byte {
	return constants.TYPE_ADD_FED_SERVER_KEY
}

func (e *AddFederatedServerSigningKey) MarshalBinary() ([]byte, error) {
	var buf primitives.Buffer

	buf.Write([]byte{e.Type()})

	data, err := e.IdentityChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	buf.Write([]byte{e.KeyPriority})

	data, err = e.PublicKey.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	return buf.DeepCopyBytes(), nil
}

func (e *AddFederatedServerSigningKey) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Add Federated server Signing Key Entry: %v", r)
		}
	}()

	newData = data
	if newData[0] != e.Type() {
		return nil, fmt.Errorf("Invalid Entry type")
	}
	newData = newData[1:]

	e.IdentityChainID = new(primitives.Hash)
	newData, err = e.IdentityChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	e.KeyPriority, newData = newData[0], newData[1:]

	newData, err = e.PublicKey.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	return
}

func (e *AddFederatedServerSigningKey) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *AddFederatedServerSigningKey) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *AddFederatedServerSigningKey) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *AddFederatedServerSigningKey) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *AddFederatedServerSigningKey) IsInterpretable() bool {
	return false
}

func (e *AddFederatedServerSigningKey) Interpret() string {
	return ""
}

func (e *AddFederatedServerSigningKey) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
