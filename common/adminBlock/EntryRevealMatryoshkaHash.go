package adminBlock

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type RevealMatryoshkaHash struct {
	IdentityChainID interfaces.IHash `json:"identitychainid"`
	MHash           interfaces.IHash `json:"mhash"`
}

var _ interfaces.Printable = (*RevealMatryoshkaHash)(nil)
var _ interfaces.BinaryMarshallable = (*RevealMatryoshkaHash)(nil)
var _ interfaces.IABEntry = (*RevealMatryoshkaHash)(nil)

func (e *RevealMatryoshkaHash) Init() {
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}
	if e.MHash == nil {
		e.MHash = primitives.NewZeroHash()
	}
}

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
	c.Init()
	return nil
}

func (e *RevealMatryoshkaHash) MarshalBinary() ([]byte, error) {
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
	err = buf.PushBinaryMarshallable(e.MHash)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

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

func (e *RevealMatryoshkaHash) String() string {
	e.Init()
	str := fmt.Sprintf("    E: %35s -- %17s %8x %12s %x",
		"RevealMatryoshkaHash",
		"IdentityChainID", e.IdentityChainID.Bytes()[3:5],
		"Hash", e.MHash.Bytes()[:5])
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
