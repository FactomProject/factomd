// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

const (
	ServerIndexNumberSize = 1
)

type ServerIndexNumber struct {
	Number uint8
}

var _ interfaces.Printable = (*ServerIndexNumber)(nil)
var _ interfaces.BinaryMarshallable = (*ServerIndexNumber)(nil)
var _ interfaces.ShortInterpretable = (*ServerIndexNumber)(nil)
var _ interfaces.IECBlockEntry = (*ServerIndexNumber)(nil)

func (e *ServerIndexNumber) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}

func (a *ServerIndexNumber) GetEntryHash() interfaces.IHash {
	return nil
}

func (b *ServerIndexNumber) IsInterpretable() bool {
	return true
}

func (b *ServerIndexNumber) Interpret() string {
	return fmt.Sprintf("ServerIndexNumber %v", b.Number)
}

func NewServerIndexNumber() *ServerIndexNumber {
	return new(ServerIndexNumber)
}

func NewServerIndexNumber2(number uint8) *ServerIndexNumber {
	sin := new(ServerIndexNumber)
	sin.Number = number
	return sin
}

func (s *ServerIndexNumber) ECID() byte {
	return ECIDServerIndexNumber
}

func (s *ServerIndexNumber) MarshalBinary() ([]byte, error) {
	buf := new(primitives.Buffer)
	buf.WriteByte(s.Number)
	return buf.DeepCopyBytes(), nil
}

func (s *ServerIndexNumber) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	buf := primitives.NewBuffer(data)
	var c byte
	if c, err = buf.ReadByte(); err != nil {
		return
	} else {
		s.Number = c
	}
	newData = buf.DeepCopyBytes()
	return
}

func (s *ServerIndexNumber) UnmarshalBinary(data []byte) (err error) {
	_, err = s.UnmarshalBinaryData(data)
	return
}

func (e *ServerIndexNumber) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *ServerIndexNumber) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *ServerIndexNumber) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *ServerIndexNumber) String() string {
	str, _ := e.JSONString()
	return str
}
