// +build all 

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestUnmarshalNilCommitEntryMsg(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(CommitEntryMsg)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalCommitEntry(t *testing.T) {
	ce := newCommitEntry()
	hex, err := ce.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	ce2, err := msgsupport.UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str := ce2.String()
	t.Logf("str - %v", str)

	if ce2.Type() != constants.COMMIT_ENTRY_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	hex2, err := ce2.(*CommitEntryMsg).MarshalBinary()
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

	if ce.IsSameAs(ce2.(*CommitEntryMsg)) != true {
		t.Errorf("CommitEntryMsg messages are not identical")
	}
}

func TestMarshalUnmarshalSignedCommitEntry(t *testing.T) {
	msg := newSignedCommitEntry()

	str, err := msg.JSONString()
	if err != nil {
		t.Error(err)
	}
	t.Logf("str1 - %v", str)
	hex, err := msg.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

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
	str, err = msg2.JSONString()
	if err != nil {
		t.Error(err)
	}
	t.Logf("str2 - %v", str)

	if msg2.Type() != constants.COMMIT_ENTRY_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	if msg.IsSameAs(msg2.(*CommitEntryMsg)) != true {
		t.Errorf("CommitEntryMsg messages are not identical")
	}

	valid, err = msg2.(*CommitEntryMsg).VerifySignature()
	if err != nil {
		t.Error(err)
	}
	if valid == false {
		t.Error("Signature is not valid")
	}
}

func newCommitEntry() *CommitEntryMsg {
	cem := new(CommitEntryMsg)

	ce := entryCreditBlock.NewCommitEntry()

	// build a CommitEntry for testing
	ce.Version = 0
	ce.MilliTime = (*primitives.ByteSlice6)(&[6]byte{1, 1, 1, 1, 1, 1})
	p, _ := hex.DecodeString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	ce.EntryHash.SetBytes(p)
	ce.Credits = 1

	// make a key and sign the msg
	if pub, privkey, err := ed.GenerateKey(rand.Reader); err != nil {
		panic(err)
	} else {
		ce.ECPubKey = (*primitives.ByteSlice32)(pub)
		ce.Sig = (*primitives.ByteSlice64)(ed.Sign(privkey, ce.CommitMsg()))
	}

	cem.CommitEntry = ce
	//cem.Timestamp = primitives.NewTimestampNow()

	return cem
}

func newSignedCommitEntry() *CommitEntryMsg {
	addserv := newCommitEntry()

	key, err := primitives.NewPrivateKeyFromHex("07c0d52cb74f4ca3106d80c4a70488426886bccc6ebc10c6bafb37bf8a65f4c38cee85c62a9e48039d4ac294da97943c2001be1539809ea5f54721f0c5477a0a")
	if err != nil {
		panic(err)
	}
	err = addserv.Sign(key)
	if err != nil {
		panic(err)
	}

	return addserv
}
