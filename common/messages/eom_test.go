// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/messages"
	"testing"

	"github.com/FactomProject/factomd/common/primitives"
)

func TestMarshalUnmarshalEOM(t *testing.T) {
	eom := newEOM()
	hex, err := eom.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	eom2, err := UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str := eom2.String()
	t.Logf("str - %v", str)

	if eom2.Type() != constants.EOM_MSG {
		t.Error("Invalid message type unmarshalled")
	}
}

func TestSignAndVerifyEOM(t *testing.T) {
	eom := newEOM()
	key, err := primitives.NewPrivateKeyFromHex("07c0d52cb74f4ca3106d80c4a70488426886bccc6ebc10c6bafb37bf8a65f4c38cee85c62a9e48039d4ac294da97943c2001be1539809ea5f54721f0c5477a0a")
	if err != nil {
		t.Error(err)
	}
	err = eom.Sign(&key)
	if err != nil {
		t.Error(err)
	}
	hex, err := eom.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	t.Logf("Sig - %x", *eom.Signature.GetSignature())
	if len(*eom.Signature.GetSignature()) == 0 {
		t.Error("Signature not present")
	}

	valid, err := eom.VerifySignature()
	if err != nil {
		t.Error(err)
	}
	if valid == false {
		t.Error("Signature is not valid")
	}

	eom2, err := UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}

	if eom2.Type() != constants.EOM_MSG {
		t.Error("Invalid message type unmarshalled")
	}
	eomProper := eom2.(*EOM)

	valid, err = eomProper.VerifySignature()
	if err != nil {
		t.Error(err)
	}
	if valid == false {
		t.Error("Signature 2 is not valid")
	}

}

func newEOM() *EOM {
	eom := new(EOM)
	eom.Minute = 3
	eom.DirectoryBlockHeight = 123456
	eom.Timestamp.SetTime(0xFF22100122FF)
	hash, _ := primitives.NewShaHashFromStr("cbd3d09db6defdc25dfc7d57f3479b339a077183cd67022e6d1ef6c041522b40")
	eom.IdentityChainID = hash
	return eom
}
