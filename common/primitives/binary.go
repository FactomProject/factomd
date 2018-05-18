// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives/random"
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
	if b2 == nil {
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

func StringToByteSlice32(s string) *ByteSlice32 {
	bin, err := DecodeBinary(s)
	if err != nil {
		return nil
	}
	bs := new(ByteSlice32)
	err = bs.UnmarshalBinary(bin)
	if err != nil {
		return nil
	}
	return bs
}

func Byte32ToByteSlice32(b [32]byte) *ByteSlice32 {
	bs := new(ByteSlice32)
	err := bs.UnmarshalBinary(b[:])
	if err != nil {
		return nil
	}
	return bs
}

func (a *ByteSlice32) IsSameAs(b *ByteSlice32) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}
	return AreBytesEqual(a[:], b[:])
}

func (bs *ByteSlice32) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice32.MarshalBinary err:%v", *pe)
		}
	}(&err)
	return bs[:], nil
}

func (bs *ByteSlice32) Fixed() [32]byte {
	return *bs
}

func (bs *ByteSlice32) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	if data == nil || len(data) < 32 {
		return nil, fmt.Errorf("Not enough data to unmarshal")
	}
	copy(bs[:], data[:32])
	newData = data[32:]
	return
}

func (bs *ByteSlice32) UnmarshalBinary(data []byte) (err error) {
	_, err = bs.UnmarshalBinaryData(data)
	return
}

func (e *ByteSlice32) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *ByteSlice32) JSONString() (string, error) {
	return EncodeJSONString(e)
}

func (bs *ByteSlice32) String() string {
	return fmt.Sprintf("%x", bs[:])
}

func (bs *ByteSlice32) MarshalText() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice32.MarshalText err:%v", *pe)
		}
	}(&err)
	return []byte(bs.String()), nil
}

type ByteSlice64 [64]byte

var _ interfaces.Printable = (*ByteSlice64)(nil)
var _ interfaces.BinaryMarshallable = (*ByteSlice64)(nil)

func (a *ByteSlice64) IsSameAs(b *ByteSlice64) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}
	return AreBytesEqual(a[:], b[:])
}

func (bs *ByteSlice64) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice64.MarshalBinary err:%v", *pe)
		}
	}(&err)
	return bs[:], nil
}

func (bs *ByteSlice64) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	if data == nil || len(data) < 64 {
		return nil, fmt.Errorf("Not enough data to unmarshal")
	}
	copy(bs[:], data[:64])
	newData = data[64:]
	return
}

func (bs *ByteSlice64) UnmarshalBinary(data []byte) (err error) {
	_, err = bs.UnmarshalBinaryData(data)
	return
}

func (e *ByteSlice64) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *ByteSlice64) JSONString() (string, error) {
	return EncodeJSONString(e)
}

func (bs *ByteSlice64) String() string {
	return fmt.Sprintf("%x", bs[:])
}

func (bs *ByteSlice64) MarshalText() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice64.MarshalText err:%v", *pe)
		}
	}(&err)
	return []byte(bs.String()), nil
}

type ByteSlice6 [6]byte

var _ interfaces.Printable = (*ByteSlice6)(nil)
var _ interfaces.BinaryMarshallable = (*ByteSlice6)(nil)

func (a *ByteSlice6) IsSameAs(b *ByteSlice6) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}
	return AreBytesEqual(a[:], b[:])
}

func (bs *ByteSlice6) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice6.MarshalBinary err:%v", *pe)
		}
	}(&err)
	return bs[:], nil
}

func (bs *ByteSlice6) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	if data == nil || len(data) < 6 {
		return nil, fmt.Errorf("Not enough data to unmarshal")
	}
	copy(bs[:], data[:6])
	newData = data[6:]
	return
}

func (bs *ByteSlice6) UnmarshalBinary(data []byte) (err error) {
	_, err = bs.UnmarshalBinaryData(data)
	return
}

func (e *ByteSlice6) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *ByteSlice6) JSONString() (string, error) {
	return EncodeJSONString(e)
}

func (bs *ByteSlice6) String() string {
	return fmt.Sprintf("%x", bs[:])
}

func (bs *ByteSlice6) MarshalText() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice6.MarshalText err:%v", *pe)
		}
	}(&err)
	return []byte(bs.String()), nil
}

type ByteSliceSig [ed25519.SignatureSize]byte

var _ interfaces.Printable = (*ByteSliceSig)(nil)
var _ interfaces.BinaryMarshallable = (*ByteSliceSig)(nil)

func (a *ByteSliceSig) IsSameAs(b *ByteSliceSig) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}
	return AreBytesEqual(a[:], b[:])
}

func (bs *ByteSliceSig) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSliceSig.MarshalBinary err:%v", *pe)
		}
	}(&err)
	return bs[:], nil
}

func (bs *ByteSliceSig) GetFixed() ([ed25519.SignatureSize]byte, error) {
	answer := [ed25519.SignatureSize]byte{}
	copy(answer[:], bs[:])

	return answer, nil
}

func (bs *ByteSliceSig) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	if data == nil || len(data) < ed25519.SignatureSize {
		return nil, fmt.Errorf("Not enough data to unmarshal")
	}
	copy(bs[:], data[:ed25519.SignatureSize])
	newData = data[ed25519.SignatureSize:]
	return
}

func (bs *ByteSliceSig) UnmarshalBinary(data []byte) (err error) {
	if len(data) < ed25519.SignatureSize {
		return fmt.Errorf("Byte slice too short to unmarshal")
	}
	copy(bs[:], data[:ed25519.SignatureSize])
	return
}

func (e *ByteSliceSig) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *ByteSliceSig) JSONString() (string, error) {
	return EncodeJSONString(e)
}

func (bs *ByteSliceSig) String() string {
	return fmt.Sprintf("%x", bs[:])
}

func (bs *ByteSliceSig) MarshalText() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSliceSig.MarshalText err:%v", *pe)
		}
	}(&err)
	return []byte(bs.String()), nil
}

func (bs *ByteSliceSig) UnmarshalText(text []byte) error {
	b, err := hex.DecodeString(string(text))
	if err != nil {
		return err
	}
	return bs.UnmarshalBinary(b)
}

type ByteSlice20 [20]byte

var _ interfaces.Printable = (*ByteSlice20)(nil)
var _ interfaces.BinaryMarshallable = (*ByteSlice20)(nil)

func (a *ByteSlice20) IsSameAs(b *ByteSlice20) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}
	return AreBytesEqual(a[:], b[:])
}

func (bs *ByteSlice20) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice20.MarshalBinary err:%v", *pe)
		}
	}(&err)
	return bs[:], nil
}

func (bs *ByteSlice20) GetFixed() ([20]byte, error) {
	answer := [20]byte{}
	copy(answer[:], bs[:])

	return answer, nil
}

func (bs *ByteSlice20) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	if data == nil || len(data) < 20 {
		return nil, fmt.Errorf("Not enough data to unmarshal")
	}
	copy(bs[:], data[:20])
	newData = data[20:]
	return
}

func (bs *ByteSlice20) UnmarshalBinary(data []byte) (err error) {
	_, err = bs.UnmarshalBinaryData(data)
	return
}

func (e *ByteSlice20) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *ByteSlice20) JSONString() (string, error) {
	return EncodeJSONString(e)
}

func (bs *ByteSlice20) String() string {
	return fmt.Sprintf("%x", bs[:])
}

func (bs *ByteSlice20) MarshalText() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice20.MarshalText err:%v", *pe)
		}
	}(&err)
	return []byte(bs.String()), nil
}

type ByteSlice struct {
	Bytes []byte
}

var _ interfaces.Printable = (*ByteSlice)(nil)
var _ interfaces.BinaryMarshallable = (*ByteSlice)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*ByteSlice)(nil)

func (a *ByteSlice) IsSameAs(b *ByteSlice) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}
	return AreBytesEqual(a.Bytes, b.Bytes)
}

func RandomByteSlice() *ByteSlice {
	bs := new(ByteSlice)
	x := random.RandNonEmptyByteSlice()
	bs.UnmarshalBinary(x)
	return bs
}

func StringToByteSlice(s string) *ByteSlice {
	bin, err := DecodeBinary(s)
	if err != nil {
		return nil
	}
	bs := new(ByteSlice)
	err = bs.UnmarshalBinary(bin)
	if err != nil {
		return nil
	}
	return bs
}

func (bs *ByteSlice) New() interfaces.BinaryMarshallableAndCopyable {
	return new(ByteSlice)
}

func (bs *ByteSlice) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice.MarshalBinary err:%v", *pe)
		}
	}(&err)
	return bs.Bytes[:], nil
}

func (bs *ByteSlice) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	bs.Bytes = make([]byte, len(data))
	copy(bs.Bytes[:], data)
	return nil, nil
}

func (bs *ByteSlice) UnmarshalBinary(data []byte) (err error) {
	_, err = bs.UnmarshalBinaryData(data)
	return
}

func (e *ByteSlice) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *ByteSlice) JSONString() (string, error) {
	return EncodeJSONString(e)
}

func (bs *ByteSlice) String() string {
	return fmt.Sprintf("%x", bs.Bytes[:])
}

func (bs *ByteSlice) MarshalText() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice.MarshalText err:%v", *pe)
		}
	}(&err)
	return []byte(bs.String()), nil
}
