// +build all

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock_test

import (
	"encoding/hex"
	"testing"

	. "github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestUnmarshalNilECBlockHeader(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(ECBlockHeader)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestStaticECBlockHeaderUnmarshal(t *testing.T) {
	ecbh := new(ECBlockHeader)
	data, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000cbb3ff38bbb90032de6965587f46dcf37551ac26e15819303057c88999b2910b4f87cfc073df0e82cdc2ed0bb992d7ea956fd32b435b099fc35f4b0696948507a66fb49a15b68a2a0ce2382e6aa6970c835497c6074bec9794ccf84bb331ad1350000000100000000000000000b0000000000000058")
	rest, err := ecbh.UnmarshalBinaryData(data)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(rest) > 0 {
		t.Error("Returned extra data")
	}

	b, err := ecbh.MarshalBinary()
	if err != nil {
		t.Errorf("%v", err)
	}
	if primitives.AreBytesEqual(b, data) == false {
		t.Errorf("Blocks are not identical - %x vs %x", data, b)
	}
}
