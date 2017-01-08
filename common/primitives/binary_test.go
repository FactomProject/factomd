// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives_test

import (
	"encoding/hex"
	. "github.com/FactomProject/factomd/common/primitives"
	"testing"
)

var testBytes []byte
var testStr string = "00010203040506070809101112131415161718192021222324252627282930313233343536373839404142434445464748495051525354555657585960616263"

func init() {
	h, err := hex.DecodeString(testStr)
	if err != nil {
		panic(err)
	}

	testBytes = h
}

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
	if areBytesIdentical(h, testBytes) == false {
		t.Errorf("Failed MarshalBinary. Expected %x, got %x", testBytes, h)
	}

	if ba.String() != testStr {
		t.Error("Failed String")
	}

	rest, err := new(ByteSlice64).UnmarshalBinaryData(append(testBytes, 0xFF))
	if err != nil {
		t.Error(err)
	}
	if areBytesIdentical(rest, []byte{0xFF}) == false {
		t.Errorf("Failed UnmarshalBinaryData - %x", rest)
	}

	json, err := ba.JSONString()
	if err != nil {
		t.Error(err)
	}
	if json != "\""+testStr+"\"" {
		t.Errorf("Failed JSONString - %s", json)
	}
}

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
	if areBytesIdentical(h, testBytes32) == false {
		t.Errorf("Failed MarshalBinary. Expected %x, got %x", testBytes32, h)
	}

	if ba.String() != testStr32 {
		t.Error("Failed String")
	}

	rest, err := new(ByteSlice32).UnmarshalBinaryData(append(testBytes32, 0xFF))
	if err != nil {
		t.Error(err)
	}
	if areBytesIdentical(rest, []byte{0xFF}) == false {
		t.Errorf("Failed UnmarshalBinaryData - %x", rest)
	}

	json, err := ba.JSONString()
	if err != nil {
		t.Error(err)
	}
	if json != "\""+testStr32+"\"" {
		t.Errorf("Failed JSONString - %s", json)
	}
}

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
	if areBytesIdentical(h, testBytes6) == false {
		t.Errorf("Failed MarshalBinary. Expected %x, got %x", testBytes6, h)
	}

	if ba.String() != testStr6 {
		t.Error("Failed String")
	}

	rest, err := new(ByteSlice6).UnmarshalBinaryData(append(testBytes6, 0xFF))
	if err != nil {
		t.Error(err)
	}
	if areBytesIdentical(rest, []byte{0xFF}) == false {
		t.Errorf("Failed UnmarshalBinaryData - %x", rest)
	}

	json, err := ba.JSONString()
	if err != nil {
		t.Error(err)
	}
	if json != "\""+testStr6+"\"" {
		t.Errorf("Failed JSONString - %s", json)
	}
}

func areBytesIdentical(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
