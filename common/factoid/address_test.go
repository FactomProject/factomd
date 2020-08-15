// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid_test

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"

	"github.com/FactomProject/ed25519"
	"github.com/PaulSnow/factom2d/common/constants"
	. "github.com/PaulSnow/factom2d/common/factoid"
	"github.com/PaulSnow/factom2d/common/primitives"
)

var _ = ed25519.Sign
var _ = rand.New

// An address
var address1 = [constants.ADDRESS_LENGTH]byte{
	0x61, 0xe3, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72, 0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
	0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c, 0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
}

// An address
var address2 = [constants.ADDRESS_LENGTH]byte{
	0x61, 0xe4 /* <= */, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72, 0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
	0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c, 0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
}

func TestAddressEquals(t *testing.T) {
	a1 := new(Address)
	a2 := new(Address)

	a1.SetBytes(address1[:])
	a2.SetBytes(address1[:])

	if a1.IsSameAs(a2) == false { // Out of the box, hashes should be equal
		t.Errorf("Addresses are not equal")
	}

	a1.SetBytes(address2[:])

	if a1.IsSameAs(a2) == true { // Now they should not be equal
		t.Errorf("Addresses are equal")
	}

	a2.SetBytes(address2[:])

	if a1.IsSameAs(a2) == false { // Back to equality!
		t.Errorf("Addresses are not equal")
	}
}

func TestFactoidAddresses(t *testing.T) {
	addr := NewAddress(primitives.Sha([]byte("A fake address")).Bytes())

	uaddr := primitives.ConvertFctAddressToUserStr(addr)

	if !primitives.ValidateFUserStr(uaddr) {
		t.Fail()
	}

	addrBack := primitives.ConvertUserStrToAddress(uaddr)

	if bytes.Compare(addrBack, addr.Bytes()) != 0 {
		t.Fail()
	}

	buaddr := []byte(uaddr)

	for i, v := range buaddr {
		for j := uint(0); j < 8; j++ {
			if !primitives.ValidateFUserStr(string(buaddr)) {
				t.Fail()
			}
			buaddr[i] = v ^ (01 << j)
			if primitives.ValidateFUserStr(string(buaddr)) {
				t.Fail()
			}
			buaddr[i] = v
		}
	}
}

func TestEntryCreditAddresses(t *testing.T) {
	addr := NewAddress(primitives.Sha([]byte("A fake address")).Bytes())

	uaddr := primitives.ConvertECAddressToUserStr(addr)

	if !primitives.ValidateECUserStr(uaddr) {
		t.Errorf("1")
	}

	addrBack := primitives.ConvertUserStrToAddress(uaddr)

	if bytes.Compare(addrBack, addr.Bytes()) != 0 {
		t.Errorf("2")
	}

	buaddr := []byte(uaddr)

	for i, v := range buaddr {
		for j := uint(0); j < 8; j++ {
			if !primitives.ValidateECUserStr(string(buaddr)) {
				t.Errorf("3")
				t.FailNow()
			}
			buaddr[i] = v ^ (01 << j)
			if primitives.ValidateECUserStr(string(buaddr)) {
				t.Errorf("4")
				t.FailNow()
			}
			buaddr[i] = v
		}
	}
}

func TestAddressMisc(t *testing.T) {
	h, err := primitives.HexToHash("ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973")
	if err != nil {
		t.Error(err)
	}
	add := CreateAddress(h)
	str := add.String()
	if strings.Contains(str, "ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973") == false {
		t.Errorf("String doesn't contain an expected address:\n%v", str)
	}
}
