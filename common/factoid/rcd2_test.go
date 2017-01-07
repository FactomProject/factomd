// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid_test

import (
	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"math/rand"
	"testing"
)

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

	if len(rcd.IsEqual(rcd2)) != 0 {
		t.Error("RCDs are not equal")
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

	if len(rcd.IsEqual(rcd2)) != 0 {
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
