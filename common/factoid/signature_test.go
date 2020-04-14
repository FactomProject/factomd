// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid_test

import (
	"math/rand"
	"testing"

	"github.com/FactomProject/ed25519"
	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

var _ = ed25519.Sign
var _ = rand.New

// Global signatures used for testing
var sig1 [64]byte
var sig2 [64]byte

var s1, s2 interfaces.ISignature

// TestUnmarshalNilFactoidSignature checks that unmarshalling nil and the empty interface result in proper errors
func TestUnmarshalNilFactoidSignature(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(FactoidSignature)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

// TestSetup_Signature just sets up the global signatures (not a real test)
func TestSetup_Signature(t *testing.T) {
	sh11 := primitives.Sha([]byte("sig first half  one")).Bytes()
	sh12 := primitives.Sha([]byte("sig second half one")).Bytes()
	sh21 := primitives.Sha([]byte("sig first half  two")).Bytes()
	sh22 := primitives.Sha([]byte("sig second half two")).Bytes()

	copy(sig1[:32], sh11)
	copy(sig1[32:], sh12)
	copy(sig2[:32], sh21)
	copy(sig2[32:], sh22)

	s1 = new(FactoidSignature)
	s1.SetSignature(sig1[:])
	s2 = new(FactoidSignature)
	s2.SetSignature(sig2[:])
}

// TestNewED25519Signature checks that signatures can be set up in cononical form
func TestNewED25519Signature(t *testing.T) {
	testData := primitives.Sha([]byte("sig first half  one")).Bytes()
	priv := testHelper.NewPrivKey(0)

	sig := NewED25519Signature(priv, testData)

	pub := testHelper.PrivateKeyToEDPub(priv)
	pub2 := [32]byte{}
	copy(pub2[:], pub)

	s := sig.Signature
	valid := ed25519.VerifyCanonical(&pub2, testData, &s)
	if valid == false {
		t.Errorf("Signature is invalid - %v", valid)
	}

	priv2 := [64]byte{}
	copy(priv2[:], append(priv, pub...)[:])

	sig2 := ed25519.Sign(&priv2, testData)

	valid = ed25519.VerifyCanonical(&pub2, testData, sig2)
	if valid == false {
		t.Errorf("Test signature is invalid - %v", valid)
	}
}

/*
func TestIsEqual_Signature(t *testing.T) {
	if s1.IsEqual(s2) == nil {
		primitives.PrtStk()
		t.Fail()
	}

	s2.SetSignature(sig1[:]) // Set to sig1 for test

	if s1.IsEqual(s2) != nil {
		primitives.PrtStk()
		t.Fail()
	}

	s2.SetSignature(sig2[:]) // Reset it back to Sig2
}

func TestMarshal_Signature_(t *testing.T) {
	data, err := s1.MarshalBinary()
	s2.UnmarshalBinaryData(data)

	if s1.IsEqual(s2) != nil {
		primitives.PrtStk()
		primitives.Prtln(err)
		t.Fail()
	}

	s2.SetSignature(sig2[:]) // Reset it back to Sig2

}
*/
