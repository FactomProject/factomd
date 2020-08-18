// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/FactomProject/goleveldb/leveldb/errors"
	"github.com/PaulSnow/factom2d/common/interfaces"
)

// Buffer contains a bytes.Buffer
type Buffer struct {
	bytes.Buffer
}

// DeepCopyBytes returns the remainder of the unread buffer. Despite its name, it DOES NOT COPY!
func (b *Buffer) DeepCopyBytes() []byte {
	// Despite the name this purposefully does not copy, copying turns out to blow up memory when unmarshalling
	// because the []bytes is very big with many messages and it all gets copied many many times
	return b.Next(b.Len())
}

// NewBuffer copies the input []byte array int a new Buffer object and returns the Buffer
func NewBuffer(buf []byte) *Buffer {
	tmp := new(Buffer)
	c := make([]byte, len(buf))
	copy(c, buf)
	tmp.Buffer = *bytes.NewBuffer(c)
	return tmp
}

// PeekByte returns the next unread byte in the buffer without advancing the read state
func (b *Buffer) PeekByte() (byte, error) {
	by, err := b.ReadByte()
	if err != nil {
		return by, err
	}
	err = b.UnreadByte()
	return by, err
}

// PushBinaryMarshallableMsgArray marshals the input message array and writes it into the Buffer
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

// PushBinaryMarshallable marshals the input and writes it to the Buffer
func (b *Buffer) PushBinaryMarshallable(bm interfaces.BinaryMarshallable) error {
	if bm == nil || reflect.ValueOf(bm).IsNil() {
		return fmt.Errorf("BinaryMarshallable is nil")
	}
	bin, err := bm.MarshalBinary()
	if err != nil {
		return err
	}
	_, err = b.Write(bin)
	return err

}

// PushMsg marshals and writes the input message to the Buffer
func (b *Buffer) PushMsg(msg interfaces.IMsg) error {
	return b.PushBinaryMarshallable(msg)
}

// PushString marshals and writes the string to the Buffer
func (b *Buffer) PushString(s string) error {
	return b.PushBytes([]byte(s))
}

// PushBytes marshals and writes the []byte array to the Buffer
func (b *Buffer) PushBytes(h []byte) error {

	l := uint64(len(h))
	err := EncodeVarInt(b, l)
	if err != nil {
		return err
	}

	_, err = b.Write(h)
	return err
}

// PushIHash marshals and writes the input hash to the Buffer
func (b *Buffer) PushIHash(h interfaces.IHash) error {
	return b.PushBytes(h.Bytes())
}

// Push appends the input []byte array to the Buffer. Return error will always
// be nil, because Write gaurantees a nil error return.
func (b *Buffer) Push(h []byte) error {
	_, err := b.Write(h)
	return err
}

// PushUInt32 writes the input uint32 to the Buffer
func (b *Buffer) PushUInt32(i uint32) error {
	return binary.Write(b, binary.BigEndian, &i)
}

// PushUInt64 writes the input uint64 to the Buffer
func (b *Buffer) PushUInt64(i uint64) error {
	return binary.Write(b, binary.BigEndian, &i)
}

// PushBool writes the input bool to the Buffer
func (b *Buffer) PushBool(boo bool) error {
	var err error
	if boo {
		_, err = b.Write([]byte{0x01})
	} else {
		_, err = b.Write([]byte{0x00})
	}
	return err
}

// PushTimestamp writes a timestamp into the Buffer
func (b *Buffer) PushTimestamp(ts interfaces.Timestamp) error {
	return b.PushInt64(ts.GetTimeMilli())
}

// PushVarInt writes the smallest possible data to the Buffer to represent the input integer
func (b *Buffer) PushVarInt(vi uint64) error {
	return EncodeVarInt(b, vi)
}

// PushByte writes the input byte to the Buffer. Returned error is always nil.
func (b *Buffer) PushByte(h byte) error {
	return b.WriteByte(h)
}

// PushInt64 writes the input int64 to the Buffer
func (b *Buffer) PushInt64(i int64) error {
	return b.PushUInt64(uint64(i))
}

// PushUInt8 writes the input int8 to the Buffer
func (b *Buffer) PushUInt8(h uint8) error {
	return b.PushByte(byte(h))
}

// PushUInt16 writes the input uint16 to the Buffer
func (b *Buffer) PushUInt16(i uint16) error {
	return binary.Write(b, binary.BigEndian, &i)
}

// PopUInt16 reads a uint16 from the Buffer
func (b *Buffer) PopUInt16() (uint16, error) {
	var i uint16
	err := binary.Read(b, binary.BigEndian, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

// PopUInt8 reads a uint8 from the Buffer
func (b *Buffer) PopUInt8() (uint8, error) {
	h, err := b.PopByte()
	if err != nil {
		return 0, err
	}
	return uint8(h), nil
}

// PushInt writes an int to the Buffer.
func (b *Buffer) PushInt(i int) error {
	return b.PushInt64(int64(i)) // Safe to cast to int64 even on 32 bit systems
}

// PopInt reads an int from the Buffer
func (b *Buffer) PopInt() (int, error) {
	i, err := b.PopInt64()
	if err != nil {
		return 0, err
	}
	return int(i), nil // Safe to cast int64 to int on 32 bit systems iff undoing PushInt
}

// PopInt64 reads an int64 from the Buffer
func (b *Buffer) PopInt64() (int64, error) {
	i, err := b.PopUInt64()
	if err != nil {
		return 0, err
	}
	return int64(i), nil
}

// PopByte reads a byte from the buffer
func (b *Buffer) PopByte() (byte, error) {
	return b.ReadByte()
}

// PopVarInt reads an integer from the Buffer
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

// PopUInt32 reads a uint32 from the Buffer
func (b *Buffer) PopUInt32() (uint32, error) {
	var i uint32
	err := binary.Read(b, binary.BigEndian, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

// PopUInt64 reads a uint64 from the Buffer
func (b *Buffer) PopUInt64() (uint64, error) {
	var i uint64
	err := binary.Read(b, binary.BigEndian, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

// PopBool reads a bool from the Buffer
func (b *Buffer) PopBool() (bool, error) {
	boo, err := b.ReadByte()
	if err != nil {
		return false, err
	}
	return boo > 0, nil
}

// PopTimestamp reads a time stamp from the Buffer
func (b *Buffer) PopTimestamp() (interfaces.Timestamp, error) {
	ts, err := b.PopInt64()
	if err != nil {
		return nil, err
	}
	return NewTimestampFromMilliseconds(uint64(ts)), nil
}

// PopString reads a string from the Buffer
func (b *Buffer) PopString() (string, error) {
	h, err := b.PopBytes()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", h), nil
}

// PopBytes reads a byte array from the Buffer
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

// PopIHash reads an hash from the Buffer
func (b *Buffer) PopIHash() (interfaces.IHash, error) {
	bb, err := b.PopBytes()
	if err != nil {
		return nil, err
	}
	return NewHash(bb), nil
}

// PopLen reads a number of bytes equal to the input length from the Buffer
func (b *Buffer) PopLen(l int) ([]byte, error) {
	answer := make([]byte, l)
	_, err := b.Read(answer)
	if err != nil {
		return nil, err
	}
	return answer, nil
}

// Pop reads a number of bytes equal to the length of the input []byte array
func (b *Buffer) Pop(h []byte) error {
	_, err := b.Read(h)
	if err != nil {
		return err
	}
	return nil
}

// PopBinaryMarshallable reads a binary marshallable interface object from the Buffer
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

// PopBinaryMarshallableMsgArray reads a message array from the Buffer
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

// PopMsg reads a message from the Buffer
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
