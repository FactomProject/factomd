// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"crypto/rand"
	"encoding/hex"
	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"testing"
)

func TestMarshalUnmarshalCommitEntry(t *testing.T) {
	ce := newCommitEntry()
	hex, err := ce.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	ce2, err := UnmarshalMessage(hex)
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

	return cem
}
