// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"fmt"
	"github.com/agl/ed25519"
	"math/rand"
	"testing"
)

var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New

func Test_Auth2_Equals(test *testing.T) {

	a1 := nextAuth2()
	a2 := a1

	if a1.IsEqual(a2) != nil {
		PrtStk()
		test.Fail()
	}

	a1 = nextAuth2()

	if a1.IsEqual(a2) == nil {
		PrtStk()
		test.Fail()
	}
}
