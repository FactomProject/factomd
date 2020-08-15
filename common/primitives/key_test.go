// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/FactomProject/ed25519"
	. "github.com/PaulSnow/factom2d/common/primitives"
)

func TestInit(t *testing.T) {
	p := new(PrivateKey)
	p.Init()
	if p.Public() == nil {
		t.Error("Should not be nil")
	}

	if p.Key == nil {
		t.Error("Should not be nil")
	}

	_, err := p.CustomMarshalText2("")
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(p.Public(), make([]byte, 32)) != 0 {
		t.Error("Should be the same")
	}
}

func TestBadHex(t *testing.T) {
	_, err := NewPrivateKeyFromHex("notgoodhex")
	if err == nil {
		t.Error("Should error on invalid hex characters")
	}

	_, err = NewPrivateKeyFromHex("")
	if err == nil {
		t.Error("Should error on no character")
	}

	_, err = NewPrivateKeyFromHex("aa")
	if err == nil {
		t.Error("Should error on no character")
	}

}

func TestUnmarshalNilPublicKey(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(PublicKey)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestGenerateKey(t *testing.T) {
	priv := new(PrivateKey)

	err := priv.GenerateKey()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if priv.Key == nil {
		t.Fatalf("bad Key")
	}
	t.Logf("PrivateKey: %v", priv.Key)

	t.Logf("PrivateKey-Hex: %v", hex.EncodeToString((*priv.Key)[:]))

	t.Logf("Pub.Key: %v", priv.Pub)
	t.Logf("Pub.Key - Hex: %v", hex.EncodeToString((*priv.Pub)[:]))
}

func TestSign(t *testing.T) {
	priv := new(PrivateKey)

	err := priv.GenerateKey()
	if err != nil {
		t.Fatalf("%v", err)
	}

	msg := "Test Message Sign"

	sig := priv.Sign([]byte(msg)).(*Signature)
	if sig.Sig == nil {
		t.Fatalf("bad Sig")
	}

	t.Logf("Sig: %v", sig.Sig)
	t.Logf("Pub.Key: %v", sig.Pub.String())

	if !sig.Verify([]byte(msg)) {
		t.Fatalf("sig.Verify retuned false")
	}
}

func TestVerify(t *testing.T) {
	priv1 := new(PrivateKey)
	priv2 := new(PrivateKey)

	err := priv1.GenerateKey()
	if err != nil {
		t.Fatalf("%v", err)
	}

	err = priv2.GenerateKey()
	if err != nil {
		t.Fatalf("%v", err)
	}

	msg1 := "Test Message Sign1"
	msg2 := "Test Message Sign2"

	sig11 := priv1.Sign([]byte(msg1)).(*Signature)
	sig12 := priv1.Sign([]byte(msg2)).(*Signature)
	sig21 := priv2.Sign([]byte(msg1)).(*Signature)
	sig22 := priv2.Sign([]byte(msg2)).(*Signature)

	if !sig11.Verify([]byte(msg1)) {
		t.Fatalf("sig11.Verify retuned false")
	}

	if sig11.Verify([]byte(msg2)) {
		t.Fatalf("sig11.Verify retuned true")
	}

	if !sig12.Verify([]byte(msg2)) {
		t.Fatalf("sig12.Verify retuned false")
	}

	if sig12.Verify([]byte(msg1)) {
		t.Fatalf("sig12.Verify retuned true")
	}

	if !sig21.Verify([]byte(msg1)) {
		t.Fatalf("sig21.Verify retuned false")
	}

	//same pub key
	sig21.Pub = sig22.Pub
	if !sig21.Verify([]byte(msg1)) {
		t.Fatalf("sig21.Verify retuned false")
	}

	//wrong pub key
	sig21.Pub = priv1.Pub
	if sig21.Verify([]byte(msg1)) {
		t.Fatalf("sig21.Verify retuned true")
	}

	if !sig22.Verify([]byte(msg2)) {
		t.Fatalf("sig22.Verify retuned false")
	}

	//wrong sig
	sig22.Sig = sig12.Sig
	if sig22.Verify([]byte(msg2)) {
		t.Fatalf("sig22.Verify retuned true")
	}

	if !priv1.Pub.Verify([]byte(msg1), (*[ed25519.SignatureSize]byte)(sig11.Sig)) {
		t.Fatalf("Pub.Verify retuned false")
	}

	if !Verify((*[32]byte)(priv1.Pub), []byte(msg1), (*[ed25519.SignatureSize]byte)(sig11.Sig)) {
		t.Fatalf("Verify retuned false")
	}

	if !VerifySlice(priv1.Pub[:], []byte(msg1), sig11.Sig[:]) {
		t.Fatalf("VerifySlice retuned false")
	}
}

func TestNewPrivateKeyFromHex(t *testing.T) {
	priv := "ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973"
	pub := "8bee2930cbe4772ae5454c4801d4ef366276f6e4cc65bac18be03607c00288c4"
	privAndPub := priv + pub
	pk1, err := NewPrivateKeyFromHex(privAndPub)
	if err != nil {
		t.Error(err)
	}
	pk2, err := NewPrivateKeyFromHex(priv)
	if err != nil {
		t.Error(err)
	}

	if AreBytesEqual(pk1.Key[:], pk2.Key[:]) == false {
		t.Error("Private keys are not equal")
	}
	if AreBytesEqual(pk1.Pub[:], pk1.Pub[:]) == false {
		t.Error("Public keys are not equal")
	}

	privKeybytes, err := hex.DecodeString(privAndPub)
	if err != nil {
		t.Error(err)
	}
	pubKeyBytes, err := hex.DecodeString(pub)
	if err != nil {
		t.Error(err)
	}

	if AreBytesEqual(pk1.Key[:], privKeybytes[:]) == false {
		t.Error("Private keys are not equal")
	}
	if AreBytesEqual(pk1.Pub[:], pubKeyBytes[:]) == false {
		t.Errorf("Public keys are not equal - %x vs %x", pk1.Pub[:], pubKeyBytes[:])
	}

	priv2 := pk1.PrivateKeyString()
	pub2 := pk1.PublicKeyString()
	if priv != priv2 {
		t.Error("Could not retrieve private key string")
	}
	if pub != pub2 {
		t.Error("Could not retrieve public key string")
	}

	priv3 := RandomPrivateKey()
	if priv3.PrivateKeyString() == priv {
		// Note: in theory this COULD fail, if you randomly retrieve the same key, but we ignore this small probability
		t.Error("Failed difference check")
	}
}

// TestPublicKeyMarshalUnmarshalText tests that a public key can be marshalled/unmarshalled from to/from text properly
func TestPublicKeyMarshalUnmarshalText(t *testing.T) {
	pub := PubKeyFromString("my test string")

	t1, err := pub.MarshalText()
	if err != nil {
		t.Errorf("%v", err)
	}
	pub2 := new(PublicKey)
	err = pub2.UnmarshalText(t1)
	if err != nil {
		t.Errorf("%v", err)
	}
	if pub.IsSameAs(pub2) == false {
		t.Error("Marshal/Unmarshal Text failed")
	}
	if pub.IsSameAs(nil) == true {
		t.Error("IsSameAs failed on nil input")
	}
	pub3, err := pub2.Copy()
	if err != nil {
		t.Errorf("%v", err)
	}
	if pub.IsSameAs(pub3) == false {
		t.Error("Public key copy failed")
	}

}
