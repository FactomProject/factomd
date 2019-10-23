package adminBlock

import (
	"fmt"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// CancelCoinbaseDescriptor Entry -------------------------
type CancelCoinbaseDescriptor struct {
	AdminIDType      uint32 `json:"adminidtype"`
	DescriptorHeight uint32 `json:"descriptor_height"`
	DescriptorIndex  uint32 `json:descriptor_index`

	// Not marshalled
	hash interfaces.IHash // cache
}

var _ interfaces.IABEntry = (*CancelCoinbaseDescriptor)(nil)
var _ interfaces.BinaryMarshallable = (*CancelCoinbaseDescriptor)(nil)

func (e *CancelCoinbaseDescriptor) Init() {
	e.AdminIDType = uint32(e.Type())
}

func (a *CancelCoinbaseDescriptor) IsSameAs(b *CancelCoinbaseDescriptor) bool {
	if a.Type() != b.Type() {
		return false
	}

	if a.DescriptorHeight != b.DescriptorHeight {
		return false
	}

	if a.DescriptorIndex != b.DescriptorIndex {
		return false
	}

	return true
}

func (e *CancelCoinbaseDescriptor) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %20s -- %17s %d %17s %d",
		"CoinbaseDescriptorCancel",
		"Height", e.DescriptorHeight,
		"Index", e.DescriptorIndex))
	return (string)(out.DeepCopyBytes())
}

func (c *CancelCoinbaseDescriptor) UpdateState(state interfaces.IState) error {
	c.Init()
	state.UpdateAuthorityFromABEntry(c)
	return nil
}

func NewCancelCoinbaseDescriptor(height, index uint32) *CancelCoinbaseDescriptor {
	e := new(CancelCoinbaseDescriptor)
	e.Init()
	e.DescriptorHeight = height
	e.DescriptorIndex = index
	return e
}

func (e *CancelCoinbaseDescriptor) Type() byte {
	return constants.TYPE_COINBASE_DESCRIPTOR_CANCEL
}

// SortedIdentity has no identity to sort by, so we will just use the hash of the cancel.
func (e *CancelCoinbaseDescriptor) SortedIdentity() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("CancelCoinbaseDescriptor.SortedIdentity() saw an interface that was nil")
		}
	}()

	return e.Hash()
}

func (e *CancelCoinbaseDescriptor) MarshalBinary() ([]byte, error) {
	e.Init()
	var buf primitives.Buffer

	err := buf.PushByte(e.Type())
	if err != nil {
		return nil, err
	}

	// Need the size of the body
	var bodybuf primitives.Buffer
	err = bodybuf.PushVarInt(uint64(e.DescriptorHeight))
	if err != nil {
		return nil, err
	}

	err = bodybuf.PushVarInt(uint64(e.DescriptorIndex))
	if err != nil {
		return nil, err
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

func (e *CancelCoinbaseDescriptor) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	e.Init()

	t, err := buf.PopByte()
	if err != nil {
		return nil, err
	}

	if t != e.Type() {
		return nil, fmt.Errorf("Invalid Entry type")
	}

	bodyLimit := uint64(buf.Len())
	bodySize, err := buf.PopVarInt()
	if err != nil {
		return nil, err
	}
	if bodySize > bodyLimit {
		return nil, fmt.Errorf(
			"Error: CancelCoinbaseDescriptor.UnmarshalBinary: body size %d is "+
				"larger than binary size %d. (uint underflow?)",
			bodySize, bodyLimit,
		)
	}

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

	height, err := bodyBuf.PopVarInt()
	if err != nil {
		return nil, err
	}

	e.DescriptorHeight = uint32(height)

	index, err := bodyBuf.PopVarInt()
	if err != nil {
		return nil, err
	}

	e.DescriptorIndex = uint32(index)
	if bodyBuf.Len() != 0 {
		return nil, fmt.Errorf("%d bytes remain in body", bodyBuf.Len())
	}

	return buf.DeepCopyBytes(), nil
}

func (e *CancelCoinbaseDescriptor) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *CancelCoinbaseDescriptor) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

func (e *CancelCoinbaseDescriptor) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

func (e *CancelCoinbaseDescriptor) IsInterpretable() bool {
	return false
}

func (e *CancelCoinbaseDescriptor) Interpret() string {
	return ""
}

func (e *CancelCoinbaseDescriptor) Hash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("CancelCoinbaseDescriptor.Hash() saw an interface that was nil")
		}
	}()

	if e.hash == nil {
		bin, err := e.MarshalBinary()
		if err != nil {
			panic(err)
		}
		e.hash = primitives.Sha(bin)
	}
	return e.hash
}
