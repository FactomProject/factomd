// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

const (
	ServerIndexNumberSize = 1
)

type ServerIndexNumber struct {
	ServerIndexNumber uint8 `json:"serverindexnumber"`
}

var _ interfaces.Printable = (*ServerIndexNumber)(nil)
var _ interfaces.BinaryMarshallable = (*ServerIndexNumber)(nil)
var _ interfaces.ShortInterpretable = (*ServerIndexNumber)(nil)
var _ interfaces.IECBlockEntry = (*ServerIndexNumber)(nil)

func (a *ServerIndexNumber) IsSameAs(b interfaces.IECBlockEntry) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}
	if a.ECID() != b.ECID() {
		return false
	}

	bb, ok := b.(*ServerIndexNumber)
	if ok == false {
		return false
	}
	if a.ServerIndexNumber != bb.ServerIndexNumber {
		return false
	}

	return true
}

func (e *ServerIndexNumber) String() string {
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf(" %-20s\n", "ServerIndexNumber"))
	out.WriteString(fmt.Sprintf("   %-20s %d\n", "Number", e.ServerIndexNumber))
	return (string)(out.DeepCopyBytes())
}

func (e *ServerIndexNumber) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}

func (e *ServerIndexNumber) GetHash() interfaces.IHash {
	return e.Hash()
}

func (a *ServerIndexNumber) GetEntryHash() interfaces.IHash {
	return nil
}

func (e *ServerIndexNumber) GetSigHash() interfaces.IHash {
	return nil
}

func (b *ServerIndexNumber) IsInterpretable() bool {
	return true
}

func (b *ServerIndexNumber) Interpret() string {
	return fmt.Sprintf("ServerIndexNumber %v", b.ServerIndexNumber)
}

func NewServerIndexNumber() *ServerIndexNumber {
	return new(ServerIndexNumber)
}

func NewServerIndexNumber2(number uint8) *ServerIndexNumber {
	sin := new(ServerIndexNumber)
	sin.ServerIndexNumber = number
	return sin
}

func (s *ServerIndexNumber) ECID() byte {
	return constants.ECIDServerIndexNumber
}

func (s *ServerIndexNumber) MarshalBinary() ([]byte, error) {
	buf := primitives.NewBuffer(nil)
	err := buf.PushByte(s.ServerIndexNumber)
	if err != nil {
		return nil, err
	}
	return buf.DeepCopyBytes(), nil
}

func (s *ServerIndexNumber) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	var err error
	s.ServerIndexNumber, err = buf.PopByte()
	if err != nil {
		return nil, err
	}
	return buf.DeepCopyBytes(), nil
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

func (e *ServerIndexNumber) GetTimestamp() interfaces.Timestamp {
	return nil
}
