package adminBlock

import (
	"bytes"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// ForwardCompatibleEntry Entry -------------------------
type ForwardCompatibleEntry struct {
	AdminIDType uint32 `json:"adminidtype"`
	Size        uint32
	Data        []byte
}

var _ interfaces.IABEntry = (*CoinbaseDescriptor)(nil)
var _ interfaces.BinaryMarshallable = (*CoinbaseDescriptor)(nil)

func (e *ForwardCompatibleEntry) Init() {
	e.AdminIDType = uint32(e.Type())
}

func (a *ForwardCompatibleEntry) IsSameAs(b *ForwardCompatibleEntry) bool {
	if a.Type() != b.Type() {
		return false
	}

	if a.Size != b.Size {
		return false
	}

	if bytes.Compare(a.Data, b.Data) != 0 {
		return false
	}

	return true
}

func (e *ForwardCompatibleEntry) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %20s -- %17s %d",
		"ForwardCompatibleEntry",
		"Size", e.Size))
	return (string)(out.DeepCopyBytes())
}

func (c *ForwardCompatibleEntry) UpdateState(state interfaces.IState) error {
	c.Init()
	return nil
}

func NewForwardCompatibleEntry(size uint32) (e *ForwardCompatibleEntry) {
	e = new(ForwardCompatibleEntry)
	e.Init()
	e.Size = size
	return
}

func (e *ForwardCompatibleEntry) Type() byte {
	return byte(e.AdminIDType)
}

func (e *ForwardCompatibleEntry) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ForwardCompatibleEntry.MarshalBinary err:%v", *pe)
		}
	}(&err)
	e.Init()
	var buf primitives.Buffer

	err = buf.PushByte(byte(e.AdminIDType))
	if err != nil {
		return nil, err
	}

	err = buf.PushVarInt(uint64(e.Size))
	if err != nil {
		return nil, err
	}

	err = buf.Push(e.Data)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *ForwardCompatibleEntry) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	e.Init()

	t, err := buf.PopByte()
	if err != nil {
		return nil, err
	}
	e.AdminIDType = uint32(t)

	if t < 0x09 {
		return nil, fmt.Errorf("Invalid Entry type, must be > 0x09")
	}

	bodyLimit := uint64(buf.Len())
	bodySize, err := buf.PopVarInt()
	if err != nil {
		return nil, err
	}
	if bodySize > bodyLimit {
		return nil, fmt.Errorf(
			"Error: ForwardCompatibleEntry.UnmarshalBinary: body size %d is "+
				"larger than binary size %d. (uint underflow?)",
			bodySize, bodyLimit,
		)
	}
	e.Size = uint32(bodySize)

	body := make([]byte, bodySize)
	n, err := buf.Read(body)
	if err != nil {
		return nil, err
	}

	if uint64(n) != bodySize {
		return nil, fmt.Errorf("Expected to read %d bytes, but got %d", bodySize, n)
	}

	bodyBuf := primitives.NewBuffer(body)

	if uint64(n) != bodySize {
		return nil, fmt.Errorf("Unable to unmarshal body")
	}

	e.Data = bodyBuf.Bytes()

	return buf.DeepCopyBytes(), nil
}

func (e *ForwardCompatibleEntry) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *ForwardCompatibleEntry) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

func (e *ForwardCompatibleEntry) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

func (e *ForwardCompatibleEntry) IsInterpretable() bool {
	return false
}

func (e *ForwardCompatibleEntry) Interpret() string {
	return ""
}

func (e *ForwardCompatibleEntry) Hash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "ForwardCompatibleEntry.Hash") }()
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
