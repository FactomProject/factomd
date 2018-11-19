// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/goleveldb/leveldb/errors"
)

type Buffer struct {
	bytes.Buffer
}

func (b *Buffer) DeepCopyBytes() []byte {
	// Despite the name this purposefully does not copy, copying turns out to blow up memory when unmarshalling
	// because the []bytes is very big with many messages and it all gets copied many many times
	return b.Next(b.Len())
}

func NewBuffer(buf []byte) *Buffer {
	tmp := new(Buffer)
	c := make([]byte, len(buf))
	copy(c, buf)
	tmp.Buffer = *bytes.NewBuffer(c)
	return tmp
}

func (b *Buffer) PeekByte() (byte, error) {
	by, err := b.ReadByte()
	if err != nil {
		return by, err
	}
	err = b.UnreadByte()
	if err != nil {
		return by, err
	}
	return by, nil
}

func (b *Buffer) PushBinaryMarshallableMsgArray(bm []interfaces.IMsg) error {
	err := b.PushInt(len(bm))
	if err != nil {
		return err
	}

	for _, v := range bm {
		err = b.PushMsg(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Buffer) PushBinaryMarshallable(bm interfaces.BinaryMarshallable) error {
	if bm == nil {
		return fmt.Errorf("BinaryMarshallable is nil")
	}
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

func (b *Buffer) PushMsg(msg interfaces.IMsg) error {
	return b.PushBinaryMarshallable(msg)
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

func (b *Buffer) PushIHash(h interfaces.IHash) error {
	return b.PushBytes(h.Bytes())
}

func (b *Buffer) Push(h []byte) error {
	_, err := b.Write(h)
	if err != nil {
		return err
	}
	return nil
}

func (b *Buffer) PushUInt32(i uint32) error {
	return binary.Write(b, binary.BigEndian, &i)
}

func (b *Buffer) PushUInt64(i uint64) error {
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

func (b *Buffer) PushTimestamp(ts interfaces.Timestamp) error {
	return b.PushInt64(ts.GetTimeMilli())
}

func (b *Buffer) PushVarInt(vi uint64) error {
	return EncodeVarInt(b, vi)
}

func (b *Buffer) PushByte(h byte) error {
	return b.WriteByte(h)
}

func (b *Buffer) PushInt64(i int64) error {
	return b.PushUInt64(uint64(i))
}

func (b *Buffer) PushUInt8(h uint8) error {
	return b.PushByte(byte(h))
}

func (b *Buffer) PushUInt16(i uint16) error {
	return binary.Write(b, binary.BigEndian, &i)
}

func (b *Buffer) PopUInt16() (uint16, error) {
	var i uint16
	err := binary.Read(b, binary.BigEndian, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (b *Buffer) PopUInt8() (uint8, error) {
	h, err := b.PopByte()
	if err != nil {
		return 0, err
	}
	return uint8(h), nil
}

func (b *Buffer) PushInt(i int) error {
	return b.PushInt64(int64(i))
}

func (b *Buffer) PopInt() (int, error) {
	i, err := b.PopInt64()
	if err != nil {
		return 0, err
	}
	return int(i), nil
}

func (b *Buffer) PopInt64() (int64, error) {
	i, err := b.PopUInt64()
	if err != nil {
		return 0, err
	}
	return int64(i), nil
}

func (b *Buffer) PopByte() (byte, error) {
	return b.ReadByte()
}

func (b *Buffer) PopVarInt() (uint64, error) {
	h := b.DeepCopyBytes()
	l, rest := DecodeVarInt(h)
	b.Reset()
	_, err := b.Write(rest)
	if err != nil {
		return 0, err
	}
	return l, nil
}

func (b *Buffer) PopUInt32() (uint32, error) {
	var i uint32
	err := binary.Read(b, binary.BigEndian, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (b *Buffer) PopUInt64() (uint64, error) {
	var i uint64
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

func (b *Buffer) PopTimestamp() (interfaces.Timestamp, error) {
	ts, err := b.PopInt64()
	if err != nil {
		return nil, err
	}
	return NewTimestampFromMilliseconds(uint64(ts)), nil
}

func (b *Buffer) PopString() (string, error) {
	h, err := b.PopBytes()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", h), nil
}

func (b *Buffer) PopBytes() ([]byte, error) {
	l, err := b.PopVarInt()
	if err != nil || int(l) < 0 {
		return nil, err
	}

	if b.Len() < int(l) {
		return nil, errors.New(fmt.Sprintf("End of Buffer Looking for %d but only have %d", l, b.Len()))
	}
	answer := make([]byte, int(l))
	al, err := b.Read(answer)
	if al != int(l) {
		return nil, errors.New("2End of Buffer")
	}
	return answer, nil
}

func (b *Buffer) PopIHash() (interfaces.IHash, error) {
	bb, err := b.PopBytes()
	if err != nil {
		return nil, err
	}
	return NewHash(bb), nil
}

func (b *Buffer) PopLen(l int) ([]byte, error) {
	answer := make([]byte, l)
	_, err := b.Read(answer)
	if err != nil {
		return nil, err
	}
	return answer, nil
}

func (b *Buffer) Pop(h []byte) error {
	_, err := b.Read(h)
	if err != nil {
		return err
	}
	return nil
}

func (b *Buffer) PopBinaryMarshallable(dst interfaces.BinaryMarshallable) error {
	if dst == nil {
		return fmt.Errorf("Destination is nil")
	}
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

func (b *Buffer) PopBinaryMarshallableMsgArray() ([]interfaces.IMsg, error) {
	l, err := b.PopInt()
	if err != nil {
		return nil, err
	}

	var msgs []interfaces.IMsg
	for i := 0; i < l; i++ {
		var msg interfaces.IMsg
		msg, err = b.PopMsg()
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}

	return msgs, nil
}

var General interfaces.IGeneralMsg

func (b *Buffer) PopMsg() (msg interfaces.IMsg, err error) {
	h := b.DeepCopyBytes()
	rest, msg, err := General.UnmarshalMessageData(h)
	if err != nil {
		return nil, err
	}
	used := len(h) - len(rest)
	_, err = b.Write(h[used:])
	if err != nil {
		return nil, err
	}
	return msg, err
}
