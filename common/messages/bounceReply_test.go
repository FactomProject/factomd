// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"testing"

	. "github.com/PaulSnow/factom2d/common/messages"
)

func TestUnmarshalNilBounceReply(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(BounceReply)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestBounceReplyMisc(t *testing.T) {
	b := new(BounceReply)

	processed := b.Process(1, nil)
	if !processed {
		t.Error("Process did not return true")
	}

	validation := b.Validate(nil)
	if validation != 1 {
		t.Error("Validate did not return 1")
	}

	err := b.Sign(nil)
	if err != nil {
		t.Error(err)
	}

	nilSig := b.GetSignature()
	if nilSig != nil {
		t.Error("BounceReply signature should be nil")
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

	b2 := new(BounceReply)
	if !b.IsSameAs(b2) {
		t.Error("BounceReplies should always be considered the same")
	}
}
