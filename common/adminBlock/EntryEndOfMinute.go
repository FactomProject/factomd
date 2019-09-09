package adminBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// EndOfMinuteEntry is an admin block entry which specifies a minute number. All preceding data was acknowledged before this minute.
// THIS OBJECT IS DEPRECATED AS OF MILESTONE 2
type EndOfMinuteEntry struct {
	AdminIDType  uint32 `json:"adminidtype"`  //  the type of action in this admin block entry: uint32(TYPE_MINUTE_NUM)
	MinuteNumber byte   `json:"minutenumber"` // the minute number
}

var _ interfaces.Printable = (*EndOfMinuteEntry)(nil)
var _ interfaces.BinaryMarshallable = (*EndOfMinuteEntry)(nil)
var _ interfaces.IABEntry = (*EndOfMinuteEntry)(nil)

// Type returns the hardcoded TYPE_MINUTE_NUM
func (e *EndOfMinuteEntry) Type() byte {
	return constants.TYPE_MINUTE_NUM
}

// UpdateState does nothing to the input state and returns nil
func (e *EndOfMinuteEntry) UpdateState(state interfaces.IState) error {
	return nil
}

// NewEndOfMinuteEntry creates a new EndOfMinuteEntry with the input minute value
func NewEndOfMinuteEntry(minuteNumber byte) *EndOfMinuteEntry {
	e := new(EndOfMinuteEntry)
	e.MinuteNumber = minuteNumber
	return e
}

// MarshalBinary marshals the object
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

// UnmarshalBinaryData unmarshals the input data into this object
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

// UnmarshalBinary unmarshals the input data into this object
func (e *EndOfMinuteEntry) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *EndOfMinuteEntry) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

// JSONString returns the json ecoded string
func (e *EndOfMinuteEntry) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

// String returns the EndOfMinuteEntry string
func (e *EndOfMinuteEntry) String() string {
	return fmt.Sprintf("    E: %35s -- %17s %d",
		"EndOfMinuteEntry",
		"Minute", e.MinuteNumber)
}

// IsInterpretable always returns true
func (e *EndOfMinuteEntry) IsInterpretable() bool {
	return true
}

// Interpret returns a string with the minute number
func (e *EndOfMinuteEntry) Interpret() string {
	return fmt.Sprintf("End of Minute %v", e.MinuteNumber)
}

// Hash marshals the object and computes its hash
func (e *EndOfMinuteEntry) Hash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("EndOfMinuteEntry.Hash() saw an interface that was nil")
		}
	}()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
