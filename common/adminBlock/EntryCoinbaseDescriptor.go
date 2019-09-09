package adminBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// CoinbaseDescriptor is an admin block entry which specifies a future coinbase transaction after 1000 directory blocks have passed.
// The CoinbaseDescriptor entry should occur every 25 blocks, in blocks whose number is divisible by 25.
type CoinbaseDescriptor struct {
	AdminIDType uint32                     `json:"adminidtype"` //  the type of action in this admin block entry: uint32(TYPE_COINBASE_DESCRIPTOR)
	Outputs     []interfaces.ITransAddress // An array containing a pair of factoid_value1 to increase a at a specific factoid_address1, ... to factoid_valueN at factoid_addressN
}

var _ interfaces.IABEntry = (*CoinbaseDescriptor)(nil)
var _ interfaces.BinaryMarshallable = (*CoinbaseDescriptor)(nil)

// Init sets the admin id type to TYPE_COINBASE_DESCRIPTOR
func (e *CoinbaseDescriptor) Init() {
	e.AdminIDType = uint32(e.Type())
}

// IsSameAs returns true iff the input coinbase descriptor is identical to this coinbase descriptor
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

// String returns this coinbase descriptor as a string
func (e *CoinbaseDescriptor) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %20s -- %17s %d",
		"CoinbaseDescriptor",
		"Number of Outputs", len(e.Outputs)))
	return (string)(out.DeepCopyBytes())
}

// UpdateState initializes this descriptor and always returns nil
func (e *CoinbaseDescriptor) UpdateState(state interfaces.IState) error {
	e.Init()
	return nil
}

// NewCoinbaseDescriptor creates a new coinbase descriptor with the input factoid values and addresses
func NewCoinbaseDescriptor(outputs []interfaces.ITransAddress) (e *CoinbaseDescriptor) {
	e = new(CoinbaseDescriptor)
	e.Init()
	e.Outputs = outputs
	return
}

// Type returns the hardcoded TYPE_COINBASE_DESCRIPTOR
func (e *CoinbaseDescriptor) Type() byte {
	return constants.TYPE_COINBASE_DESCRIPTOR
}

// MarshalBinary marshals the coinbase descriptor
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

// UnmarshalBinaryData unmarshals the input data into this coinbase descriptor
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

// UnmarshalBinary unmarshals the input data into this coinbase descriptor
func (e *CoinbaseDescriptor) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *CoinbaseDescriptor) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *CoinbaseDescriptor) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

// IsInterpretable always returns false
func (e *CoinbaseDescriptor) IsInterpretable() bool {
	return false
}

// Interpret always returns the empty string ""
func (e *CoinbaseDescriptor) Interpret() string {
	return ""
}

// Hash marshals the coinbase descriptor and takes its hash
func (e *CoinbaseDescriptor) Hash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("CoinbaseDescriptor.Hash() saw an interface that was nil")
		}
	}()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
