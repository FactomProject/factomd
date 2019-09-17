package adminBlock

import (
	"fmt"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// CancelCoinbaseDescriptor is an admin block entry which cancels a specific output in a previously created coinbase descriptor.
// The cancelled factoids go to the grant pool intead of the specified address in the previous coinbase descriptor. This entry is
// triggered by a majority of authrity nodes voting for it
type CancelCoinbaseDescriptor struct {
	AdminIDType      uint32 `json:"adminidtype"`       //  the type of action in this admin block entry: uint32(TYPE_COINBASE_DESCRIPTOR_CANCEL)
	DescriptorHeight uint32 `json:"descriptor_height"` // The previous directory block height the original coinbase descriptor is present
	DescriptorIndex  uint32 `json:"descriptor_index"`  // The specific index into the coinbase output array to be cancelled at the directory block height above

	// Not marshalled
	hash interfaces.IHash // cache
}

var _ interfaces.IABEntry = (*CancelCoinbaseDescriptor)(nil)
var _ interfaces.BinaryMarshallable = (*CancelCoinbaseDescriptor)(nil)

// Init initializes the CancelCoinbaseDescriptor to TYPE_COINBASE_DESCRIPTOR_CANCEL
func (e *CancelCoinbaseDescriptor) Init() {
	e.AdminIDType = uint32(e.Type())
}

// IsSameAs returns true iff the input CancelCoinbaseDescriptor is indentical to this one
func (e *CancelCoinbaseDescriptor) IsSameAs(b *CancelCoinbaseDescriptor) bool {
	if e.Type() != b.Type() {
		return false
	}

	if e.DescriptorHeight != b.DescriptorHeight {
		return false
	}

	if e.DescriptorIndex != b.DescriptorIndex {
		return false
	}

	return true
}

// String returns this CancelCoinbaseDescriptor as a string
func (e *CancelCoinbaseDescriptor) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %20s -- %17s %d %17s %d",
		"CoinbaseDescriptorCancel",
		"Height", e.DescriptorHeight,
		"Index", e.DescriptorIndex))
	return (string)(out.DeepCopyBytes())
}

// UpdateState updates the factomd d state with information about the cancelled coinbase request
func (e *CancelCoinbaseDescriptor) UpdateState(state interfaces.IState) error {
	e.Init()
	state.UpdateAuthorityFromABEntry(e)
	return nil
}

// NewCancelCoinbaseDescriptor creates a new CancelCoinbaseDescriptor with the given inputs
func NewCancelCoinbaseDescriptor(height, index uint32) *CancelCoinbaseDescriptor {
	e := new(CancelCoinbaseDescriptor)
	e.Init()
	e.DescriptorHeight = height
	e.DescriptorIndex = index
	return e
}

// Type returns the hardcoded TYPE_COINBASE_DESCRIPTOR_CANCEL
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

// MarshalBinary marshals the the CancelCoinbaseDescriptor object
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

// UnmarshalBinaryData unmarshals the input data into this CancelCoinbaseDescriptor
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

// UnmarshalBinary unmarshals the input data into this CancelCoinbaseDescriptor
func (e *CancelCoinbaseDescriptor) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *CancelCoinbaseDescriptor) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *CancelCoinbaseDescriptor) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

// IsInterpretable always returns false
func (e *CancelCoinbaseDescriptor) IsInterpretable() bool {
	return false
}

// Interpret always returns the empty string ""
func (e *CancelCoinbaseDescriptor) Interpret() string {
	return ""
}

// Hash marshals the CancelCoinbaseDescriptor and computes its hash
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
