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
