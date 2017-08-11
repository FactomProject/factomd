// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/messages"
)

func TestUnmarshalNilBounce(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(Bounce)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestBounceAttributes(t *testing.T) {
	b := new(Bounce)

	v := b.Validate(nil)
	if v != 1 {
		t.Errorf("Should be 1, found %d", v)
	}

	b.SetValid(0)
	v = b.Validate(nil)
	if v != 0 {
		t.Errorf("Should be 0, found %d", v)
	}

	b.SetValid(-1)
	v = b.Validate(nil)
	if v != -1 {
		t.Errorf("Should be -1, found %d", v)
	}

	if b.Processed() {
		t.Error("Processed should be false")
	}

	b.FollowerExecute(nil)
	if !b.Processed() {
		t.Error("Processed should be true")
	}
}

func TestBounceMisc(t *testing.T) {
	b := new(Bounce)

	b.Process(1, nil)
	if !b.Processed() {
		t.Error("Processed should be true")
	}

	err := b.Sign(nil)
	if err != nil {
		t.Error(err)
	}

	nilSig := b.GetSignature()
	if nilSig != nil {
		t.Error("Bounce signature should be nil")
	}

	_, err = b.JSONByte()
	if err != nil {
		t.Error(err)
	}

	_, err = b.JSONString()
	if err != nil {
		t.Error(err)
	}

	sigVer, err := b.VerifySignature()
	if err != nil {
		t.Error(err)
	}
	if !sigVer {
		t.Error("VerifySignature should be true")
	}

	b2 := new(Bounce)
	if !b.IsSameAs(b2) {
		t.Error("Bounces should always be considered the same")
	}

	b2.LeaderExecute(nil)
	if !b2.Processed() {
		t.Error("Processed should be true")
	}
}
