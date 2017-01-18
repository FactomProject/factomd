package adminBlock

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// DB Signature Entry -------------------------
type IncreaseServerCount struct {
	Amount byte
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

func (e *IncreaseServerCount) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer

	buf.Write([]byte{e.Type()})
	buf.Write([]byte{e.Amount})

	return buf.DeepCopyBytes(), nil
}

func (e *IncreaseServerCount) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Entry Increase Server Count: %v", r)
		}
	}()

	newData = data
	newData = newData[1:]
	e.Amount, newData = newData[0], newData[1:]

	return
}

func (e *IncreaseServerCount) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *IncreaseServerCount) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *IncreaseServerCount) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *IncreaseServerCount) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
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

func (e *IncreaseServerCount) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
