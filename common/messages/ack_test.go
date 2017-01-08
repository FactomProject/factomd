// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/messages"

	"github.com/FactomProject/factomd/common/primitives"
)

func TestMarshalUnmarshalAck(t *testing.T) {
	ack := newSignedAck()
	str, err := ack.JSONString()
	if err != nil {
		t.Error(err)
	}
	t.Logf("str1 - %v", str)
	hex, err := ack.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	ack2, err := UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str, err = ack2.JSONString()
	if err != nil {
		t.Error(err)
	}
	t.Logf("str2 - %v", str)

	if ack2.Type() != constants.ACK_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	if ack.IsSameAs(ack2.(*Ack)) == false {
		t.Error("Acks are not the same")
	}
}

func TestSignAndVerifyAck(t *testing.T) {
	ack := newSignedAck()

	hex, err := ack.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	t.Logf("Sig - %x", *ack.Signature.GetSignature())
	if len(*ack.Signature.GetSignature()) == 0 {
		t.Error("Signature not present")
	}

	valid, err := ack.VerifySignature()
	if err != nil {
		t.Error(err)
	}
	if valid == false {
		t.Error("Signature is not valid")
	}

	ack2, err := UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}

	if ack2.Type() != constants.ACK_MSG {
		t.Error("Invalid message type unmarshalled")
	}
	ackProper := ack2.(*Ack)

	valid, err = ackProper.VerifySignature()
	if err != nil {
		t.Error(err)
	}
	if valid == false {
		t.Error("Signature 2 is not valid")
	}
}

func newAck() *Ack {
	ack := new(Ack)
	ack.Timestamp = primitives.NewTimestampNow()
	hash, err := primitives.NewShaHashFromStr("cbd3d09db6defdc25dfc7d57f3479b339a077183cd67022e6d1ef6c041522b40")
	if err != nil {
		panic(err)
	}
	ack.MessageHash = hash

	hash, err = primitives.NewShaHashFromStr("bbd3d09db6defdc25dfc7d57f3479b339a077183cd67022e6d1ef6c041522b40")
	if err != nil {
		panic(err)
	}
	ack.MessageHash = hash

	ack.DBHeight = 123
	ack.Height = 456

	hash, err = primitives.NewShaHashFromStr("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		panic(err)
	}
	ack.SerialHash = hash

	hash, err = primitives.NewShaHashFromStr("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	if err != nil {
		panic(err)
	}
	ack.LeaderChainID = hash

	return ack
}

func newSignedAck() *Ack {
	ack := newAck()

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
