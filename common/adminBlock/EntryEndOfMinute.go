package adminBlock

import (
	"fmt"
	"os"

	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
)

type EndOfMinuteEntry struct {
	AdminIDType  uint32 `json:"adminidtype"`
	MinuteNumber byte   `json:"minutenumber"`
}

var _ interfaces.Printable = (*EndOfMinuteEntry)(nil)
var _ interfaces.BinaryMarshallable = (*EndOfMinuteEntry)(nil)
var _ interfaces.IABEntry = (*EndOfMinuteEntry)(nil)

func (m *EndOfMinuteEntry) Type() byte {
	return constants.TYPE_MINUTE_NUM
}

func (c *EndOfMinuteEntry) UpdateState(state interfaces.IState) error {
	return nil
}

func NewEndOfMinuteEntry(minuteNumber byte) *EndOfMinuteEntry {
	e := new(EndOfMinuteEntry)
	e.MinuteNumber = minuteNumber
	return e
}

func (e *EndOfMinuteEntry) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "EndOfMinuteEntry.MarshalBinary err:%v", *pe)
		}
	}(&err)
	var buf primitives.Buffer

	e.AdminIDType = uint32(e.Type())

	err = buf.PushByte(e.Type())
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(e.MinuteNumber)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *EndOfMinuteEntry) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	b, err := buf.PopByte()
	if err != nil {
		return nil, err
	}
	if b != e.Type() {
		return nil, fmt.Errorf("Invalid Entry type")
	}

	e.MinuteNumber, err = buf.PopByte()
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *EndOfMinuteEntry) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *EndOfMinuteEntry) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

func (e *EndOfMinuteEntry) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

func (e *EndOfMinuteEntry) String() string {
	return fmt.Sprintf("    E: %35s -- %17s %d",
		"EndOfMinuteEntry",
		"Minute", e.MinuteNumber)
}

func (e *EndOfMinuteEntry) IsInterpretable() bool {
	return true
}

func (e *EndOfMinuteEntry) Interpret() string {
	return fmt.Sprintf("End of Minute %v", e.MinuteNumber)
}

func (e *EndOfMinuteEntry) Hash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "EndOfMinuteEntry.Hash") }()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
