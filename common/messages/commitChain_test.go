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

func TestUnmarshalNilCommitChainMsg(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(CommitChainMsg)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalCommitChain(t *testing.T) {
	cc := newCommitChain()
	hex, err := cc.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	cc2, err := msgsupport.UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str := cc2.String()
	t.Logf("str - %v", str)

	if cc2.Type() != constants.COMMIT_CHAIN_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	hex2, err := cc2.(*CommitChainMsg).MarshalBinary()
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

	if cc.IsSameAs(cc2.(*CommitChainMsg)) == false {
		t.Error("CommitChainMsgs are not the same")
	}

	hex, err = cc2.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled 2 - %x", hex)
}

func TestSignAndVerifyCommitChain(t *testing.T) {
	msg := newSignedCommitChain()

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

	if msg2.Type() != constants.COMMIT_CHAIN_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	valid, err = msg2.(*CommitChainMsg).VerifySignature()
	if err != nil {
		t.Error(err)
	}
	if valid == false {
		t.Error("Signature 2 is not valid")
	}
}

func newCommitChain() *CommitChainMsg {
	msg := new(CommitChainMsg)

	cc := entryCreditBlock.NewCommitChain()

	cc.Version = 0x11
	cc.MilliTime = (*primitives.ByteSlice6)(&[6]byte{1, 1, 1, 1, 1, 1})
	p, _ := hex.DecodeString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	cc.ChainIDHash.SetBytes(p)
	p, _ = hex.DecodeString("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	cc.Weld.SetBytes(p)
	p, _ = hex.DecodeString("cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
	cc.EntryHash.SetBytes(p)
	cc.Credits = 11

	// make a key and sign the msg
	if pub, privkey, err := ed.GenerateKey(rand.Reader); err != nil {
		panic(err)
	} else {
		cc.ECPubKey = (*primitives.ByteSlice32)(pub)
		cc.Sig = (*primitives.ByteSlice64)(ed.Sign(privkey, cc.CommitMsg()))
	}

	msg.CommitChain = cc
	//msg.Timestamp = primitives.NewTimestampNow()

	return msg
}

func newSignedCommitChain() *CommitChainMsg {
	msg := newCommitChain()

	key, err := primitives.NewPrivateKeyFromHex("07c0d52cb74f4ca3106d80c4a70488426886bccc6ebc10c6bafb37bf8a65f4c38cee85c62a9e48039d4ac294da97943c2001be1539809ea5f54721f0c5477a0a")
	if err != nil {
		panic(err)
	}
	err = msg.Sign(key)
	if err != nil {
		panic(err)
	}

	return msg
}
