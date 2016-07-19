package adminBlock

import (
	"bytes"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type AddReplaceMatryoshkaHash struct {
	IdentityChainID interfaces.IHash
	MHash           interfaces.IHash
}

var _ interfaces.Printable = (*AddReplaceMatryoshkaHash)(nil)
var _ interfaces.BinaryMarshallable = (*AddReplaceMatryoshkaHash)(nil)
var _ interfaces.IABEntry = (*AddReplaceMatryoshkaHash)(nil)

func (e *AddReplaceMatryoshkaHash) String() string {
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %35s -- %17s %8x %12s %8s\n",
		"AddReplaceMatryoshkaHash",
		"IdentityChainID", e.IdentityChainID.Bytes()[:4],
		"MHash", e.MHash.String()[:8]))
	return (string)(out.DeepCopyBytes())
}

func (m *AddReplaceMatryoshkaHash) Type() byte {
	return constants.TYPE_ADD_MATRYOSHKA
}

func (c *AddReplaceMatryoshkaHash) UpdateState(state interfaces.IState) {
	state.UpdateAuthorityFromABEntry(c)
}

func NewAddReplaceMatryoshkaHash(identityChainID interfaces.IHash, mHash interfaces.IHash) *AddReplaceMatryoshkaHash {
	e := new(AddReplaceMatryoshkaHash)
	e.IdentityChainID = identityChainID
	e.MHash = mHash
	return e
}

func (e *AddReplaceMatryoshkaHash) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer

	buf.Write([]byte{e.Type()})
	buf.Write(e.IdentityChainID.Bytes())
	buf.Write(e.MHash.Bytes())

	return buf.DeepCopyBytes(), nil
}

func (e *AddReplaceMatryoshkaHash) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Add Replace Matryoshka Hash: %v", r)
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

func (e *AddReplaceMatryoshkaHash) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *AddReplaceMatryoshkaHash) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *AddReplaceMatryoshkaHash) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *AddReplaceMatryoshkaHash) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *AddReplaceMatryoshkaHash) IsInterpretable() bool {
	return false
}

func (e *AddReplaceMatryoshkaHash) Interpret() string {
	return ""
}

func (e *AddReplaceMatryoshkaHash) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
