// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/state"
)

func TestUnmarshalNilAddServerMsg(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(AddServerMsg)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalAddServer(t *testing.T) {
	addserv := newAddServer()

	str, err := addserv.JSONString()
	if err != nil {
		t.Error(err)
	}
	t.Logf("str1 - %v", str)
	hex, err := addserv.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	addserv2, err := msgsupport.UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str, err = addserv2.JSONString()
	if err != nil {
		t.Error(err)
	}
	t.Logf("str2 - %v", str)

	if addserv2.Type() != constants.ADDSERVER_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	if addserv.IsSameAs(addserv2.(*AddServerMsg)) != true {
		t.Errorf("AddServer messages are not identical")
	}
}

func TestMarshalUnmarshalSignedAddServer(t *testing.T) {
	addserv := newSignedAddServer()

	str, err := addserv.JSONString()
	if err != nil {
		t.Error(err)
	}
	t.Logf("str1 - %v", str)
	hex, err := addserv.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	valid, err := addserv.VerifySignatures()
	if err != nil {
		t.Error(err)
	}
	if len(valid) != len(addserv.GetSignatures()) {
		t.Error("Some signatures not valid")
	}

	addserv2, err := msgsupport.UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str, err = addserv2.JSONString()
	if err != nil {
		t.Error(err)
	}
	t.Logf("str2 - %v", str)

	if addserv2.Type() != constants.ADDSERVER_MSG {
		t.Error("Invalid message type unmarshalled")
	}

	if addserv.IsSameAs(addserv2.(*AddServerMsg)) != true {
		t.Errorf("AddServer messages are not identical")
	}

	valid, err = addserv2.(*AddServerMsg).VerifySignatures()
	if err != nil {
		t.Error(err)
	}
	if len(valid) != len(addserv2.(*AddServerMsg).GetSignatures()) {
		t.Error("Signatures 2 are not valid")
	}
}

func newAddServer() *AddServerMsg {
	addserv := new(AddServerMsg)
	addserv.Timestamp = primitives.NewTimestampNow()
	addserv.ServerChainID = primitives.Sha([]byte("FNode0"))
	addserv.ServerType = 0
	addserv.Signatures = factoid.NewFullSignatureBlock()
	return addserv
}

func newSignedAddServer() *AddServerMsg {
	addserv := newAddServer()

	key, err := primitives.NewPrivateKeyFromHex("07c0d52cb74f4ca3106d80c4a70488426886bccc6ebc10c6bafb37bf8a65f4c38cee85c62a9e48039d4ac294da97943c2001be1539809ea5f54721f0c5477a0a")
	if err != nil {
		panic(err)
	}
	err = addserv.AddSignature(key)
	if err != nil {
		panic(err)
	}

	return addserv
}

// TODO: Add test for signed messages (See ack_test.go)

func TestMultisigAdd(t *testing.T) {
	msg := newAddServer()

	keys := make([]*primitives.PrivateKey, 64)
	for i := range keys {
		keys[i] = primitives.RandomPrivateKey()
		msg.AddSignature(keys[i])
	}

	data, err := msg.MarshalForSignature()
	if err != nil {
		t.Error(err)
	}

	unique := make(map[string]bool)
	for i, sig := range msg.GetSignatures() {
		unique[fmt.Sprintf("%x", sig.GetKey())] = true
		found := false
		for _, key := range keys {
			if bytes.Equal(sig.GetKey(), key.Public()) {
				if err := primitives.VerifySignature(data, key.Public(), sig.Bytes()); err != nil {
					t.Errorf("signature #%d (%x) did not verify", i, sig.GetKey())
				} else {
					found = true
					break
				}
			}
		}

		if !found {
			t.Errorf("signature #%d (%x) was not part of the original keyset", i, sig.GetKey())
		}
	}

	if len(unique) != len(keys) {
		t.Errorf("only found %d of %d signatures", len(unique), len(keys))
	}
}

func TestSigVerify(t *testing.T) {
	msg := newAddServer()

	// key 0 ok
	// key 1 ok
	// key 2 corrupt
	// key 3 duplicate of 1
	keys := make([]*primitives.PrivateKey, 4)
	for i := range keys {
		if i == 3 {
			keys[i] = keys[1]
		} else {
			keys[i] = primitives.RandomPrivateKey()
		}
		msg.AddSignature(keys[i])
	}

	// unmarshal and remarshal is as easy as getting through all the interfaces
	raw, err := msg.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	// corrupt the third of four signatures
	raw[1+6+32+1+(32+64)+(32+64)+32]++

	msg = new(AddServerMsg)
	if err := msg.UnmarshalBinary(raw); err != nil {
		t.Error("failed to unmarshal corrupted data")
	}

	sigs := msg.GetSignatures()
	if vsigs, err := msg.VerifySignatures(); err != nil {
		t.Error("failed to verify sigs", err)
	} else {
		if len(sigs) != len(vsigs)+2 {
			t.Errorf("unexpected sig count. want = (4,2), got = (%d, %d)", len(sigs), len(vsigs))
		} else {
			indexes := make(map[int]bool)
			for i, k := range keys {
				for _, sig := range vsigs {
					if bytes.Equal(k.Public(), sig.GetKey()) {
						indexes[i] = true
					}
				}
			}

			if indexes[2] {
				t.Error("the sig we corrupted was inside", indexes)
			}
			// index 3 is also set because the signature for that key exists (duplicate)
			// this is just an oddity in the test script, duplicate detection is proven
			// via the length check

			for _, i := range []int{0, 1} {
				if !indexes[i] {
					t.Errorf("sig %d missing", i)
				}
			}
		}
	}
}

var authorities []interfaces.IAuthority

type fakeState struct {
	state.State
}

func (fs *fakeState) GetAuthorities() []interfaces.IAuthority {
	return authorities
}

func TestMultisigValidate(t *testing.T) {
	var s interfaces.IState
	fs := new(fakeState)
	fs.IdentityChainID = primitives.ZeroHash

	s = fs

	// generate keys and authorities)
	keys := make([]*primitives.PrivateKey, 32)
	authorities = make([]interfaces.IAuthority, 32)
	for i := range authorities {
		keys[i] = primitives.RandomPrivateKey()
		authorities[i] = identity.RandomAuthority()
		authorities[i].(*identity.Authority).SigningKey = *keys[i].Pub
	}

	msg := newAddServer()

	if v := msg.Validate(s); v >= 0 {
		t.Fatalf("message validated with 0/32 sigs: %d", v)
	}

	for i := 0; i < 16; i++ {
		msg.AddSignature(keys[i])
		if v := msg.Validate(s); v >= 0 {
			t.Fatalf("message validated with %d/32 sigs: %d", len(msg.GetSignatures()), v)
		}
	}

	for i := 16; i < len(authorities); i++ {
		msg.AddSignature(keys[i])
		if v := msg.Validate(s); v < 1 {
			verlen, err := msg.VerifySignatures()
			if err != nil {
				t.Error(err)
			}
			t.Errorf("message failed to validate with %d/32 (%d valid) sigs: %d", len(msg.GetSignatures()), len(verlen), v)
		}
	}

}
