// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
)

//"testing"

//. "github.com/FactomProject/factomd/common/messages"

func newRemoveServer() *RemoveServerMsg {
	addserv := new(RemoveServerMsg)
	addserv.Timestamp = primitives.NewTimestampNow()
	addserv.ServerChainID = primitives.Sha([]byte("FNode0"))
	addserv.ServerType = 0
	addserv.Signatures = factoid.NewFullSignatureBlock()
	return addserv
}

func TestRemoveServerMultisigAdd(t *testing.T) {
	msg := newRemoveServer()

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

func TestRemoveServerSigVerify(t *testing.T) {
	msg := newRemoveServer()

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

	msg = new(RemoveServerMsg)
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

func TestRemoveServerMultisigValidate(t *testing.T) {
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

	msg := newRemoveServer()
	msg.ServerChainID = authorities[0].GetAuthorityChainID()

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
