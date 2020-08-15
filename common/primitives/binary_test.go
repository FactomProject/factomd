// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/FactomProject/ed25519"
	. "github.com/PaulSnow/factom2d/common/primitives"
	"github.com/PaulSnow/factom2d/common/primitives/random"
)

var testBytes []byte
var testStr string = "00010203040506070809101112131415161718192021222324252627282930313233343536373839404142434445464748495051525354555657585960616263"

// init initializes the testBytes variable above
func init() {
	h, err := hex.DecodeString(testStr)
	if err != nil {
		panic(err)
	}

	testBytes = h
}

// TestUnmarshalNilByteSlice32 tests that unmarshaling catches inappropriate empty inputs
func TestUnmarshalNilByteSlice32(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(ByteSlice32)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

// TestUnmarshalNilByteSlice64 tests that unmarshaling catches inappropriate empty inputs
func TestUnmarshalNilByteSlice64(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(ByteSlice64)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

// TestUnmarshalNilByteSlice6 tests that unmarshaling catches inappropriate empty inputs
func TestUnmarshalNilByteSlice6(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(ByteSlice6)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

// TestUnmarshalNilByteSliceSig tests that unmarshaling catches inappropriate empty inputs
func TestUnmarshalNilByteSliceSig(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(ByteSliceSig)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

// TestUnmarshalNilByteSlice20 tests that unmarshaling catches inappropriate empty inputs
func TestUnmarshalNilByteSlice20(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(ByteSlice20)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

// TestAreBytesEqual checks the AreBytesEqual fuction with N random ByteSlices in a variety of ways:
// 1) ByteSlice when copied and checked against each other, are in fact equal
// 2) Small perturbations to the copied ByteSlice are correctly returned as not equal
// 3) nil comparisons to the ByteSlice return not equal
// Finally, checks nil to nil to ensure returns equal
func TestAreBytesEqual(t *testing.T) {
	for i := 0; i < 1000; i++ {
		b1 := random.RandByteSlice()
		b2 := make([]byte, len(b1))
		copy(b2, b1)

		if AreBytesEqual(b1, b2) == false {
			t.Errorf("Equal bytes are not equal")
		}
		if len(b2) == 0 {
			continue
		}

		b2[0] = (b2[0] + 1) % 255

		if AreBytesEqual(b1, b2) == true {
			t.Errorf("Unequal bytes are equal")
		}

		b2 = b1[1:]

		if AreBytesEqual(b1, b2) == true {
			t.Errorf("Unequal bytes are equal")
		}

		if AreBytesEqual(b1, nil) == true {
			t.Errorf("Unequal bytes are equal")
		}
		if len(b2) == 0 {
			continue
		}
		if AreBytesEqual(nil, b2) == true {
			t.Errorf("Unequal bytes are equal")
		}
	}

	if AreBytesEqual(nil, nil) == false {
		t.Errorf("Equal bytes are not equal")
	}
}

// TestAreBinaryMarshallablesEqual checks N random hashes in a variety of ways:
// 1) hashes when copied and checked against each other and themselves, are in fact equal
// 2) different random hashes are correctly returned as not equal
// 3) nil comparisons to the hash return not equal
// Finally, checks nil to nil to ensure returns equal
func TestAreBinaryMarshallablesEqual(t *testing.T) {
	for i := 0; i < 1000; i++ {
		h1 := RandomHash()
		h2 := h1.Copy()

		ok, err := AreBinaryMarshallablesEqual(h1, h1)
		if err != nil {
			t.Errorf("%v", err)
		}
		if ok == false {
			t.Errorf("Equal BMs are not equal")
		}

		ok, err = AreBinaryMarshallablesEqual(h2, h2)
		if err != nil {
			t.Errorf("%v", err)
		}
		if ok == false {
			t.Errorf("Equal BMs are not equal")
		}

		ok, err = AreBinaryMarshallablesEqual(h1, h2)
		if err != nil {
			t.Errorf("%v", err)
		}
		if ok == false {
			t.Errorf("Equal BMs are not equal")
		}

		h2 = RandomHash()

		ok, err = AreBinaryMarshallablesEqual(h1, h2)
		if err != nil {
			t.Errorf("%v", err)
		}
		if ok == true {
			t.Errorf("Unequal BMs are equal")
		}
		ok, err = AreBinaryMarshallablesEqual(h1, nil)
		if err != nil {
			t.Errorf("%v", err)
		}
		if ok == true {
			t.Errorf("Unequal BMs are equal")
		}
		ok, err = AreBinaryMarshallablesEqual(nil, h2)
		if err != nil {
			t.Errorf("%v", err)
		}
		if ok == true {
			t.Errorf("Unequal BMs are equal")
		}
	}
	ok, err := AreBinaryMarshallablesEqual(nil, nil)
	if err != nil {
		t.Errorf("%v", err)
	}
	if ok == false {
		t.Errorf("Equal BMs are not equal")
	}
}

// TestEncodeBinary checks N random []byte arrays can be encoded and decoded properly
func TestEncodeBinary(t *testing.T) {
	for i := 0; i < 1000; i++ {
		h1 := random.RandByteSlice()
		s := EncodeBinary(h1)

		if s != fmt.Sprintf("%x", h1) {
			t.Errorf("Invalid string")
		}

		h2, err := DecodeBinary(s)
		if err != nil {
			t.Errorf("%v", err)
		}
		if AreBytesEqual(h1, h2) == false {
			t.Errorf("Invalid byte slice")
		}
	}
}

// TestStringToByteSlice32 checks that N random strings can be saved into the ByteSlice32 and recalled
func TestStringToByteSlice32(t *testing.T) {
	for i := 0; i < 1000; i++ {
		h := random.RandByteSliceOfLen(32)
		s := fmt.Sprintf("%x", h)
		b := StringToByteSlice32(s)
		if b.String() != s {
			t.Errorf("Invalid BS32 parsed")
		}
		b2 := b
		if b.IsSameAs(b2) == false {
			t.Errorf("ByteSlice32 IsSameAs could not detect same input")
		}
		if b.IsSameAs(nil) == true {
			t.Errorf("ByteSlice32 IsSameAs error on nil input ")
		}
	}
}

// TestByte32ToByteSlice32 checks that N random []byte arrays can be converted to ByteSlice32
func TestByte32ToByteSlice32(t *testing.T) {
	for i := 0; i < 1000; i++ {
		h := random.RandByteSliceOfLen(32)
		fixed := [32]byte{}
		copy(fixed[:], h)
		b := Byte32ToByteSlice32(fixed)
		if b.String() != fmt.Sprintf("%x", h) {
			t.Errorf("Invalid BS32 parsed")
		}
	}
}

// TestByteSliceSig checks that N random ByteSliceSigs can be marshalled and unmarshalled appropriately
func TestByteSliceSig(t *testing.T) {
	for i := 0; i < 1000; i++ {
		bss := new(ByteSliceSig)
		b1 := random.RandByteSliceOfLen(ed25519.SignatureSize)

		err := bss.UnmarshalBinary(b1)
		if err != nil {
			t.Error(err)
		}

		b2, err := bss.MarshalBinary()
		if err != nil {
			t.Error(err)
		}
		if AreBytesEqual(b1, b2) == false {
			t.Errorf("Equal bytes are not equal")
		}

		f, err := bss.GetFixed()
		if err != nil {
			t.Error(err)
		}
		if AreBytesEqual(b1, f[:]) == false {
			t.Errorf("Equal bytes are not equal")
		}

		bssame := new(ByteSliceSig)
		_ = bssame.UnmarshalBinary(b1) // Shouldn't throw error because already proven unmarshalable above
		if bssame.IsSameAs(bss) == false {
			t.Errorf("ByteSliceSig IsSameAs could not detect same input")
		}
		if bssame.IsSameAs(nil) == true {
			t.Errorf("ByteSliceSig IsSameAs error on nil input ")
		}

		extra := random.RandByteSlice()
		b3 := append(b1, extra...)

		bss2 := new(ByteSliceSig)
		extra2, err := bss2.UnmarshalBinaryData(b3)
		if err != nil {
			t.Error(err)
		}
		if AreBytesEqual(extra, extra2) == false {
			t.Errorf("Equal bytes are not equal")
		}
		if bss.String() != bss2.String() {
			t.Errorf("BSSs are not equal")
		}

		t1, err := bss.MarshalText()
		if err != nil {
			t.Error(err)
		}
		bss3 := new(ByteSliceSig)
		bss3.UnmarshalText(t1)

		if bss.String() != bss3.String() {
			t.Errorf("BSSs are not equal")
		}
	}
}

// TestByteSlice20 checks that N random ByteSlice20 can be marshalled and unmarshalled appropriately
func TestByteSlice20(t *testing.T) {
	for i := 0; i < 1000; i++ {
		bss := new(ByteSlice20)
		b1 := random.RandByteSliceOfLen(20)

		err := bss.UnmarshalBinary(b1)
		if err != nil {
			t.Error(err)
		}

		b2, err := bss.MarshalBinary()
		if err != nil {
			t.Error(err)
		}
		if AreBytesEqual(b1, b2) == false {
			t.Errorf("Equal bytes are not equal")
		}

		f, err := bss.GetFixed()
		if err != nil {
			t.Error(err)
		}
		if AreBytesEqual(b1, f[:]) == false {
			t.Errorf("Equal bytes are not equal")
		}

		bssame := new(ByteSlice20)
		_ = bssame.UnmarshalBinary(b1) // Shouldn't throw error because already proven unmarshalable above
		if bssame.IsSameAs(bss) == false {
			t.Errorf("ByteSlice20 IsSameAs could not detect same input")
		}
		if bssame.IsSameAs(nil) == true {
			t.Errorf("ByteSlice20 IsSameAs error on nil input ")
		}

		extra := random.RandByteSlice()
		b3 := append(b1, extra...)

		bss2 := new(ByteSlice20)
		extra2, err := bss2.UnmarshalBinaryData(b3)
		if err != nil {
			t.Error(err)
		}
		if AreBytesEqual(extra, extra2) == false {
			t.Errorf("Equal bytes are not equal")
		}
		if bss.String() != bss2.String() {
			t.Errorf("BSSs are not equal")
		}
	}
}

// TestByteSlice checks that N random ByteSlice can be marshalled and unmarshalled appropriately
func TestByteSlice(t *testing.T) {
	for i := 0; i < 1000; i++ {
		bss := new(ByteSlice)
		b1 := random.RandByteSlice()

		err := bss.UnmarshalBinary(b1)
		if err != nil {
			t.Error(err)
		}

		b2, err := bss.MarshalBinary()
		if err != nil {
			t.Error(err)
		}
		if AreBytesEqual(b1, b2) == false {
			t.Errorf("Equal bytes are not equal")
		}

		bssame := new(ByteSlice)
		_ = bssame.UnmarshalBinary(b1) // Shouldn't throw error because already proven unmarshalable above
		if bssame.IsSameAs(bss) == false {
			t.Errorf("ByteSlice IsSameAs could not detect same input")
		}
		if bssame.IsSameAs(nil) == true {
			t.Errorf("ByteSlice IsSameAs error on nil input ")
		}

		extra := random.RandByteSlice()
		b3 := append(b1, extra...)

		bss2 := new(ByteSlice)
		extra2, err := bss2.UnmarshalBinaryData(b3)
		if err != nil {
			t.Error(err)
		}
		if len(extra2) > 0 {
			t.Errorf("ByteSlice did not unmarshal all of the data")
		}

		bss3 := StringToByteSlice(bss2.String())
		if bss3.String() != bss2.String() {
			t.Errorf("Equal ByteSlices are not equal")
		}
	}
}

// TestBA64Misc checks that ByteSlice64 can be marshalled and unmarshalled appropriately
func TestBA64Misc(t *testing.T) {
	ba := new(ByteSlice64)

	err := ba.UnmarshalBinary(testBytes)
	if err != nil {
		t.Error(err)
	}

	h, err := ba.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if AreBytesEqual(h, testBytes) == false {
		t.Errorf("Failed MarshalBinary. Expected %x, got %x", testBytes, h)
	}

	if ba.String() != testStr {
		t.Error("Failed String")
	}

	rest, err := new(ByteSlice64).UnmarshalBinaryData(append(testBytes, 0xFF))
	if err != nil {
		t.Error(err)
	}
	if AreBytesEqual(rest, []byte{0xFF}) == false {
		t.Errorf("Failed UnmarshalBinaryData - %x", rest)
	}

	json, err := ba.JSONString()
	if err != nil {
		t.Error(err)
	}
	if json != "\""+testStr+"\"" {
		t.Errorf("Failed JSONString - %s", json)
	}
	ba2 := ba
	if ba.IsSameAs(ba2) == false {
		t.Errorf("ByteSlice64 IsSameAs could not detect same input")
	}
	if ba.IsSameAs(nil) == true {
		t.Errorf("ByteSlice64 IsSameAs error on nil input ")
	}
}

// TestBA32Misc checks that ByteSlice32 can be marshalled and unmarshalled appropriately
func TestBA32Misc(t *testing.T) {
	ba := new(ByteSlice32)
	testStr32 := testStr[:64]
	testBytes32 := testBytes[:32]

	err := ba.UnmarshalBinary(testBytes32)
	if err != nil {
		t.Error(err)
	}

	h, err := ba.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if AreBytesEqual(h, testBytes32) == false {
		t.Errorf("Failed MarshalBinary. Expected %x, got %x", testBytes32, h)
	}

	if ba.String() != testStr32 {
		t.Error("Failed String")
	}

	rest, err := new(ByteSlice32).UnmarshalBinaryData(append(testBytes32, 0xFF))
	if err != nil {
		t.Error(err)
	}
	if AreBytesEqual(rest, []byte{0xFF}) == false {
		t.Errorf("Failed UnmarshalBinaryData - %x", rest)
	}

	json, err := ba.JSONString()
	if err != nil {
		t.Error(err)
	}
	if json != "\""+testStr32+"\"" {
		t.Errorf("Failed JSONString - %s", json)
	}
	ba2 := ba
	if ba.IsSameAs(ba2) == false {
		t.Errorf("ByteSlice32 IsSameAs could not detect same input")
	}
	if ba.IsSameAs(nil) == true {
		t.Errorf("ByteSlice32 IsSameAs error on nil input ")
	}
}

// TestBA6Misc checks that ByteSlice6 can be marshalled and unmarshalled appropriately
func TestBA6Misc(t *testing.T) {
	ba := new(ByteSlice6)
	testStr6 := testStr[:12]
	testBytes6 := testBytes[:6]

	err := ba.UnmarshalBinary(testBytes6)
	if err != nil {
		t.Error(err)
	}

	h, err := ba.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if AreBytesEqual(h, testBytes6) == false {
		t.Errorf("Failed MarshalBinary. Expected %x, got %x", testBytes6, h)
	}

	if ba.String() != testStr6 {
		t.Error("Failed String")
	}

	rest, err := new(ByteSlice6).UnmarshalBinaryData(append(testBytes6, 0xFF))
	if err != nil {
		t.Error(err)
	}
	if AreBytesEqual(rest, []byte{0xFF}) == false {
		t.Errorf("Failed UnmarshalBinaryData - %x", rest)
	}

	json, err := ba.JSONString()
	if err != nil {
		t.Error(err)
	}
	if json != "\""+testStr6+"\"" {
		t.Errorf("Failed JSONString - %s", json)
	}
	ba2 := ba
	if ba.IsSameAs(ba2) == false {
		t.Errorf("ByteSlice6 IsSameAs could not detect same input")
	}
	if ba.IsSameAs(nil) == true {
		t.Errorf("ByteSlice6 IsSameAs error on nil input ")
	}
}
