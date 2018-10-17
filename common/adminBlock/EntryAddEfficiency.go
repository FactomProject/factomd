package adminBlock

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// AddEfficiency Entry -------------------------
type AddEfficiency struct {
	AdminIDType     uint32 `json:"adminidtype"`
	IdentityChainID interfaces.IHash
	Efficiency      uint16
}

var _ interfaces.IABEntry = (*AddEfficiency)(nil)
var _ interfaces.BinaryMarshallable = (*AddEfficiency)(nil)

func (e *AddEfficiency) Init() {
	e.AdminIDType = uint32(e.Type())
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}
}

func (a *AddEfficiency) IsSameAs(b *AddEfficiency) bool {
	if a.Type() != b.Type() {
		return false
	}

	if !a.IdentityChainID.IsSameAs(b.IdentityChainID) {
		return false
	}

	if a.Efficiency != b.Efficiency {
		return false
	}

	return true
}

func (e *AddEfficiency) SortedIdentity() interfaces.IHash {
	return e.IdentityChainID
}

func (e *AddEfficiency) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %20s -- %17s %8x %12s %d",
		"AddAuditServer",
		"IdentityChainID", e.IdentityChainID.Bytes()[3:6],
		"Efficiency", e.Efficiency))
	return (string)(out.DeepCopyBytes())
}

func (c *AddEfficiency) UpdateState(state interfaces.IState) error {
	c.Init()
	//state.AddAuditServer(c.DBHeight, c.IdentityChainID)
	state.UpdateAuthorityFromABEntry(c)

	return nil
}

func NewAddEfficiency(chainID interfaces.IHash, efficiency uint16) (e *AddEfficiency) {
	e = new(AddEfficiency)
	e.Init()
	e.IdentityChainID = chainID
	if efficiency > 10000 {
		efficiency = 10000
	}
	e.Efficiency = efficiency
	return
}

func (e *AddEfficiency) Type() byte {
	return constants.TYPE_ADD_FACTOID_EFFICIENCY
}

func (e *AddEfficiency) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "AddEfficiency.MarshalBinary err:%v", *pe)
		}
	}(&err)
	e.Init()
	var buf primitives.Buffer

	err = buf.PushByte(e.Type())
	if err != nil {
		return nil, err
	}

	// Need the size of the body
	var bodybuf primitives.Buffer
	err = bodybuf.PushIHash(e.IdentityChainID)
	if err != nil {
		return nil, err
	}

	err = bodybuf.PushUInt16(e.Efficiency)
	if err != nil {
		return nil, err
	}
	// end body

	err = buf.PushVarInt(uint64(bodybuf.Len()))
	if err != nil {
		return nil, err
	}

	err = buf.Push(bodybuf.Bytes())
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *AddEfficiency) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	e.Init()

	b, err := buf.PopByte()
	if err != nil {
		return nil, err
	}

	if b != e.Type() {
		return nil, fmt.Errorf("Invalid Entry type")
	}

	bl, err := buf.PopVarInt()
	if err != nil {
		return nil, err
	}
	if bl > 1000 {
		// TODO: replace this message with a proper error
		return nil, fmt.Errorf("Error: AddEfficiency.UnmarshalBinary: body length too long (uint underflow?)")
	}

	body := make([]byte, bl)
	n, err := buf.Read(body)
	if err != nil {
		return nil, err
	}

	if uint64(n) != bl {
		return nil, fmt.Errorf("Expected to read %d bytes, but got %d", bl, n)
	}

	bodyBuf := primitives.NewBuffer(body)

	if uint64(n) != bl {
		return nil, fmt.Errorf("Unable to unmarshal body")
	}

	e.IdentityChainID, err = bodyBuf.PopIHash()
	if err != nil {
		return nil, err
	}

	e.Efficiency, err = bodyBuf.PopUInt16()
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *AddEfficiency) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *AddEfficiency) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

func (e *AddEfficiency) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

func (e *AddEfficiency) IsInterpretable() bool {
	return false
}

func (e *AddEfficiency) Interpret() string {
	return ""
}

func (e *AddEfficiency) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
