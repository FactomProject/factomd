// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"testing"

	"github.com/PaulSnow/factom2d/common/constants"
	. "github.com/PaulSnow/factom2d/common/messages"
	"github.com/PaulSnow/factom2d/common/messages/msgsupport"
	"github.com/PaulSnow/factom2d/common/primitives"
)

func TestUnmarshalNilEOM(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(EOM)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalEOM(t *testing.T) {
	msg := newEOM()

	hex, err := msg.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	msg2, err := msgsupport.UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str := msg2.String()
	t.Logf("str - %v", str)

	if msg2.Type() != constants.EOM_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	hex2, err := msg2.(*EOM).MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if len(hex) != len(hex2) {
		t.Error("Hexes aren't of identical length")
	}
	for i := range hex {
		if hex[i] != hex2[i] {
			t.Error("Hexes do not match")
		}
	}

	if msg.IsSameAs(msg2.(*EOM)) != true {
		t.Errorf("EOM messages are not identical")
	}
}

func TestSignAndVerifyEOM(t *testing.T) {
	msg := newSignedEOM()
	hex, err := msg.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	t.Logf("Sig - %x", *msg.Signature.GetSignature())
	if len(*msg.Signature.GetSignature()) == 0 {
		t.Error("Signature not present")
	}

	valid, err := msg.VerifySignature()
	if err != nil {
		t.Error(err)
	}
	if valid == false {
		t.Error("Signature is not valid")
	}

	msg2, err := msgsupport.UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}

	if msg2.Type() != constants.EOM_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	valid, err = msg2.(*EOM).VerifySignature()
	if err != nil {
		t.Error(err)
	}
	if valid == false {
		t.Error("Signature 2 is not valid")
	}
}

func newEOM() *EOM {
	eom := new(EOM)
	eom.Timestamp = primitives.NewTimestampFromMilliseconds(0xFF22100122FF)
	eom.Minute = 3
	h, err := primitives.NewShaHashFromStr("deadbeef00000000000000000000000000000000000000000000000000000000")
	if err != nil {
		panic(err)
	}
	eom.ChainID = h
	eom.DBHeight = 123456

	return eom
}

func newSignedEOM() *EOM {
	ack := newEOM()

	key, err := primitives.NewPrivateKeyFromHex("07c0d52cb74f4ca3106d80c4a70488426886bccc6ebc10c6bafb37bf8a65f4c38cee85c62a9e48039d4ac294da97943c2001be1539809ea5f54721f0c5477a0a")
	if err != nil {
		panic(err)
	}
	err = ack.Sign(key)
	if err != nil {
		panic(err)
	}

	return ack
}

func TestNoResend(t *testing.T) {
	eom := newEOM()
	eom.SetNoResend(true)
	if !eom.GetNoResend() {
		t.Error("NoResend is false after being set to true")
	}
	eom.SetNoResend(false)
	if eom.GetNoResend() {
		t.Error("NoResend is true after being set to false")
	}
}

func TestSentInvalid(t *testing.T) {
	eom := newEOM()
	eom.MarkSentInvalid(true)
	if !eom.SentInvalid() {
		t.Error("SentInvalid is false after being set to true")
	}
	eom.MarkSentInvalid(false)
	if eom.SentInvalid() {
		t.Error("SentInvalid is true after being set to false")
	}
}

func TestIsStalled(t *testing.T) {
	eom := newEOM()
	eom.SetStall(true)
	if !eom.IsStalled() {
		t.Error("IsStalled is false after being set to true")
	}
	eom.SetStall(false)
	if eom.IsStalled() {
		t.Error("IsStalled is true after being set to false")
	}
}

func TestOrigin(t *testing.T) {
	eom := newEOM()
	eom.SetOrigin(123)
	if eom.GetOrigin() != 123 {
		t.Error("SetOrigin/GetOrigin mismatch")
	}
	eom.SetOrigin(321)
	if eom.GetOrigin() != 321 {
		t.Error("SetOrigin/GetOrigin mismatch")
	}
}

func TestNetworkOrigin(t *testing.T) {
	eom := newEOM()
	eom.SetNetworkOrigin("FNode00")
	if eom.GetNetworkOrigin() != "FNode00" {
		t.Error("SetNetworkOrigin/GetNetworkOrigin mismatch")
	}
	eom.SetNetworkOrigin("FNode123")
	if eom.GetNetworkOrigin() != "FNode123" {
		t.Error("SetNetworkOrigin/GetNetworkOrigin mismatch")
	}
}
