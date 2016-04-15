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

}

// Create a new DB Signature Entry
func NewAddFederatedServerSigningKey(identityChainID interfaces.IHash) (e *AddFederatedServerSigningKey) {
	e = new(AddFederatedServerSigningKey)
	e.IdentityChainID = primitives.NewHash(identityChainID.Bytes())
	return
}

func (e *AddFederatedServerSigningKey) Type() byte {
	return constants.TYPE_ADD_FED_SERVER
}

func (e *AddFederatedServerSigningKey) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	buf.Write([]byte{e.Type()})

	data, err = e.IdentityChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	return buf.Bytes(), nil
}

func (e *AddFederatedServerSigningKey) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	newData = data
	newData = newData[1:]

	e.IdentityChainID = new(primitives.Hash)
	newData, err = e.IdentityChainID.UnmarshalBinaryData(newData)
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

func (e *AddFederatedServerSigningKey) String() string {
	str := fmt.Sprintf("Add Server with Identity Chain ID = %x", e.IdentityChainID.Bytes()[:5])
	return str
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
