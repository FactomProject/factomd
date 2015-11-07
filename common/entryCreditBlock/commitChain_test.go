package entryCreditBlock_test

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	ed "github.com/FactomProject/ed25519"
	. "github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestCommitChainMarshalUnmarshal(t *testing.T) {
	fmt.Printf("---\nTestCommitChainMarshalUnmarshal\n---\n")

	cc := NewCommitChain()

	// test MarshalBinary on a zeroed CommitChain
	if p, err := cc.MarshalBinary(); err != nil {
		t.Error(err)
	} else if z := make([]byte, CommitChainSize); string(p) != string(z) {
		t.Errorf("Marshal failed on zeroed CommitChain")
	}

	// build a CommitChain for testing
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
		t.Error(err)
	} else {
		cc.ECPubKey = (*primitives.ByteSlice32)(pub)
		cc.Sig = (*primitives.ByteSlice64)(ed.Sign(privkey, cc.CommitMsg()))
	}

	// marshal and unmarshal the commit and see if it matches
	cc2 := NewCommitChain()
	p, err := cc.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("%x\n", p)
	err = cc2.UnmarshalBinary(p)
	if err != nil {
		t.Error(err)
	}

	if !cc2.IsValid() {
		t.Errorf("signature did not match after unmarshalbinary")
	}
}

func TestCommitChainMarshalUnmarshalEmpty(t *testing.T) {
	fmt.Printf("---\nTestCommitChainMarshalUnmarshalEmpty\n---\n")

	cc := NewCommitChain()

	// marshal and unmarshal the commit and see if it matches
	cc2 := NewCommitChain()
	p, err := cc.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("%x\n", p)
	err = cc2.UnmarshalBinary(p)
	if err != nil {
		t.Error(err)
	}

	// Can't be valid if it isn't signed.
	//if !cc2.IsValid() {
	//	t.Errorf("signature did not match after unmarshalbinary")
	//}
}
