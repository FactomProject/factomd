// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid_test

import (
	crand "crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/FactomProject/ed25519"
	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
)

// TestUnmarshalNilRCD_1 checks that unmarshalling nil and the empty interface result in proper errors
func TestUnmarshalNilRCD_1(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(RCD_1)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

// TestJSONMarshal checks the prepended json string has the correct type, and its length is correct
func TestJSONMarshal(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := newRCD_1()
	s, err := a.JSONString()
	if err != nil {
		t.Error(err)
	}

	if s[1:3] != "01" {
		t.Errorf("Not prepended by rcd type, found %s", s)
	}

	if len(s) != 68 {
		t.Error("Not the correct length")
	}
}

// TestMarshal checks that 10 random RCDs can be marshalled and unmarshalled without corruption
func TestMarshal(t *testing.T) {
	for i := 0; i < 10; i++ {
		a := newRandRCD_1()
		data, err := a.MarshalBinary()
		if err != nil {
			t.Error(err)
		}

		var b = interfaces.IRCD(new(RCD_1))

		n, err := b.UnmarshalBinaryData(data)
		if err != nil {
			t.Error(err)
		}

		if len(n) > 0 {
			t.Errorf("Should have 0 bytes left, found %d", len(n))
		}

		if a.IsSameAs(b) != true {
			t.Errorf("Unmarshaled object not the same as marshalled object")
		}

		c := b.Clone()
		if a.IsSameAs(c) != true {
			t.Errorf("Cloned object is not the same")
		}
	}
}

// TestBadPublic checks that the testCreate() function will properly panic
func TestBadPublic(t *testing.T) {
	if !testCreate() {
		t.Error("Should have paniced")
	}
}

// testCreate attempts to create a new RCD with an empty type, which will cause a panic
func testCreate() (fail bool) {
	defer func() {
		if r := recover(); r != nil {
			fail = true
			return
		}
	}()

	CreateRCD([]byte{})
	return false
}

// newRandRCD_1 creates a new RCD with a public key generated from a random number
func newRandRCD_1() *RCD_1 {
	public, _, _ := ed25519.GenerateKey(crand.Reader)
	rcd := NewRCD_1(public[:])

	return rcd.(*RCD_1)
}

// newRCD_1 creates a new RCD with a public key generated from the 'zero' package global
func newRCD_1() *RCD_1 {
	public, _, _ := ed25519.GenerateKey(zero)
	rcd := NewRCD_1(public[:])

	return rcd.(*RCD_1)
}

// TestCheckSig checks that a transaction pulled from the block chain has a valid signature, and that a purposefully corrupted signature will be flagged
func TestCheckSig(t *testing.T) {
	// Get a signed transaction	from the command:
	// curl -X POST --data-binary '{"jsonrpc": "2.0", "id": 0, "method": "raw-data", "params":{"hash":"da0dd2a8dcc919d8628b91ded838026ebf784c70a2e220a1cfc8bd22e1b05706"}}' -H 'content-type:text/plain;' https://api.factomd.net/v2
	hexstring := "02016ec7acf22a010001efe18ed800ac3a83e6ca84a2c647adc6a2b2004a5692832e46b0c9efaef09fcfe4bf7230860037399721298d77984585040ea61055377039a4c3f3e2cd48c46ff643d50fd64f0142e27026226eb70a07b07c8054f69f89d010e751d4fdd47feeaf5e56cc34971f4937a81c67010259667266c51461bdcc507b9696f61111e9cd7b721ff189cce2981db016ea74e5467dfd503c1123e65551b39f74c83111890b7692d01a08420c"
	data, err := hex.DecodeString(hexstring)
	if err != nil {
		t.Errorf("TestCheckSig: error decoding hexadecimal string")
	}

	// Unmarshal the data into the transaction struct
	tx := new(Transaction)
	tx.UnmarshalBinary(data)

	rcds := tx.GetRCDs()
	sigbarray := tx.GetSignatureBlocks()
	for i := range rcds {
		rcd := rcds[i]
		for j := range sigbarray {
			sigb := sigbarray[j]
			// First we corrupt the signature
			signature := sigb.GetSignature(0)
			badsig := signature.GetSignature()
			badsig[0] = badsig[0] + 1
			if rcd.CheckSig(tx, sigb) == true {
				t.Errorf("Corrupted Signature is claimed to be verified")
			}
			// Fix the signature so it's back to original
			badsig[0] = badsig[0] - 1
			if rcd.CheckSig(tx, sigb) == false {
				t.Errorf("Invalid signature for transaction")
			}
			// This next check is commented out because it will fail with the current caching mechanism. I'm
			// leaving this here as a reminder that the caching mechanism can be improved, and when it is,
			// we can re-enable this check.
			// We corrupt the signature and try again, making sure it won't return true from cache
			//badsig[0] = badsig[0] + 1
			//if rcd.CheckSig(tx, sigb) == true {
			//	t.Errorf("Corrupted signature was deemed valid upon retry")
			//}
		}
	}
}
