// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"fmt"
	"os"

	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
)

const (
	// ServerIndexNumberSize is the size of the object below (1 uint8 member)
	ServerIndexNumberSize = 1
)

// ServerIndexNumber is an entry credit block entry. Data after this type of entry was acknowledged by the server with the stored index
type ServerIndexNumber struct {
	ServerIndexNumber uint8 `json:"serverindexnumber"` // the server index number
}

var _ interfaces.Printable = (*ServerIndexNumber)(nil)
var _ interfaces.BinaryMarshallable = (*ServerIndexNumber)(nil)
var _ interfaces.ShortInterpretable = (*ServerIndexNumber)(nil)
var _ interfaces.IECBlockEntry = (*ServerIndexNumber)(nil)

// IsSameAs checks that the input object is identical to this object
func (e *ServerIndexNumber) IsSameAs(b interfaces.IECBlockEntry) bool {
	if e == nil || b == nil {
		if e == nil && b == nil {
			return true
		}
		return false
	}
	if e.ECID() != b.ECID() {
		return false
	}

	bb, ok := b.(*ServerIndexNumber)
	if ok == false {
		return false
	}
	if e.ServerIndexNumber != bb.ServerIndexNumber {
		return false
	}

	return true
}

// String returns this object as a string
func (e *ServerIndexNumber) String() string {
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf(" %-20s\n", "ServerIndexNumber"))
	out.WriteString(fmt.Sprintf("   %-20s %d\n", "Number", e.ServerIndexNumber))
	return (string)(out.DeepCopyBytes())
}

// Hash marshals this object and computes its sha
func (e *ServerIndexNumber) Hash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "ServerIndexNumber.Hash") }()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}

// GetHash returns the hash of this object
func (e *ServerIndexNumber) GetHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "ServerIndexNumber.GetHash") }()

	return e.Hash()
}

// GetEntryHash always returns nil
func (e *ServerIndexNumber) GetEntryHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "ServerIndexNumber.GetEntryHash") }()

	return nil
}

// GetSigHash always returns nil
func (e *ServerIndexNumber) GetSigHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "ServerIndexNumber.GetSigHash") }()

	return nil
}

// IsInterpretable always returns true
func (e *ServerIndexNumber) IsInterpretable() bool {
	return true
}

// Interpret returns the ServerIndexNumber as a string
func (e *ServerIndexNumber) Interpret() string {
	return fmt.Sprintf("ServerIndexNumber %v", e.ServerIndexNumber)
}

// NewServerIndexNumber creates a new server index number object
func NewServerIndexNumber() *ServerIndexNumber {
	return new(ServerIndexNumber)
}

// NewServerIndexNumber2 creates a new server index number object, with the ServerIndexNumber set to the input
func NewServerIndexNumber2(number uint8) *ServerIndexNumber {
	sin := new(ServerIndexNumber)
	sin.ServerIndexNumber = number
	return sin
}

// ECID returns the entry credit id ECIDServerIndexNumber
func (e *ServerIndexNumber) ECID() byte {
	return constants.ECIDServerIndexNumber
}

// MarshalBinary marshals this object
func (e *ServerIndexNumber) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ServerIndexNumber.MarshalBinary err:%v", *pe)
		}
	}(&err)
	buf := primitives.NewBuffer(nil)
	err = buf.PushByte(e.ServerIndexNumber)
	if err != nil {
		return nil, err
	}
	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input data into this object
func (e *ServerIndexNumber) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	var err error
	e.ServerIndexNumber, err = buf.PopByte()
	if err != nil {
		return nil, err
	}
	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinary unmarshals the input data into this object
func (e *ServerIndexNumber) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *ServerIndexNumber) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *ServerIndexNumber) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

// GetTimestamp always returns nil
func (e *ServerIndexNumber) GetTimestamp() interfaces.Timestamp {
	return nil
}
