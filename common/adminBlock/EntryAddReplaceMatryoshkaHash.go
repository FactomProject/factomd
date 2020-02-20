package adminBlock

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type AddReplaceMatryoshkaHash struct {
	AdminIDType     uint32           `json:"adminidtype"`
	IdentityChainID interfaces.IHash `json:"identitychainid"`
	MHash           interfaces.IHash `json:"mhash"`
}

var _ interfaces.Printable = (*AddReplaceMatryoshkaHash)(nil)
var _ interfaces.BinaryMarshallable = (*AddReplaceMatryoshkaHash)(nil)
var _ interfaces.IABEntry = (*AddReplaceMatryoshkaHash)(nil)

func (e *AddReplaceMatryoshkaHash) Init() {
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}
	if e.MHash == nil {
		e.MHash = primitives.NewZeroHash()
	}
	e.AdminIDType = uint32(e.Type())
}

func (e *AddReplaceMatryoshkaHash) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %35s -- %17s %8x %12s %8s",
		"AddReplaceMatryoshkaHash",
		"IdentityChainID", e.IdentityChainID.Bytes()[3:6],
		"MHash", e.MHash.String()[:8]))
	return (string)(out.DeepCopyBytes())
}

func (m *AddReplaceMatryoshkaHash) Type() byte {
	return constants.TYPE_ADD_MATRYOSHKA
}

func (c *AddReplaceMatryoshkaHash) UpdateState(state interfaces.IState) error {
	c.Init()
	state.UpdateAuthorityFromABEntry(c)
	return nil
}

func NewAddReplaceMatryoshkaHash(identityChainID interfaces.IHash, mHash interfaces.IHash) *AddReplaceMatryoshkaHash {
	e := new(AddReplaceMatryoshkaHash)
	e.IdentityChainID = identityChainID
	e.MHash = mHash
	return e
}

func (e *AddReplaceMatryoshkaHash) SortedIdentity() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "AddReplaceMatryoshkaHash.SortedIdentity") }()

	return e.IdentityChainID
}

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

func (e *AddReplaceMatryoshkaHash) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *AddReplaceMatryoshkaHash) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

func (e *AddReplaceMatryoshkaHash) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

func (e *AddReplaceMatryoshkaHash) IsInterpretable() bool {
	return false
}

func (e *AddReplaceMatryoshkaHash) Interpret() string {
	return ""
}

func (e *AddReplaceMatryoshkaHash) Hash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "AddReplaceMatryoshkaHash.Hash") }()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
