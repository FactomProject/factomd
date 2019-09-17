package adminBlock

import (
	"fmt"
	"os"
	"reflect"

	"bytes"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// ForwardCompatibleEntry is an admin block entry with generic size and generic data
type ForwardCompatibleEntry struct {
	AdminIDType uint32 `json:"adminidtype"` // the type of action in this admin block entry
	Size        uint32 `json:"size"`        // the length of the byte array
	Data        []byte `json:"data"`        // the data for this entry
}

var _ interfaces.IABEntry = (*CoinbaseDescriptor)(nil)
var _ interfaces.BinaryMarshallable = (*CoinbaseDescriptor)(nil)

// Init initializes the object type
func (e *ForwardCompatibleEntry) Init() {
	e.AdminIDType = uint32(e.Type())
}

// IsSameAs returns true iff the input object is identical to this object
func (e *ForwardCompatibleEntry) IsSameAs(b *ForwardCompatibleEntry) bool {
	if e.Type() != b.Type() {
		return false
	}

	if e.Size != b.Size {
		return false
	}

	if bytes.Compare(e.Data, b.Data) != 0 {
		return false
	}

	return true
}

// String returns this objects string
func (e *ForwardCompatibleEntry) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %20s -- %17s %d",
		"ForwardCompatibleEntry",
		"Size", e.Size))
	return (string)(out.DeepCopyBytes())
}

// UpdateState does not interact with the input state object, merely initilizes this object and returns nil
func (e *ForwardCompatibleEntry) UpdateState(state interfaces.IState) error {
	e.Init()
	return nil
}

// NewForwardCompatibleEntry creates a new ForwardCompatibleEntry of a given size
func NewForwardCompatibleEntry(size uint32) (e *ForwardCompatibleEntry) {
	e = new(ForwardCompatibleEntry)
	e.Init()
	e.Size = size
	return
}

// Type returns the AdminIDType for this object
func (e *ForwardCompatibleEntry) Type() byte {
	return byte(e.AdminIDType)
}

// MarshalBinary marshals this object
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

// UnmarshalBinaryData unmarshals the input data into this object
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

// UnmarshalBinary unmarshals the input data into this object
func (e *ForwardCompatibleEntry) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *ForwardCompatibleEntry) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *ForwardCompatibleEntry) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

// IsInterpretable always returns false
func (e *ForwardCompatibleEntry) IsInterpretable() bool {
	return false
}

// Interpret always returns the empty string ""
func (e *ForwardCompatibleEntry) Interpret() string {
	return ""
}

// Hash marshals the object and computes its hash
func (e *ForwardCompatibleEntry) Hash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("ForwardCompatibleEntry.Hash() saw an interface that was nil")
		}
	}()
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
