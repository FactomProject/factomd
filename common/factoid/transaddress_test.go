// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid_test

import (
	"fmt"
	"github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"math/rand"
	"testing"
)

var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New

// An address
var transAddress1 = [constants.ADDRESS_LENGTH]byte{
	0x61, 0xe3, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72, 0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
	0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c, 0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
}

// An address
var transAddress2 = [constants.ADDRESS_LENGTH]byte{
	0x61, 0xe4 /* <= */, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72, 0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
	0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c, 0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
}

func Test_TAddressEquals(t *testing.T) {
	a1 := new(TransAddress)
	a2 := new(TransAddress)

	a1.Amount = 5
	a2.Amount = 5

	a1.Address = new(Address)
	a2.Address = new(Address)

	a1.Address.SetBytes(transAddress1[:])
	a2.Address.SetBytes(transAddress1[:])

	if a1.IsEqual(a2) != nil { // Out of the box, hashes should be equal
		t.Error("Out of the box, hashes should be equal")
	}

	a1.Address.SetBytes(transAddress2[:])

	if a1.IsEqual(a2) == nil { // Now they should not be equal
		t.Error("Now they should not be equal")
	}

	a2.Address.SetBytes(transAddress2[:])

	if a1.IsEqual(a2) != nil { // Back to equality!
		t.Error("Back to equality!")
	}

	a1.Amount = 6

	if a1.IsEqual(a2) == nil { // Amounts are not equal
		t.Error("Amounts are not equal")
	}
}

func TestTransAddressMarshalUnmarshal(t *testing.T) {
	ta := new(TransAddress)
	ta.SetAmount(12345678)
	h, err := primitives.HexToHash("ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973")
	if err != nil {
		t.Error(err)
	}
	add := h.(interfaces.IAddress)
	ta.SetAddress(add)

	hex, err := ta.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	ta2 := new(TransAddress)
	err = ta2.UnmarshalBinary(hex)
	if err != nil {
		t.Error(err)
	}
	json1, err := ta.JSONString()
	if err != nil {
		t.Error(err)
	}
	json2, err := ta2.JSONString()
	if err != nil {
		t.Error(err)
	}
	if json1 != json2 {
		t.Error("JSONs are not identical")
	}
}

func TestTransAddressMisc(t *testing.T) {
	ta := new(TransAddress)
	ta.SetAmount(12345678)
	h, err := primitives.HexToHash("ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973")
	if err != nil {
		t.Error(err)
	}
	add := h.(interfaces.IAddress)
	ta.SetAddress(add)

	text, err := ta.CustomMarshalText()
	if err != nil {
		t.Error(err)
	}
	if text != nil {
		t.Error("Text isn't nil when it should be")
	}
	str := ta.String()
	if str != "" {
		t.Error("Str isn't empty when it should be")
	}
}
