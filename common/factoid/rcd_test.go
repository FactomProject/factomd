// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/FactomProject/ed25519"
	. "github.com/FactomProject/factomd/common/factoid"
)

var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New

// TestUnmarshalNilBinaryAuth checks that unmarshalling nil or the empty interface results in errors
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

// TestAuth2_Equals checks that each call to nextAuth2 produces a different return
func TestAuth2_Equals(t *testing.T) {
	a1 := nextAuth2()
	a2 := a1

	if a1.IsSameAs(a2) == false {
		t.Errorf("Addresses are not equal")
	}

	a1 = nextAuth2()

	if a1.IsSameAs(a2) == true {
		t.Errorf("Addresses are equal")
	}
}
