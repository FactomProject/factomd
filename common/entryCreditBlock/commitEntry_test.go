// +build all

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	. "github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestUnmarshalNilCommitEntry(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(CommitEntry)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestMiscEC(t *testing.T) {
	//chain commit from factom-cli get ecbheight 28556
	ecbytes, _ := hex.DecodeString("0001538b7fe6fd249f6eed5336f91eb6b506b1f4683c0e03aa8d8c59cf54299b945d41a73b44e90117ef7a21d1a616d65e6b73f3c6a7ad5c49340a6c2592872020ec60767ff00d7dc38e2fc16991f2705244c83cc36e5b4ca796dbbf168601b55d6fc34187a8de061b096f3266f3f6dd986e3f2150a1b14ada29cc9c0fc3a1d1a1875f11dc6cfd0b")
	ec := NewCommitEntry()
	ec.UnmarshalBinary(ecbytes)

	expected := fmt.Sprint("249f6eed5336f91eb6b506b1f4683c0e03aa8d8c59cf54299b945d41a73b44e9")
	got := fmt.Sprint(ec.GetEntryHash())
	if expected != got {
		t.Errorf("Entry Commit comparison failed - %v vs %v", expected, got)
	}

	expected = fmt.Sprint("9c406b5f2bf32f9cad3cb44b1dbcd6513d35979e6795984cc4f00e604a540c19")
	got = fmt.Sprint(ec.Hash())
	if expected != got {
		t.Errorf("Entry Commit comparison failed - %v vs %v", expected, got)
	}

	expected = fmt.Sprint("2016-03-18 20:52:08")
	got = ec.GetTimestamp().UTCString()
	if expected != got {
		t.Errorf("Entry Commit comparison failed - %v vs %v", expected, got)
	}

	ecbytes_badsig, _ := hex.DecodeString("0001538b7fe6fd249f6eed5336f91eb6b506b1f4683c0e03aa8d8c59cf54299b945d41a73b44e90117ef7a21d1a616d65e6b73f3c6a7ad5c49340a6c2592872020ec60767ff00d7dc38e2fc16991f2705244c83cc36e5b4ca796dbbf168601b55d6fc34187a8de061b096f3266f3f6dd986e3f2150a1b14ada29cc9c0fc3a1d1a1875f11dc6cfd00")
	ec_badsig := NewCommitEntry()
	ec_badsig.UnmarshalBinary(ecbytes_badsig)

	if nil != ec.ValidateSignatures() {
		t.Errorf("Entry Commit comparison failed")
	}

	if nil == ec_badsig.ValidateSignatures() {
		t.Errorf("Entry Commit comparison failed")
	}

	cc2 := NewCommitEntry()
	cc2.UnmarshalBinary(ecbytes)

	if ec.IsSameAs(ec_badsig) {
		t.Errorf("Entry Commit comparison failed")
	}

	if !ec.IsSameAs(cc2) {
		t.Errorf("Entry Commit comparison failed")
	}
}

func TestStringEC(t *testing.T) {
	//chain commit from factom-cli get ecbheight 28556
	ecbytes, _ := hex.DecodeString("0001538b7fe6fd249f6eed5336f91eb6b506b1f4683c0e03aa8d8c59cf54299b945d41a73b44e90117ef7a21d1a616d65e6b73f3c6a7ad5c49340a6c2592872020ec60767ff00d7dc38e2fc16991f2705244c83cc36e5b4ca796dbbf168601b55d6fc34187a8de061b096f3266f3f6dd986e3f2150a1b14ada29cc9c0fc3a1d1a1875f11dc6cfd0b")
	ec := NewCommitEntry()
	ec.UnmarshalBinary(ecbytes)

	got := fmt.Sprintf("%v\n", ec.String())
	expected := "ehash[249f6e] Credits[1] PublicKey[17ef7a] Sig[c38e2f]\n"
	if got != expected {
		t.Errorf("Entry Commit comparison failed")
	}
}

func TestCommitEntryMarshalUnmarshalStatic(t *testing.T) {
	ce := NewCommitEntry()
	data, _ := hex.DecodeString("0001538b7fe6fd249f6eed5336f91eb6b506b1f4683c0e03aa8d8c59cf54299b945d41a73b44e90117ef7a21d1a616d65e6b73f3c6a7ad5c49340a6c2592872020ec60767ff00d7dc38e2fc16991f2705244c83cc36e5b4ca796dbbf168601b55d6fc34187a8de061b096f3266f3f6dd986e3f2150a1b14ada29cc9c0fc3a1d1a1875f11dc6cfd0b")
	rest, err := ce.UnmarshalBinaryData(data)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(rest) > 0 {
		t.Error("Returned extra data")
	}
	h := ce.GetHash()
	expected := "9c406b5f2bf32f9cad3cb44b1dbcd6513d35979e6795984cc4f00e604a540c19"
	if h.String() != expected {
		t.Errorf("Wrong hash - %v vs %v", h.String(), expected)
	}

	h = ce.GetSigHash()
	expected = "29be46067fa1aa19e139a9db305d46035e24c4ff1b77c58ccb66028e70e7d180"
	if h.String() != expected {
		t.Errorf("Wrong hash - %v vs %v", h.String(), expected)
	}
}

func TestCommitEntryIsValid(t *testing.T) {
	c := NewCommitEntry()
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

	c.Credits = 10
	err = c.Sign(p.Key[:])
	if err != nil {
		t.Error(err)
	}
	if !c.IsValid() {
		t.Error("Credits are 10, should be valid")
	}

	c.Credits = 1
	err = c.Sign(p.Key[:])
	if err != nil {
		t.Error(err)
	}
	if !c.IsValid() {
		t.Error("Credits are 1, should be valid")
	}

	c.Credits = 11
	err = c.Sign(p.Key[:])
	if err != nil {
		t.Error(err)
	}
	if c.IsValid() {
		t.Error("Credits are 11, should be invalid")
	}
}
