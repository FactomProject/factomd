package adminBlock

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type EndOfMinuteEntry struct {
	MinuteNumber byte
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

func (e *EndOfMinuteEntry) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer

	buf.Write([]byte{e.Type()})
	buf.Write([]byte{e.MinuteNumber})

	return buf.DeepCopyBytes(), nil
}

func (e *EndOfMinuteEntry) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling End Of Minute: %v", r)
		}
	}()
	newData = data
	if newData[0] != e.Type() {
		return nil, fmt.Errorf("Invalid Entry type")
	}

	newData = newData[1:]
	e.MinuteNumber, newData = newData[0], newData[1:]

	return
}

func (e *EndOfMinuteEntry) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *EndOfMinuteEntry) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *EndOfMinuteEntry) JSONString() (string, error) {
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

func (e *EndOfMinuteEntry) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
