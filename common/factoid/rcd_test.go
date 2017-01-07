// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid_test

import (
	"fmt"
	"github.com/FactomProject/ed25519"
	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"math/rand"
	"testing"
)

var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New

func Test_Auth2_Equals(test *testing.T) {

	a1 := nextAuth2_rcd()
	a2 := a1

	if a1.IsEqual(a2) != nil {
		primitives.PrtStk()
		test.Fail()
	}

	a1 = nextAuth2_rcd()

	if a1.IsEqual(a2) == nil {
		primitives.PrtStk()
		test.Fail()
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
