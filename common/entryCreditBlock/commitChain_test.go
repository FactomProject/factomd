// +build all

package entryCreditBlock_test

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"fmt"

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

func TestMiscCC(t *testing.T) {
	//chain commit from factom-cli get ecbheight 28556
	ccbytes, _ := hex.DecodeString("0001538b8480e3c5be4e952b9c5e711e1d5022580f1a600f24daa7302387dc547280162443524a3016ce3104cafd88c48545abbd4dd98e90d870039f436c0efd572c58371f06dcdb5884280d38c9f037139841253e256ba2fee183dfde1a6936b5773f1284fc400b9b7cfddf8f8209b10249dfc60e1cf5ff9252b1a1e0c5db178d3f616695b99b8eeaa83e3e1e0af73e47832127ed9e729649c8d17eb14f6c49db810a7d20a09cc68ff9ca017caa1fcc513c9b579f6e4d91c262aa70621de851559a1e80ab674b0a")
	cc := NewCommitChain()
	cc.UnmarshalBinary(ccbytes)

	expected := fmt.Sprint("db5884280d38c9f037139841253e256ba2fee183dfde1a6936b5773f1284fc40")
	got := fmt.Sprint(cc.GetEntryHash())
	if expected != got {
		t.Errorf("Commit Chain comparison failed - %v vs %v", expected, got)
	}

	expected = fmt.Sprint("c09a488f1a070332fb51b6519d49744ec5fc4335e1ab8f7002e0fa5ce7bb4c7b")
	got = fmt.Sprint(cc.Hash())
	if expected != got {
		t.Errorf("Commit Chain comparison failed - %v vs %v", expected, got)
	}

	expected = fmt.Sprint("2016-03-18 20:57:10")
	got = cc.GetTimestamp().UTCString()
	if expected != got {
		t.Errorf("Commit Chain comparison failed - %v vs %v", expected, got)
	}

	ccbytes_badsig, _ := hex.DecodeString("0001538b8480e3c5be4e952b9c5e711e1d5022580f1a600f24daa7302387dc547280162443524a3016ce3104cafd88c48545abbd4dd98e90d870039f436c0efd572c58371f06dcdb5884280d38c9f037139841253e256ba2fee183dfde1a6936b5773f1284fc400b9b7cfddf8f8209b10249dfc60e1cf5ff9252b1a1e0c5db178d3f616695b99b8eeaa83e3e1e0af73e47832127ed9e729649c8d17eb14f6c49db810a7d20a09cc68ff9ca017caa1fcc513c9b579f6e4d91c262aa70621de851559a1e80ab674b00")
	cc_badsig := NewCommitChain()
	cc_badsig.UnmarshalBinary(ccbytes_badsig)

	if nil != cc.ValidateSignatures() {
		t.Errorf("Commit Chain comparison failed")
	}

	if nil == cc_badsig.ValidateSignatures() {
		t.Errorf("Commit Chain comparison failed")
	}

	cc2 := NewCommitChain()
	cc2.UnmarshalBinary(ccbytes)

	if cc.IsSameAs(cc_badsig) {
		t.Errorf("Commit Chain comparison failed")
	}

	if !cc.IsSameAs(cc2) {
		t.Errorf("Commit Chain comparison failed")
	}
}

func TestCommitChainIsValid(t *testing.T) {
	c := NewCommitChain()
	c.Credits = 0
	c.Init()
	p, _ := primitives.NewPrivateKeyFromHex("0000000000000000000000000000000000000000000000000000000000000000")
	err := c.Sign(p.Key[:])
	if err != nil {
		t.Error(err)
	}

	if c.IsValid() {
		t.Error("Credits are 0, should be invalid")
	}

	c.Credits = 11
	err = c.Sign(p.Key[:])
	if err != nil {
		t.Error(err)
	}
	if !c.IsValid() {
		t.Error("Credits are 11, should be valid")
	}

	c.Credits = 20
	err = c.Sign(p.Key[:])
	if err != nil {
		t.Error(err)
	}
	if !c.IsValid() {
		t.Error("Credits are 20, should be valid")
	}

	c.Credits = 21
	err = c.Sign(p.Key[:])
	if err != nil {
		t.Error(err)
	}
	if c.IsValid() {
		t.Error("Credits are 21, should be invalid")
	}
}
