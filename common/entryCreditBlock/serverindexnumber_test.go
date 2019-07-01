// +build all 

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/entryCreditBlock"
)

func TestUnmarshalNilServerIndexNumber(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(ServerIndexNumber)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestServerIndexMarshalUnmarshal(t *testing.T) {
	si1 := NewServerIndexNumber()
	si1.ServerIndexNumber = 3
	b, err := si1.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if len(b) != 1 {
		t.Error("Invalid byte length")
	}
	if b[0] != 3 {
		t.Error("Invalid byte")
	}

	si2 := NewServerIndexNumber()
	err = si2.UnmarshalBinary(b)
	if err != nil {
		t.Error(err)
	}
	if si1.ServerIndexNumber != si2.ServerIndexNumber {
		t.Error("Invalid data unmarshalled")
	}
}

func TestServerIndexNumberMisc(t *testing.T) {
	si := NewServerIndexNumber()
	si.ServerIndexNumber = 4
	if si.IsInterpretable() == false {
		t.Fail()
	}
	if si.Interpret() != "ServerIndexNumber 4" {
		t.Fail()
	}
	if si.Hash().String() != "e52d9c508c502347344d8c07ad91cbd6068afc75ff6292f062a09ca381c89e71" {
		t.Fail()
	}
}
