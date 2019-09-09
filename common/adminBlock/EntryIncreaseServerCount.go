package adminBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// IncreaseServerCount is an admin block entry which instructs the factomd state to incrase the server count
type IncreaseServerCount struct {
	AdminIDType uint32 `json:"adminidtype"` //  the type of action in this admin block entry: uint32(TYPE_ADD_SERVER_COUNT)
	Amount      byte   `json:"amount"`      // the number of servers to add to the existing count
}

var _ interfaces.IABEntry = (*IncreaseServerCount)(nil)
var _ interfaces.BinaryMarshallable = (*IncreaseServerCount)(nil)

// NewIncreaseServerCount creates a new IncreaseServerCount with the given input
func NewIncreaseServerCount(num byte) (e *IncreaseServerCount) {
	e = new(IncreaseServerCount)
	e.Amount = num
	return
}

// UpdateState does nothing and returns nil
func (e *IncreaseServerCount) UpdateState(state interfaces.IState) error {
	return nil
}

// Type returns the hardcoded TYPE_ADD_SERVER_COUNT
func (e *IncreaseServerCount) Type() byte {
	return constants.TYPE_ADD_SERVER_COUNT
}

// MarshalBinary marshals the IncreaseServerCount object
func (e *IncreaseServerCount) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "IncreaseServerCount.MarshalBinary err:%v", *pe)
		}
	}(&err)
	var buf primitives.Buffer

	e.AdminIDType = uint32(e.Type())

	err = buf.PushByte(e.Type())
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(e.Amount)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input data into this IncreaseServerCount object
func (e *IncreaseServerCount) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	b, err := buf.PopByte()
	if err != nil {
		return nil, err
	}
	if b != e.Type() {
		return nil, fmt.Errorf("Invalid Entry type")
	}

	e.Amount, err = buf.PopByte()
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinary unmarshals the input data into this IncreaseServerCount object
func (e *IncreaseServerCount) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *IncreaseServerCount) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *IncreaseServerCount) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

// String returns the IncreaseServerCount string
func (e *IncreaseServerCount) String() string {
	str := fmt.Sprintf("    E: %35s -- by %d", "Increase Server Count", e.Amount)
	return str
}

// IsInterpretable always returns false
func (e *IncreaseServerCount) IsInterpretable() bool {
	return false
}

// Interpret always returns the empty string ""
func (e *IncreaseServerCount) Interpret() string {
	return ""
}

// Hash marshals the IncreaseServerCount object and computes its hash
func (e *IncreaseServerCount) Hash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("IncreaseServerCount.Hash() saw an interface that was nil")
		}
	}()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
