// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid_test

import (
	"encoding/hex"
	"testing"

	"github.com/FactomProject/ed25519"
	. "github.com/PaulSnow/factom2d/common/factoid"
	"github.com/PaulSnow/factom2d/testHelper"
)

func TestUnmarshalNilSignatureBlock(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(SignatureBlock)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestSignatureBlock(t *testing.T) {
	priv := testHelper.NewPrivKey(0)
	testData, err := hex.DecodeString("00112233445566778899")
	if err != nil {
		t.Error(err)
	}

	sig := NewSingleSignatureBlock(priv, testData)

	rcd := testHelper.NewFactoidRCDAddress(0)
	pub := rcd.(*RCD_1).GetPublicKey()
	pub2 := [32]byte{}
	copy(pub2[:], pub)

	s := sig.Signatures[0].(*FactoidSignature).Signature
	valid := ed25519.VerifyCanonical(&pub2, testData, &s)
	if valid == false {
		t.Error("Invalid signature")
	}
}
