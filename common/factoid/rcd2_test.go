// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid_test

import (
	"math/rand"
	"testing"

	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
)

func TestUnmarshalNilRCD_2(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(RCD_2)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestRCD2MarshalUnmarshal(t *testing.T) {
	rcd := nextAuth2_rcd2()

	hex, err := rcd.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	rcd2 := new(RCD_2)
	rest, err := rcd2.UnmarshalBinaryData(hex)
	if err != nil {
		t.Error(err)
	}
	if len(rest) > 1 {
		t.Error("Returned spare data when it shouldn't")
	}

	hex2, err := rcd2.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if len(hex) != len(hex2) {
		t.Error("Different lengths of marshalled data returned")
	}
	for i := range hex {
		if hex[i] != hex2[i] {
			t.Error("Marshalled data is not identical")
		}
	}

	if rcd.IsSameAs(rcd2) == false {
		t.Error("RCDs are not equal")
	}
}

func TestUnmarshalBadRCD2(t *testing.T) {
	rcd := nextAuth2_rcd2()

	p, err := rcd.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	// wright bad signature count to rcd2
	p[3] = 0xff

	rcd2 := new(RCD_2)
	_, err = rcd2.UnmarshalBinaryData(p)
	if err == nil {
		t.Error("RCD2 should have errored on unmarshal", rcd2)
	} else {
		t.Log(err)
	}
}

func TestRCD2Clone(t *testing.T) {
	rcd := nextAuth2_rcd2()

	hex, err := rcd.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	rcd2 := rcd.Clone().(*RCD_2)

	hex2, err := rcd2.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if len(hex) != len(hex2) {
		t.Error("Different lengths of marshalled data returned")
	}
	for i := range hex {
		if hex[i] != hex2[i] {
			t.Error("Marshalled data is not identical")
		}
	}

	if rcd.IsSameAs(rcd2) == false {
		t.Error("RCDs are not equal")
	}
}

func nextAuth2_rcd2() *RCD_2 {
	if r == nil {
		r = rand.New(rand.NewSource(1))
	}
	n := r.Int()%4 + 1
	m := r.Int()%4 + n
	addresses := make([]interfaces.IAddress, m, m)
	for j := 0; j < m; j++ {
		addresses[j] = nextAddress()
	}

	rcd, _ := NewRCD_2(n, m, addresses)
	return rcd.(*RCD_2)
}
