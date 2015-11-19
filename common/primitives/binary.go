package primitives

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
)

func AreBytesEqual(b1, b2 []byte) bool {
	if len(b1) != len(b2) {
		return false
	}
	for i := range b1 {
		if b1[i] != b2[i] {
			return false
		}
	}
	return true
}

func AreBinaryMarshallablesEqual(b1, b2 interfaces.BinaryMarshallable) (bool, error) {
	if b1 == nil {
		if b2 == nil {
			return true, nil
		}
		return false, nil
	}
	bytes1, err := b1.MarshalBinary()
	if err != nil {
		return false, err
	}
	bytes2, err := b2.MarshalBinary()
	if err != nil {
		return false, err
	}
	return AreBytesEqual(bytes1, bytes2), nil
}

func EncodeBinary(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

func DecodeBinary(bytes string) ([]byte, error) {
	return hex.DecodeString(bytes)
}

type ByteSlice32 [32]byte

var _ interfaces.Printable = (*ByteSlice32)(nil)
var _ interfaces.BinaryMarshallable = (*ByteSlice32)(nil)

func (bs *ByteSlice32) MarshalBinary() ([]byte, error) {
	return bs[:], nil
}

func (bs *ByteSlice32) MarshalledSize() uint64 {
	return 32
}

func (bs *ByteSlice32) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	copy(bs[:], data[:32])
	newData = data[32:]
	return
}

func (bs *ByteSlice32) UnmarshalBinary(data []byte) (err error) {
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

func (bs *ByteSlice32) MarshalText() ([]byte, error) {
	return []byte(bs.String()), nil
}

type ByteSlice64 [64]byte

var _ interfaces.Printable = (*ByteSlice64)(nil)
var _ interfaces.BinaryMarshallable = (*ByteSlice64)(nil)

func (bs *ByteSlice64) MarshalBinary() ([]byte, error) {
	return bs[:], nil
}

func (bs *ByteSlice64) MarshalledSize() uint64 {
	return 64
}

func (bs *ByteSlice64) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	copy(bs[:], data[:64])
	newData = data[64:]
	return
}

func (bs *ByteSlice64) UnmarshalBinary(data []byte) (err error) {
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

func (bs *ByteSlice64) MarshalText() ([]byte, error) {
	return []byte(bs.String()), nil
}

type ByteSlice6 [6]byte

var _ interfaces.Printable = (*ByteSlice6)(nil)
var _ interfaces.BinaryMarshallable = (*ByteSlice6)(nil)

func (bs *ByteSlice6) MarshalBinary() ([]byte, error) {
	return bs[:], nil
}

func (bs *ByteSlice6) MarshalledSize() uint64 {
	return 6
}

func (bs *ByteSlice6) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	copy(bs[:], data[:6])
	newData = data[6:]
	return
}

func (bs *ByteSlice6) UnmarshalBinary(data []byte) (err error) {
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

func (bs *ByteSlice6) MarshalText() ([]byte, error) {
	return []byte(bs.String()), nil
}
