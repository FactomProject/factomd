// +build all 

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives_test

import (
	"encoding/hex"
	"testing"

	. "github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

func TestUnmarshalNilSignature(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(Signature)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalSignature(t *testing.T) {
	sigS := "0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a83efbcbed19b5842e5aa06e66c41d8b61826d95d50c1cbc8bd5373f986c370547133462a9ffa0dcff025a6ad26747c95f1bdd88e2596fc8c6eaa8a2993c72c050002"

	sig := new(Signature)
	h, err := hex.DecodeString(sigS)
	if err != nil {
		t.Errorf("%v", err)
	}

	rest, err := sig.UnmarshalBinaryData(h)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(rest) != 2 {
		t.Errorf("Invalid rest - %x", rest)
	}
}

func TestVerifySignature(t *testing.T) {
	msg := "Test Message Sign"
	sigS := "265ea436f627fc7817488f13bac6b10a4bf73a53f88de8b8b8de0aefb3fa5357e1770f6de3534f2671cbe69dd442c21800a61ef047e3393ca932743f75cf2506"
	pubS := "a34fec8b47929d01db00cae8d2e83acd4530f777b636a9dfb35b604a8cc4680d"

	sig := new(Signature)
	h, err := hex.DecodeString(pubS)
	if err != nil {
		t.Errorf("%v", err)
	}
	sig.SetPub(h)

	h, err = hex.DecodeString(sigS)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = sig.SetSignature(h)
	if err != nil {
		t.Errorf("%v", err)
	}

	if !sig.Verify([]byte(msg)) {
		t.Fatalf("sig.Verify retuned false")
	}
}

func TestSignatureMisc(t *testing.T) {
	priv1 := new(PrivateKey)

	err := priv1.GenerateKey()
	if err != nil {
		t.Fatalf("%v", err)
	}

	msg1 := "Test Message Sign1"
	msg2 := "Test Message Sign2"

	sig1 := priv1.Sign([]byte(msg1))

	if !sig1.Verify([]byte(msg1)) {
		t.Fatalf("sig1.Verify retuned false")
	}

	if sig1.Verify([]byte(msg2)) {
		t.Fatalf("sig1.Verify retuned true")
	}

	sigBytes := append(sig1.GetKey(), (*sig1.GetSignature())[:]...)

	sig2 := priv1.Sign([]byte(msg2))
	sig2.UnmarshalBinary(sigBytes)

	if !sig2.Verify([]byte(msg1)) {
		t.Fatalf("sig2.Verify retuned false")
	}

	if sig2.Verify([]byte(msg2)) {
		t.Fatalf("sig2.Verify retuned true")
	}

	pub := sig2.GetKey()
	pub2 := sig2.GetKey()

	if len(pub) != len(pub2) {
		t.Error("Public key length mismatch")
	}
	for i := range pub {
		if pub[i] != pub2[i] {
			t.Error("Pub keys are not identical")
		}
	}

	if sig1.IsSameAs(sig2) == false {
		t.Errorf("Signatures are not identical")
	}
}

func TestSignature(t *testing.T) {
	for i := 0; i < 1000; i++ {
		priv1 := new(PrivateKey)

		err := priv1.GenerateKey()
		if err != nil {
			t.Fatalf("%v", err)
		}

		data := random.RandByteSlice()

		sig := Sign(priv1.Key[:], data)

		pub, err := priv1.Pub.MarshalBinary()
		if err != nil {
			t.Errorf("%v", err)
		}

		err = VerifySignature(data, pub, sig)
		if err != nil {
			t.Errorf("%v", err)
		}

		sig = Sign(priv1.Key[:32], data)

		err = VerifySignature(data, pub, sig)
		if err != nil {
			t.Errorf("%v", err)
		}
	}
}
