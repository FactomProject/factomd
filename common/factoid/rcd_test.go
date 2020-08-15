// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/FactomProject/ed25519"
	. "github.com/PaulSnow/factom2d/common/factoid"
	"github.com/PaulSnow/factom2d/common/interfaces"
)

var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New

func TestUnmarshalNilBinaryAuth(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	_, _, err := UnmarshalBinaryAuth(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	_, _, err = UnmarshalBinaryAuth([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestAuth2_Equals(t *testing.T) {
	a1 := nextAuth2_rcd()
	a2 := a1

	if a1.IsSameAs(a2) == false {
		t.Errorf("Addresses are not equal")
	}

	a1 = nextAuth2_rcd()

	if a1.IsSameAs(a2) == true {
		t.Errorf("Addresses are equal")
	}
}

func nextAuth2_rcd() interfaces.IRCD {
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
	return rcd
}
