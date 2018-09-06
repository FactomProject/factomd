// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

const (
	MinuteNumberSize = 1
)

type MinuteNumber struct {
	Number uint8 `json:"number"`
}

var _ interfaces.Printable = (*MinuteNumber)(nil)
var _ interfaces.BinaryMarshallable = (*MinuteNumber)(nil)
var _ interfaces.ShortInterpretable = (*MinuteNumber)(nil)
var _ interfaces.IECBlockEntry = (*MinuteNumber)(nil)

func (a *MinuteNumber) IsSameAs(b interfaces.IECBlockEntry) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}
	if a.ECID() != b.ECID() {
		return false
	}

	bb, ok := b.(*MinuteNumber)
	if ok == false {
		return false
	}
	if a.Number != bb.Number {
		return false
	}

	return true
}

func (e *MinuteNumber) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}

func (e *MinuteNumber) GetHash() interfaces.IHash {
	return e.Hash()
}

func (e *MinuteNumber) GetSigHash() interfaces.IHash {
	return nil
}

func (a *MinuteNumber) GetEntryHash() interfaces.IHash {
	return nil
}

func (b *MinuteNumber) IsInterpretable() bool {
	return true
}

func (b *MinuteNumber) Interpret() string {
	return fmt.Sprintf("MinuteNumber %v", b.Number)
}

func NewMinuteNumber(number uint8) *MinuteNumber {
	mn := new(MinuteNumber)
	mn.Number = number
	return mn
}

func (m *MinuteNumber) ECID() byte {
	return constants.ECIDMinuteNumber
}

func (m *MinuteNumber) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "MinuteNumber.MarshalBinary err:%v", *pe)
		}
	}(&err)
	buf := primitives.NewBuffer(nil)
	err = buf.PushByte(m.Number)
	if err != nil {
		return nil, err
	}
	return buf.DeepCopyBytes(), nil
}

func (m *MinuteNumber) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	var err error
	m.Number, err = buf.PopByte()
	if err != nil {
		return nil, err
	}
	return buf.DeepCopyBytes(), nil
}

func (m *MinuteNumber) UnmarshalBinary(data []byte) (err error) {
	_, err = m.UnmarshalBinaryData(data)
	return
}

func (e *MinuteNumber) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MinuteNumber) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MinuteNumber) String() string {
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf(" %-20s\n", "MinuteNumber"))
	out.WriteString(fmt.Sprintf("   %-20s %d\n", "Number", e.Number))
	return (string)(out.DeepCopyBytes())
}

func (e *MinuteNumber) GetTimestamp() interfaces.Timestamp {
	return nil
}
