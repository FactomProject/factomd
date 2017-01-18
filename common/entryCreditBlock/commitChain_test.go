package entryCreditBlock_test

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	ed "github.com/FactomProject/ed25519"
	. "github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestUnmarshalNilCommitChain(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(CommitChain)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestCommitChainMarshalUnmarshal(t *testing.T) {
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

func TestCommitChainMarshalUnmarshalStatic(t *testing.T) {
	cc := NewCommitChain()
	data, _ := hex.DecodeString("000155ba3b3b3ae027e70117c916df3525ef18f602a7f2faf6c797029b5b2e608666c0d23f78b7ba54c4f60eb6303ab967c7a1e56ec2431d42cd1b373dc40884979e8c0aa21b0938d6fa3bfee0b0e21de90585092a99ac6b903d361491fffc7ede7336300049f30b3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da2927f1388aa1c9ab6a52d7689665010cff815c422d4a2a613ebc826f272f4a2a3ff9020e51e615a9ebce9dce99e36d30a7fbbe67855508e60a2233a39eef832204")
	rest, err := cc.UnmarshalBinaryData(data)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(rest) > 0 {
		t.Error("Returned extra data")
	}
	h := cc.GetHash()
	expected := "0dc03fe1a046afad51a7532a97fe2f02d03a2b0bddd3fdacc7641573cf17cfea"
	if h.String() != expected {
		t.Errorf("Wrong hash - %v vs %v", h.String(), expected)
	}

	h = cc.GetSigHash()
	expected = "7326c1fe594589f6b0d15b7204d1d8161fd432c821b0c3ee4eb7c6aa31e97062"
	if h.String() != expected {
		t.Errorf("Wrong hash - %v vs %v", h.String(), expected)
	}
}

func TestCommitChainMarshalUnmarshalEmpty(t *testing.T) {
	cc := NewCommitChain()

	// marshal and unmarshal the commit and see if it matches
	cc2 := NewCommitChain()
	p, err := cc.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("%x\n", p)
	err = cc2.UnmarshalBinary(p)
	if err == nil {
		t.Error("Error is nil when it shouldn't be")
	}
}
