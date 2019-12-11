// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

const (
	// MinuteNumberSize is the size of the object below (1 uint8 member)
	MinuteNumberSize = 1
)

// MinuteNumber is an entry credit block entry type. Data preceding this type of entry was acknowledged before the stored minute number
type MinuteNumber struct {
	Number uint8 `json:"number"` // The minute number
}

var _ interfaces.Printable = (*MinuteNumber)(nil)
var _ interfaces.BinaryMarshallable = (*MinuteNumber)(nil)
var _ interfaces.ShortInterpretable = (*MinuteNumber)(nil)
var _ interfaces.IECBlockEntry = (*MinuteNumber)(nil)

// IsSameAs checks that the input object is identical to this object
func (e *MinuteNumber) IsSameAs(b interfaces.IECBlockEntry) bool {
	if e == nil || b == nil {
		if e == nil && b == nil {
			return true
		}
		return false
	}
	if e.ECID() != b.ECID() {
		return false
	}

	bb, ok := b.(*MinuteNumber)
	if ok == false {
		return false
	}
	if e.Number != bb.Number {
		return false
	}

	return true
}

// Hash marshals this object and computes its sha
func (e *MinuteNumber) Hash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("MinuteNumber.Hash() saw an interface that was nil")
		}
	}()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}

// GetHash marshals this object and computes its hash
func (e *MinuteNumber) GetHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("MinuteNumber.GetHash() saw an interface that was nil")
		}
	}()

	return e.Hash()
}

// GetSigHash always returns nil
func (e *MinuteNumber) GetSigHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("MinuteNumber.GetSigHash() saw an interface that was nil")
		}
	}()

	return nil
}

// GetEntryHash always returns nil
func (e *MinuteNumber) GetEntryHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("MinuteNumber.GetEntryHash() saw an interface that was nil")
		}
	}()

	return nil
}

// IsInterpretable always returns true
func (e *MinuteNumber) IsInterpretable() bool {
	return true
}

// Interpret returns the minute number as a string
func (e *MinuteNumber) Interpret() string {
	return fmt.Sprintf("MinuteNumber %v", e.Number)
}

// NewMinuteNumber returns a new minute number object with the input number stored as the minute
func NewMinuteNumber(number uint8) *MinuteNumber {
	mn := new(MinuteNumber)
	mn.Number = number
	return mn
}

// ECID returns the entry credit id ECIDMinuteNumber
func (e *MinuteNumber) ECID() byte {
	return constants.ECIDMinuteNumber
}

// MarshalBinary marshals this object
func (e *MinuteNumber) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "MinuteNumber.MarshalBinary err:%v", *pe)
		}
	}(&err)
	buf := primitives.NewBuffer(nil)
	err = buf.PushByte(e.Number)
	if err != nil {
		return nil, err
	}
	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input data into this object
func (e *MinuteNumber) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	var err error
	e.Number, err = buf.PopByte()
	if err != nil {
		return nil, err
	}
	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinary unmarshals the input data into this object
func (e *MinuteNumber) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *MinuteNumber) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *MinuteNumber) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

// String returns this object as a string
func (e *MinuteNumber) String() string {
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf(" %-20s\n", "MinuteNumber"))
	out.WriteString(fmt.Sprintf("   %-20s %d\n", "Number", e.Number))
	return (string)(out.DeepCopyBytes())
}

// GetTimestamp always returns nil
func (e *MinuteNumber) GetTimestamp() interfaces.Timestamp {
	return nil
}
