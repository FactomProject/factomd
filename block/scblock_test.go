// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package block

import (
	"encoding/binary"
	"fmt"
	"github.com/agl/ed25519"
    sc "github.com/FactomProject/simplecoin"
    "math/rand"
	"testing"
    
)

var _ = sc.Prt
var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Write

func Test_create_block(test *testing.T) {
	scb := NewSCBlock(1000, 0)
	data, err := scb.MarshalBinary()
	if err != nil {
		sc.PrtStk()
		test.Fail()
	}
	scb2 := NewSCBlock(1000, 0)
	data, err = scb2.UnmarshalBinaryData(data)

	if !scb.IsEqual(scb2) {
		sc.PrtStk()
		test.Fail()
	}
}
