package adminBlock

import (
	"bytes"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type RevealMatryoshkaHash struct {
	IdentityChainID interfaces.IHash
	MHash           interfaces.IHash
}

var _ interfaces.Printable = (*RevealMatryoshkaHash)(nil)
var _ interfaces.BinaryMarshallable = (*RevealMatryoshkaHash)(nil)
var _ interfaces.IABEntry = (*RevealMatryoshkaHash)(nil)

func (m *RevealMatryoshkaHash) Type() byte {
	return constants.TYPE_REVEAL_MATRYOSHKA
}

func NewRevealMatryoshkaHash(identityChainID interfaces.IHash, mHash interfaces.IHash) *RevealMatryoshkaHash {
	e := new(RevealMatryoshkaHash)
	e.IdentityChainID = identityChainID
	e.MHash = mHash
	return e
}

func (c *RevealMatryoshkaHash) UpdateState(state interfaces.IState) error {

	return nil
}

func (e *RevealMatryoshkaHash) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer

	buf.Write([]byte{e.Type()})
	buf.Write(e.IdentityChainID.Bytes())
	buf.Write(e.MHash.Bytes())

	return buf.DeepCopyBytes(), nil
}

func (e *RevealMatryoshkaHash) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Reveal Matryoshka Hash: %v", r)
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
	e.MHash = new(primitives.Hash)
	newData, err = e.MHash.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	return
}

func (e *RevealMatryoshkaHash) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *RevealMatryoshkaHash) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *RevealMatryoshkaHash) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *RevealMatryoshkaHash) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *RevealMatryoshkaHash) String() string {
	str, _ := e.JSONString()
	return str
}

func (e *RevealMatryoshkaHash) IsInterpretable() bool {
	return false
}

func (e *RevealMatryoshkaHash) Interpret() string {
	return ""
}

func (e *RevealMatryoshkaHash) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
