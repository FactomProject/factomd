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

func TestMarshalUnmarshalCommitChain(t *testing.T) {
	cc := newCommitChain()
	hex, err := cc.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	cc2, err := UnmarshalMessage(hex)
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
}

func newCommitChain() *CommitChainMsg {
	msg := new(CommitChainMsg)

	cc := entryCreditBlock.NewCommitChain()

	cc.Version = 0
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

	return msg
}
