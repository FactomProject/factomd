// +build all 

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package adminBlock_test

import (
	"fmt"
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
)

var _ = fmt.Print

func TestUnmarshalNilABlockHeader(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	{
		a := new(ABlockHeader)

		data, err := a.MarshalBinary()
		if err != nil || data == nil {
			t.Error("Should be able to marshal an Admin block header")
		}
		b := new(ABlockHeader)
		b.UnmarshalBinary(data)
		if !a.IsSameAs(b) {
			t.Error("Failed to marshal/unmarshal header")
		}
	}
	a := new(ABlockHeader)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Error("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Error("Error is nil when it shouldn't be")
	}
}

func TestUnmarshalNilABlockHeaderWithAdminBlk(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := createSmallTestAdminBlock()
	data, err := a.MarshalBinary()
	if err != nil || data == nil {
		t.Error("Should be able to marshal an Admin block header")
	}
	b := createSmallTestAdminBlock()
	b.UnmarshalBinary(data)
	if !a.IsSameAs(b) {
		t.Error("Failed to marshal/unmarshal header")
	}

}
