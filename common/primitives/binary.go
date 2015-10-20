package primitives

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
)

type ByteArray []byte

func (ba ByteArray) Bytes() []byte {
	newArray := make([]byte, len(ba))
	copy(newArray, ba[:])
	return newArray
}

func (ba ByteArray) SetBytes(newArray []byte) error {
	copy(ba[:], newArray[:])
	return nil
}

func (ba ByteArray) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	//fmt.Println("uint64(len(ba) ",uint64(len(ba)))

	binary.Write(&buf, binary.BigEndian, uint64(len(ba)))
	buf.Write(ba)
	return buf.Bytes(), nil
}

func (ba ByteArray) MarshalledSize() uint64 {
	//fmt.Println("uint64(len(ba) + 8)",uint64(len(ba) + 8))
	return uint64(len(ba) + 8)
}

func (ba ByteArray) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data
	count := binary.BigEndian.Uint64(newData[0:8])

	newData = newData[8:]

	tmp := make([]byte, count)

	copy(tmp[:], newData[:count])
	newData = newData[count:]

	return
}

func (ba ByteArray) UnmarshalBinary(data []byte) (err error) {
	_, err = ba.UnmarshalBinaryData(data)
	return
}

func NewByteArray(newHash []byte) (*ByteArray, error) {
	var sh ByteArray
	err := sh.SetBytes(newHash)
	if err != nil {
		return nil, err
	}
	return &sh, err
}

type ByteSlice32 [32]byte

var _ interfaces.Printable = (*ByteSlice32)(nil)
var _ interfaces.BinaryMarshallable = (*ByteSlice32)(nil)

func (bs ByteSlice32) MarshalBinary() ([]byte, error) {
	return bs[:], nil
}

func (bs ByteSlice32) MarshalledSize() uint64 {
	return 32
}

func (bs ByteSlice32) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	copy(bs[:], data[:32])
	newData = data[:32]
	return
}

func (bs ByteSlice32) UnmarshalBinary(data []byte) (err error) {
	copy(bs[:], data[:32])
	return
}

func (e *ByteSlice32) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *ByteSlice32) JSONString() (string, error) {
	return EncodeJSONString(e)
}

func (e *ByteSlice32) JSONBuffer(b *bytes.Buffer) error {
	return EncodeJSONToBuffer(e, b)
}

func (bs *ByteSlice32) String() string {
	return fmt.Sprintf("%x", bs[:])
}

func (bs ByteSlice32) MarshalText() ([]byte, error) {
	return []byte(bs.String()), nil
}

type ByteSlice64 [64]byte

var _ interfaces.Printable = (*ByteSlice64)(nil)
var _ interfaces.BinaryMarshallable = (*ByteSlice64)(nil)

func (bs ByteSlice64) MarshalBinary() ([]byte, error) {
	return bs[:], nil
}

func (bs ByteSlice64) MarshalledSize() uint64 {
	return 64
}

func (bs ByteSlice64) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	copy(bs[:], data[:64])
	newData = data[:64]
	return
}

func (bs ByteSlice64) UnmarshalBinary(data []byte) (err error) {
	copy(bs[:], data[:64])
	return
}

func (e *ByteSlice64) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *ByteSlice64) JSONString() (string, error) {
	return EncodeJSONString(e)
}

func (e *ByteSlice64) JSONBuffer(b *bytes.Buffer) error {
	return EncodeJSONToBuffer(e, b)
}

func (bs *ByteSlice64) String() string {
	return fmt.Sprintf("%x", bs[:])
}

func (bs ByteSlice64) MarshalText() ([]byte, error) {
	return []byte(bs.String()), nil
}

type ByteSlice6 [6]byte

var _ interfaces.Printable = (*ByteSlice6)(nil)
var _ interfaces.BinaryMarshallable = (*ByteSlice6)(nil)

func (bs ByteSlice6) MarshalBinary() ([]byte, error) {
	return bs[:], nil
}

func (bs ByteSlice6) MarshalledSize() uint64 {
	return 6
}

func (bs ByteSlice6) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	copy(bs[:], data[:6])
	newData = data[:6]
	return
}

func (bs ByteSlice6) UnmarshalBinary(data []byte) (err error) {
	copy(bs[:], data[:6])
	return
}

func (e *ByteSlice6) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *ByteSlice6) JSONString() (string, error) {
	return EncodeJSONString(e)
}

func (e *ByteSlice6) JSONBuffer(b *bytes.Buffer) error {
	return EncodeJSONToBuffer(e, b)
}

func (bs *ByteSlice6) String() string {
	return fmt.Sprintf("%x", bs[:])
}

func (bs ByteSlice6) MarshalText() ([]byte, error) {
	return []byte(bs.String()), nil
}
