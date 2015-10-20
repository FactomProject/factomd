package primitives_test

import (
	"encoding/hex"
	. "github.com/FactomProject/factomd/common/primitives"
	"testing"
)

var testBytes []byte

func init() {
	testStr := "00010203040506070809101112131415161718192021222324252627282930313233343536373839404142434445464748495051525354555657585960616263"

	h, err := hex.DecodeString(testStr)
	if err != nil {
		panic(err)
	}

	testBytes = h
}

func TestBAMisc(t *testing.T) {
	ba := new(ByteArray)

	prefix, err := hex.DecodeString("0000000000000040")
	if err != nil {
		t.Error(err)
	}

	err = ba.UnmarshalBinary(append(prefix, testBytes...))
	if err != nil {
		t.Error(err)
	}

	h, err := ba.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if areBytesIdentical(h, append(prefix, testBytes...)) == false {
		t.Errorf("Failed MarshalBinary. Expected %x, got %x", append(prefix, testBytes...), h)
	}

	ba2, err := NewByteArray(testBytes)
	if err != nil {
		t.Error(err)
	}
	h2, err := ba2.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if areBytesIdentical(h2, h) == false {
		t.Errorf("Failed NewByteArray")
	}

	if areBytesIdentical(ba2.Bytes(), testBytes) == false {
		t.Errorf("Failed NewByteArray")
	}
	if ba2.MarshalledSize() != 72 {
		t.Error("Failed MarshalledSize")
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
