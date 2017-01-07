// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid_test

import (
	"fmt"
	"github.com/FactomProject/ed25519"
	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
	"math/rand"
	"testing"
)

var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New

var sig1 [64]byte
var sig2 [64]byte

var s1, s2 interfaces.ISignature

func Test_Setup_Signature(test *testing.T) {
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

	//    txt1,_:=s1.CustomMarshalText()
	//    txt2,_:=s2.CustomMarshalText()
	//    Prtln(string(txt1))
	//    Prtln(string(txt2))
}

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
func Test_IsEqual_Signature(test *testing.T) {

	if s1.IsEqual(s2) == nil {
		primitives.PrtStk()
		test.Fail()
	}

	s2.SetSignature(sig1[:]) // Set to sig1 for test

	if s1.IsEqual(s2) != nil {
		primitives.PrtStk()
		test.Fail()
	}

	s2.SetSignature(sig2[:]) // Reset it back to Sig2
}

func Test_Marshal_Signature_(test *testing.T) {
	data, err := s1.MarshalBinary()
	s2.UnmarshalBinaryData(data)

	if s1.IsEqual(s2) != nil {
		primitives.PrtStk()
		primitives.Prtln(err)
		test.Fail()
	}

	s2.SetSignature(sig2[:]) // Reset it back to Sig2

}
*/
