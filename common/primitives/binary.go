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

	llog "github.com/FactomProject/factomd/log"
)

// AreBytesEqual returns true iff the lengths and byte values of the input []byte arrays are equal
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

// AreBinaryMarshallablesEqual returns true if the input interfaces are both nil, or both exist
// and can be marshalled into byte identical arrays
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

// UnmarshalBinaryDataOfLength unmarshals an arbitrary input []byte array up to specified length, returning residual
func UnmarshalBinaryDataOfLength(dest []byte, source []byte, length int) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
			llog.LogPrintf("recovery", "Error unmarshalling: %v", r)
		}
	}()
	if (source == nil && length > 0) || len(source) < length {
		return nil, fmt.Errorf("Not enough data to unmarshal")
	}
	copy(dest[:], source[:length])
	newData = source[length:]
	return
}

// EncodeBinary returns the hexadecimal (0-F) encoding of input byte array
func EncodeBinary(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

// DecodeBinary returns a byte array of the decoded input hexadecimal (0-F) string
func DecodeBinary(bytes string) ([]byte, error) {
	return hex.DecodeString(bytes)
}

// ByteSlice32 is a fixed [32]byte array
type ByteSlice32 [32]byte

var _ interfaces.Printable = (*ByteSlice32)(nil)
var _ interfaces.BinaryMarshallable = (*ByteSlice32)(nil)

// StringToByteSlice32 converts the input hexidecimal string (0-F) to a new ByteSlice32
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

// Byte32ToByteSlice32 returns a new ByteSlice32 containing the input data
func Byte32ToByteSlice32(b [32]byte) *ByteSlice32 {
	bs := new(ByteSlice32)
	err := bs.UnmarshalBinary(b[:])
	if err != nil {
		return nil
	}
	return bs
}

// IsSameAs returns true iff input ByteSlice32 is binary identical to this ByteSlice
func (bs *ByteSlice32) IsSameAs(b *ByteSlice32) bool {
	if bs == nil || b == nil {
		if bs == nil && b == nil {
			return true
		}
		return false
	}
	return AreBytesEqual(bs[:], b[:])
}

// MarshalBinary marshals this ByteSlice32 into []byte array
func (bs *ByteSlice32) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice32.MarshalBinary err:%v", *pe)
		}
	}(&err)
	return bs[:], nil
}

// Fixed returns the internal fixed byte array data
func (bs *ByteSlice32) Fixed() [32]byte {
	return *bs
}

// UnmarshalBinaryData unmarshals the input data into this ByteSlice32
func (bs *ByteSlice32) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	newData, err = UnmarshalBinaryDataOfLength(bs[:], data, 32)
	return
}

// UnmarshalBinary unmarshals the input data into this ByteSlice32
func (bs *ByteSlice32) UnmarshalBinary(data []byte) (err error) {
	_, err = bs.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the encoded json data of this ByteSlice32
func (bs *ByteSlice32) JSONByte() ([]byte, error) {
	return EncodeJSON(bs)
}

// JSONString returns the encoded json byte string of this ByteSlice32
func (bs *ByteSlice32) JSONString() (string, error) {
	return EncodeJSONString(bs)
}

// String returns the hexidecimal string (0-F) of this ByteSlice32
func (bs *ByteSlice32) String() string {
	return fmt.Sprintf("%x", bs[:])
}

// MarshalText marshals this ByteSlice32 into the returned array
func (bs *ByteSlice32) MarshalText() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice32.MarshalText err:%v", *pe)
		}
	}(&err)
	return []byte(bs.String()), nil
}

// ByteSlice64 is a fixed [64]byte array
type ByteSlice64 [64]byte

var _ interfaces.Printable = (*ByteSlice64)(nil)
var _ interfaces.BinaryMarshallable = (*ByteSlice64)(nil)

// IsSameAs returns true iff the input ByteSlice64 is binary identical to this ByteSlice64
func (bs *ByteSlice64) IsSameAs(b *ByteSlice64) bool {
	if bs == nil || b == nil {
		if bs == nil && b == nil {
			return true
		}
		return false
	}
	return AreBytesEqual(bs[:], b[:])
}

// MarshalBinary marshals this ByteSlice64
func (bs *ByteSlice64) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice64.MarshalBinary err:%v", *pe)
		}
	}(&err)
	return bs[:], nil
}

// UnmarshalBinaryData unmarshals the input data into this ByteSlice64
func (bs *ByteSlice64) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	newData, err = UnmarshalBinaryDataOfLength(bs[:], data, 64)
	return
}

// UnmarshalBinary unmarshals the input data into this ByteSlice64
func (bs *ByteSlice64) UnmarshalBinary(data []byte) (err error) {
	_, err = bs.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded data
func (bs *ByteSlice64) JSONByte() ([]byte, error) {
	return EncodeJSON(bs)
}

// JSONString returns the json encoded byte string
func (bs *ByteSlice64) JSONString() (string, error) {
	return EncodeJSONString(bs)
}

// String returns the json encoded hexidecimal (0-F) string
func (bs *ByteSlice64) String() string {
	return fmt.Sprintf("%x", bs[:])
}

// MarshalText marshals this ByteSlice64 into text
func (bs *ByteSlice64) MarshalText() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice64.MarshalText err:%v", *pe)
		}
	}(&err)
	return []byte(bs.String()), nil
}

// ByteSlice6 is a fixed [6]byte array
type ByteSlice6 [6]byte

var _ interfaces.Printable = (*ByteSlice6)(nil)
var _ interfaces.BinaryMarshallable = (*ByteSlice6)(nil)

// IsSameAs returns true iff the input ByteSlice6 is binary identical to this ByteSlice6
func (bs *ByteSlice6) IsSameAs(b *ByteSlice6) bool {
	if bs == nil || b == nil {
		if bs == nil && b == nil {
			return true
		}
		return false
	}
	return AreBytesEqual(bs[:], b[:])
}

// MarshalBinary marshals this ByteSlice6 into a []byte array
func (bs *ByteSlice6) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice6.MarshalBinary err:%v", *pe)
		}
	}(&err)
	return bs[:], nil
}

// UnmarshalBinaryData unmarshals the input data into the ByteSlice6
func (bs *ByteSlice6) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	newData, err = UnmarshalBinaryDataOfLength(bs[:], data, 6)
	return
}

// UnmarshalBinary unmarshals the input data into the ByteSlice6
func (bs *ByteSlice6) UnmarshalBinary(data []byte) (err error) {
	_, err = bs.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded data of the ByteSlice6
func (bs *ByteSlice6) JSONByte() ([]byte, error) {
	return EncodeJSON(bs)
}

// JSONString returns the json encoded byte string of the ByteSlice6
func (bs *ByteSlice6) JSONString() (string, error) {
	return EncodeJSONString(bs)
}

// String returns the hexidecimal (0-F) string of the ByteSlice6S
func (bs *ByteSlice6) String() string {
	return fmt.Sprintf("%x", bs[:])
}

// MarshalText marshals the ByteSlice6 into text
func (bs *ByteSlice6) MarshalText() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice6.MarshalText err:%v", *pe)
		}
	}(&err)
	return []byte(bs.String()), nil
}

// ByteSliceSig is a fixed byte array of the ed25519 signature length
type ByteSliceSig [ed25519.SignatureSize]byte

var _ interfaces.Printable = (*ByteSliceSig)(nil)
var _ interfaces.BinaryMarshallable = (*ByteSliceSig)(nil)

// IsSameAs returns true iff the input ByteSliceSig is binary identical to this ByteSliceSig
func (bs *ByteSliceSig) IsSameAs(b *ByteSliceSig) bool {
	if bs == nil || b == nil {
		if bs == nil && b == nil {
			return true
		}
		return false
	}
	return AreBytesEqual(bs[:], b[:])
}

// MarshalBinary marshals this ByteSliceSig into []byte array
func (bs *ByteSliceSig) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSliceSig.MarshalBinary err:%v", *pe)
		}
	}(&err)
	return bs[:], nil
}

// GetFixed returns a new copy of the internal byte array
func (bs *ByteSliceSig) GetFixed() ([ed25519.SignatureSize]byte, error) {
	answer := [ed25519.SignatureSize]byte{}
	copy(answer[:], bs[:])

	return answer, nil
}

// UnmarshalBinaryData unmarshals the input data into the ByteSliceSig
func (bs *ByteSliceSig) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	newData, err = UnmarshalBinaryDataOfLength(bs[:], data, ed25519.SignatureSize)
	return
}

// UnmarshalBinary unmarshals the input data into the ByteSliceSig
func (bs *ByteSliceSig) UnmarshalBinary(data []byte) (err error) {
	if len(data) < ed25519.SignatureSize {
		return fmt.Errorf("Byte slice too short to unmarshal")
	}
	copy(bs[:], data[:ed25519.SignatureSize])
	return
}

// JSONByte returns the json encoded data of the ByteSliceSig
func (bs *ByteSliceSig) JSONByte() ([]byte, error) {
	return EncodeJSON(bs)
}

// JSONString returns the json encoded byte string of the ByteSliceSig
func (bs *ByteSliceSig) JSONString() (string, error) {
	return EncodeJSONString(bs)
}

// String returns a hexidecimal (0-F) string of the ByteSliceSig
func (bs *ByteSliceSig) String() string {
	return fmt.Sprintf("%x", bs[:])
}

// MarshalText marshals the ByteSliceSig into text
func (bs *ByteSliceSig) MarshalText() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSliceSig.MarshalText err:%v", *pe)
		}
	}(&err)
	return []byte(bs.String()), nil
}

// UnmarshalText unmarshals the input text to the ByteSliceSig
func (bs *ByteSliceSig) UnmarshalText(text []byte) error {
	b, err := hex.DecodeString(string(text))
	if err != nil {
		return err
	}
	return bs.UnmarshalBinary(b)
}

// ByteSlice20 is a fixed [20]byte
type ByteSlice20 [20]byte

var _ interfaces.Printable = (*ByteSlice20)(nil)
var _ interfaces.BinaryMarshallable = (*ByteSlice20)(nil)

// IsSameAs returns true iff input ByteSlice20 is binary identical to this ByteSlice20
func (bs *ByteSlice20) IsSameAs(b *ByteSlice20) bool {
	if bs == nil || b == nil {
		if bs == nil && b == nil {
			return true
		}
		return false
	}
	return AreBytesEqual(bs[:], b[:])
}

// MarshalBinary returns the byte array of the ByteSlice20
func (bs *ByteSlice20) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice20.MarshalBinary err:%v", *pe)
		}
	}(&err)
	return bs[:], nil
}

// GetFixed returns a new copy of the internal byte array
func (bs *ByteSlice20) GetFixed() ([20]byte, error) {
	answer := [20]byte{}
	copy(answer[:], bs[:])

	return answer, nil
}

// UnmarshalBinaryData unmarshals the input data into the ByteSlice20
func (bs *ByteSlice20) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	newData, err = UnmarshalBinaryDataOfLength(bs[:], data, 20)
	return
}

// UnmarshalBinary unmarshals the input data into the ByteSlice20
func (bs *ByteSlice20) UnmarshalBinary(data []byte) (err error) {
	_, err = bs.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the encoded JSON format data
func (bs *ByteSlice20) JSONByte() ([]byte, error) {
	return EncodeJSON(bs)
}

// JSONString returns the encoded JSON format byte string
func (bs *ByteSlice20) JSONString() (string, error) {
	return EncodeJSONString(bs)
}

// String returns a hexidecimal (0-F) string of the internal data
func (bs *ByteSlice20) String() string {
	return fmt.Sprintf("%x", bs[:])
}

// MarshalText marshals the internal ByteSlice20
func (bs *ByteSlice20) MarshalText() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice20.MarshalText err:%v", *pe)
		}
	}(&err)
	return []byte(bs.String()), nil
}

// ByteSlice contains a []byte
type ByteSlice struct {
	Bytes []byte
}

var _ interfaces.Printable = (*ByteSlice)(nil)
var _ interfaces.BinaryMarshallable = (*ByteSlice)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*ByteSlice)(nil)

// IsSameAs returns true iff the input ByteSlice is binary identical to this ByteSlice
func (bs *ByteSlice) IsSameAs(b *ByteSlice) bool {
	if bs == nil || b == nil {
		if bs == nil && b == nil {
			return true
		}
		return false
	}
	return AreBytesEqual(bs.Bytes, b.Bytes)
}

// RandomByteSlice returns a random non empty ByteSlice of length 1 <= len <= 63
func RandomByteSlice() *ByteSlice {
	bs := new(ByteSlice)
	x := random.RandNonEmptyByteSlice()
	bs.UnmarshalBinary(x)
	return bs
}

// StringToByteSlice converts the input string to a ByteSlice
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

// New returns a new ByteSlice
func (bs *ByteSlice) New() interfaces.BinaryMarshallableAndCopyable {
	return new(ByteSlice)
}

// MarshalBinary returns the byte array of the ByteSlice
func (bs *ByteSlice) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice.MarshalBinary err:%v", *pe)
		}
	}(&err)
	return bs.Bytes[:], nil
}

// UnmarshalBinaryData copies the input byte array to the ByteSlice
func (bs *ByteSlice) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	bs.Bytes = make([]byte, len(data))
	newData, err = UnmarshalBinaryDataOfLength(bs.Bytes[:], data, len(data))
	return
}

// UnmarshalBinary unmarshals the input byte array into the ByteSlice
func (bs *ByteSlice) UnmarshalBinary(data []byte) (err error) {
	_, err = bs.UnmarshalBinaryData(data)
	return
}

// JSONByte marshals the ByteSlice into json format
func (bs *ByteSlice) JSONByte() ([]byte, error) {
	return EncodeJSON(bs)
}

// JSONString marshals the ByteSlice into json format byte string
func (bs *ByteSlice) JSONString() (string, error) {
	return EncodeJSONString(bs)
}

// String returns a hexidecimal (0-F) string of the ByteSlice
func (bs *ByteSlice) String() string {
	return fmt.Sprintf("%x", bs.Bytes[:])
}

// MarshalText marshals the receiver into text
func (bs *ByteSlice) MarshalText() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ByteSlice.MarshalText err:%v", *pe)
		}
	}(&err)
	return []byte(bs.String()), nil
}
