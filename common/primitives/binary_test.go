// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	. "github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
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
}
