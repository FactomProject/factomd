package adminBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// DB Signature Entry -------------------------
type IncreaseServerCount struct {
	AdminIDType uint32 `json:"adminidtype"`
	Amount      byte   `json:"amount"`
}

var _ interfaces.IABEntry = (*IncreaseServerCount)(nil)
var _ interfaces.BinaryMarshallable = (*IncreaseServerCount)(nil)

// Create a new DB Signature Entry
func NewIncreaseSererCount(num byte) (e *IncreaseServerCount) {
	e = new(IncreaseServerCount)
	e.Amount = num
	return
}

func (c *IncreaseServerCount) UpdateState(state interfaces.IState) error {
	return nil
}

func (e *IncreaseServerCount) Type() byte {
	return constants.TYPE_ADD_SERVER_COUNT
}

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

func (e *IncreaseServerCount) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *IncreaseServerCount) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

func (e *IncreaseServerCount) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

func (e *IncreaseServerCount) String() string {
	str := fmt.Sprintf("    E: %35s -- by %d", "Increase Server Count", e.Amount)
	return str
}

func (e *IncreaseServerCount) IsInterpretable() bool {
	return false
}

func (e *IncreaseServerCount) Interpret() string {
	return ""
}

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
