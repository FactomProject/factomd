// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"fmt"
	"github.com/FactomProject/ed25519"
	"math/rand"
	"testing"
)

var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New

// A hash
var hash = [ADDRESS_LENGTH]byte{
	0x61, 0xe3, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72, 0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
	0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c, 0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
}

func Test_HashEquals(test *testing.T) {
	h1 := new(Hash)
	h2 := new(Hash)

	if h1.IsEqual(h2) != nil { // Out of the box, hashes should be equal
		PrtStk()
		test.Fail()
	}

	h1.SetBytes(hash[:])

	if h1.IsEqual(h2) == nil { // Now they should not be equal
		PrtStk()
		test.Fail()
	}

	h2.SetBytes(hash[:])

	if h1.IsEqual(h2) != nil { // Back to equality!
		PrtStk()
		test.Fail()
	}
}
