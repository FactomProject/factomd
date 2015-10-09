package AdminBlock

import (
	"bytes"
	"fmt"
	. "github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
)

type EndOfMinuteEntry struct {
	entryType byte
	EOM_Type  byte
}

var _ Printable = (*EndOfMinuteEntry)(nil)
var _ BinaryMarshallable = (*EndOfMinuteEntry)(nil)
var _ ABEntry = (*EndOfMinuteEntry)(nil)

func (m *EndOfMinuteEntry) Type() byte {
	return m.entryType
}

func (e *EndOfMinuteEntry) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	buf.Write([]byte{e.entryType})

	buf.Write([]byte{e.EOM_Type})

	return buf.Bytes(), nil
}

func (e *EndOfMinuteEntry) MarshalledSize() uint64 {
	var size uint64 = 0
	size += 1 // Type (byte)
	size += 1 // EOM_Type (byte)

	return size
}

func (e *EndOfMinuteEntry) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data

	e.entryType, newData = newData[0], newData[1:]
	e.EOM_Type, newData = newData[0], newData[1:]

	return
}

func (e *EndOfMinuteEntry) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *EndOfMinuteEntry) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *EndOfMinuteEntry) JSONString() (string, error) {
	return EncodeJSONString(e)
}

func (e *EndOfMinuteEntry) JSONBuffer(b *bytes.Buffer) error {
	return EncodeJSONToBuffer(e, b)
}

func (e *EndOfMinuteEntry) String() string {
	str, _ := e.JSONString()
	return str
}

func (e *EndOfMinuteEntry) IsInterpretable() bool {
	return true
}

func (e *EndOfMinuteEntry) Interpret() string {
	return fmt.Sprintf("End of Minute %v", e.EOM_Type)
}

func (e *EndOfMinuteEntry) Hash() IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return Sha(bin)
}
