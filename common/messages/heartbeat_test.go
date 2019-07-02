// +build all

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/identity"
	. "github.com/FactomProject/factomd/common/messages"
	. "github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

func TestUnmarshalNilHeartbeat(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(Heartbeat)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalHeartbeat(t *testing.T) {
	msg := newHeartbeat()

	hex, err := msg.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	msg2, err := UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str := msg2.String()
	t.Logf("str - %v", str)

	if msg2.Type() != constants.HEARTBEAT_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	hex2, err := msg2.(*Heartbeat).MarshalBinary()
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

	if msg.IsSameAs(msg2.(*Heartbeat)) != true {
		t.Errorf("Heartbeat messages are not identical")
	}
}

func TestSignAndVerifyHeartbeat(t *testing.T) {
	msg := newSignedHeartbeat()
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

	msg2, err := UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}

	if msg2.Type() != constants.HEARTBEAT_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	valid, err = msg2.(*Heartbeat).VerifySignature()
	if err != nil {
		t.Error(err)
	}
	if valid == false {
		t.Error("Signature 2 is not valid")
	}
}

func newHeartbeat() *Heartbeat {
	eom := new(Heartbeat)
	eom.Timestamp = primitives.NewTimestampNow()
	h, err := primitives.NewShaHashFromStr("deadbeef00000000000000000000000000000000000000000000000000000000")
	if err != nil {
		panic(err)
	}
	eom.DBlockHash = h
	h, err = primitives.NewShaHashFromStr("deadbeef00000000000000000000000000000000000000000000000000000000")
	if err != nil {
		panic(err)
	}
	eom.IdentityChainID = h

	return eom
}

func newSignedHeartbeat() *Heartbeat {
	ack := newHeartbeat()

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

func TestValidHeatbeat(t *testing.T) {
	s := testHelper.CreateAndPopulateTestStateAndStartValidator()
	a := identity.RandomAuthority()

	pkey := primitives.RandomPrivateKey()
	a.SigningKey = *pkey.Pub

	s.IdentityControl.Authorities[a.AuthorityChainID.Fixed()] = a

	h := newSignedHeartbeat()
	// To pass validate, we need
	// 	The timestamp to be near state timestamp
	//	The height to be high
	h.IdentityChainID = a.AuthorityChainID
	h.DBHeight = 100
	h.Timestamp = s.GetTimestamp()

	h.Signature = nil
	err := h.Sign(pkey)
	if err != nil {
		t.Error(err)
	}

	if h.Validate(s) != 1 {
		t.Errorf("Exp %d found %d", 1, h.Validate(s))
	}
}
