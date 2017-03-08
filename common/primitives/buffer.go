// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
)

type Buffer struct {
	bytes.Buffer
}

func (b *Buffer) DeepCopyBytes() []byte {
	return b.Next(b.Len())
}

func (b *Buffer) PushBinaryMarshallable(bm interfaces.BinaryMarshallable) error {
	bin, err := bm.MarshalBinary()
	if err != nil {
		return err
	}
	_, err = b.Write(bin)
	if err != nil {
		return err
	}
	return nil
}

func (b *Buffer) PushString(s string) error {
	return b.PushBytes([]byte(s))
}

func (b *Buffer) PushBytes(h []byte) error {
	l := uint64(len(h))
	err := EncodeVarInt(b, l)
	if err != nil {
		return err
	}

	_, err = b.Write(h)
	if err != nil {
		return err
	}

	return nil
}

func (b *Buffer) PushUInt32(i uint32) error {
	return binary.Write(b, binary.BigEndian, &i)
}

func (b *Buffer) PushBool(boo bool) error {
	var err error
	if boo {
		_, err = b.Write([]byte{0x01})
	} else {
		_, err = b.Write([]byte{0x00})
	}
	return err
}

func (b *Buffer) PopUInt32() (uint32, error) {
	var i uint32
	err := binary.Read(b, binary.BigEndian, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (b *Buffer) PopBool() (bool, error) {
	boo, err := b.ReadByte()
	if err != nil {
		return false, err
	}
	return boo > 0, nil
}

func (b *Buffer) PopString() (string, error) {
	h, err := b.PopBytes()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", h), nil
}

func (b *Buffer) PopBytes() ([]byte, error) {
	h := b.DeepCopyBytes()
	l, rest := DecodeVarInt(h)

	answer := make([]byte, int(l))
	copy(answer, rest)
	remainder := rest[int(l):]

	b.Reset()
	_, err := b.Write(remainder)
	if err != nil {
		return nil, err
	}
	return answer, nil
}

func (b *Buffer) PopBinaryMarshallable(dst interfaces.BinaryMarshallable) error {
	h := b.DeepCopyBytes()
	rest, err := dst.UnmarshalBinaryData(h)
	if err != nil {
		return err
	}

	b.Reset()
	_, err = b.Write(rest)
	if err != nil {
		return err
	}
	return nil
}

func NewBuffer(buf []byte) *Buffer {
	tmp := new(Buffer)
	tmp.Buffer = *bytes.NewBuffer(buf)
	return tmp
}
