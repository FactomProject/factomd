package adminBlock

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// CoinbaseDescriptor Entry -------------------------
type CoinbaseDescriptor struct {
	AdminIDType uint32 `json:"adminidtype"`
	Outputs     []interfaces.ITransAddress
}

var _ interfaces.IABEntry = (*CoinbaseDescriptor)(nil)
var _ interfaces.BinaryMarshallable = (*CoinbaseDescriptor)(nil)

func (e *CoinbaseDescriptor) Init() {
	e.AdminIDType = uint32(e.Type())
}

func (a *CoinbaseDescriptor) IsSameAs(b *CoinbaseDescriptor) bool {
	if a.Type() != b.Type() {
		return false
	}

	for i := range a.Outputs {
		if !a.Outputs[i].IsSameAs(b.Outputs[i]) {
			return false
		}
	}
	return true
}

func (e *CoinbaseDescriptor) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %20s -- %17s %d",
		"CoinbaseDescriptor",
		"Number of Outputs", len(e.Outputs)))
	return (string)(out.DeepCopyBytes())
}

func (c *CoinbaseDescriptor) UpdateState(state interfaces.IState) error {
	c.Init()
	return nil
}

func NewCoinbaseDescriptor(outputs []interfaces.ITransAddress) (e *CoinbaseDescriptor) {
	e = new(CoinbaseDescriptor)
	e.Init()
	e.Outputs = outputs
	return
}

func (e *CoinbaseDescriptor) Type() byte {
	return constants.TYPE_COINBASE_DESCRIPTOR
}

func (e *CoinbaseDescriptor) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "CoinbaseDescriptor.MarshalBinary err:%v", *pe)
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
	for _, t := range e.Outputs {
		err = bodybuf.PushBinaryMarshallable(t)
		if err != nil {
			return nil, err
		}
	}

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

func (e *CoinbaseDescriptor) UnmarshalBinaryData(data []byte) ([]byte, error) {
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

	if e.Outputs == nil {
		e.Outputs = make([]interfaces.ITransAddress, 0)
	}

	for {
		if bodyBuf.Len() == 0 {
			break
		}

		it := new(factoid.TransAddress)
		err = bodyBuf.PopBinaryMarshallable(it)
		if err != nil {
			return nil, err
		}

		it.SetUserAddress(primitives.ConvertFctAddressToUserStr(it.Address))

		e.Outputs = append(e.Outputs, it)
	}

	return buf.DeepCopyBytes(), nil
}

func (e *CoinbaseDescriptor) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *CoinbaseDescriptor) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

func (e *CoinbaseDescriptor) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

func (e *CoinbaseDescriptor) IsInterpretable() bool {
	return false
}

func (e *CoinbaseDescriptor) Interpret() string {
	return ""
}

func (e *CoinbaseDescriptor) Hash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "CoinbaseDescriptor.Hash") }()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
