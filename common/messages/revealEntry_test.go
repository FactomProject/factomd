// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/entryBlock"
	"github.com/PaulSnow/factom2d/common/entryCreditBlock"
	"github.com/PaulSnow/factom2d/common/interfaces"
	. "github.com/PaulSnow/factom2d/common/messages"
	. "github.com/PaulSnow/factom2d/common/messages/msgsupport"
	"github.com/PaulSnow/factom2d/common/primitives"
	"github.com/PaulSnow/factom2d/common/primitives/random"
	"github.com/PaulSnow/factom2d/state"
	"github.com/PaulSnow/factom2d/testHelper"
)

var _ = fmt.Println

func TestUnmarshalNilRevealEntryMsg(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(RevealEntryMsg)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshalRevealEntry(t *testing.T) {
	re := newRevealEntry()
	hex, err := re.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	re2, err := UnmarshalMessage(hex)
	if err != nil {
		t.Error(err)
	}
	str := re2.String()
	t.Logf("str - %v", str)

	if re2.Type() != constants.REVEAL_ENTRY_MSG {
		t.Errorf("Invalid message type unmarshalled - got %v, expected %v", re2.Type(), constants.REVEAL_ENTRY_MSG)
	}

	hex2, err := re2.(*RevealEntryMsg).MarshalBinary()
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

func newRevealEntry() *RevealEntryMsg {
	re := new(RevealEntryMsg)

	entry := new(entryBlock.Entry)

	entry.ExtIDs = make([]primitives.ByteSlice, 0, 5)
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: []byte("1asdfadfasdf")})
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: []byte("")})
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: []byte("3")})
	entry.ChainID = new(primitives.Hash)
	entry.ChainID.SetBytes(constants.EC_CHAINID)

	entry.Content = primitives.ByteSlice{Bytes: []byte("1asdf asfas dfsg\"08908098(*)*^*&%&%&$^#%##%$$@$@#$!$#!$#@!~@!#@!%#@^$#^&$*%())_+_*^*&^&\"\"?>?<<>/./,")}

	re.Entry = entry

	return re
}

func TestValidRevealMsg(t *testing.T) {
	s := testHelper.CreateAndPopulateTestStateAndStartValidator()

	if v := testValid(1, 0, s); v != -2 {
		t.Error("Should be -2 found ", v)
	}

	if v := testValid(15, 12000, s); v != -1 {
		t.Error("Should be -1, found ", v)
	}

	if v := testValid(0, 12000, s); v != -1 {
		t.Error("Should be -1, found ", v)
	}
}

func testValid(ecs uint8, dataSize int, s *state.State) int {
	com := NewCommitEntryMsg()
	com.CommitEntry = entryCreditBlock.NewCommitEntry()

	m := newRevealEntryWithContentSizeX(dataSize)
	com.CommitEntry.Credits = ecs
	com.CommitEntry.EntryHash = m.Entry.GetHash()
	s.PutCommit(m.Entry.GetHash(), com)

	return m.Validate(s)
}

func newRevealEntryWithContentSizeX(size int) *RevealEntryMsg {
	re := new(RevealEntryMsg)

	entry := new(entryBlock.Entry)

	entry.ExtIDs = make([]primitives.ByteSlice, 0, 5)
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: []byte("1asdfadfasdf")})
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: []byte("")})
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: []byte("3")})
	entry.ChainID = new(primitives.Hash)
	entry.ChainID.SetBytes(constants.EC_CHAINID)

	entry.Content = primitives.ByteSlice{Bytes: random.RandByteSliceOfLen(size)}
	entry.ChainID = entryBlock.NewChainID(entry)

	re.Entry = entry

	return re
}

func newMaliciousRevealEntry() *RevealEntryMsg {
	re := new(RevealEntryMsg)

	entry := new(entryBlock.Entry)

	entry.ExtIDs = make([]primitives.ByteSlice, 0, 5)
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: []byte("ThisDoes")})
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: []byte("NotHash")})
	entry.ExtIDs = append(entry.ExtIDs, primitives.ByteSlice{Bytes: []byte("To5f5a0fa")})
	entry.ChainID = new(primitives.Hash)
	p, _ := hex.DecodeString("5f5a0fa853e7a84752fa90546915d4ab3c1e031e2cb785cfbcb2d93c211fea0b")
	entry.ChainID.SetBytes(p)

	entry.Content = primitives.ByteSlice{Bytes: []byte("1asdfadfasdf")}

	re.Entry = entry

	return re
}

func TestRevealMaliciousFirstEntryReveal(t *testing.T) {
	testState := testHelper.CreateAndPopulateTestStateAndStartValidator()

	m := newMaliciousRevealEntry()
	goodEntry := newSignedCommitChain()

	testState.PutCommit(m.Entry.GetHash(), goodEntry)

	if m.Validate(testState) > -1 {
		t.Error("Malicious RevealEntry message improperly considered valid (hash of extIDs does not match chainID)")
	}
}

func TestRevealEntry2ChainID(t *testing.T) {

	type test struct {
		ExternalIDs [][]byte
		ChainID     interfaces.IHash
	}

	var workingTests []*test

	at := new(test)
	at.ExternalIDs = append(at.ExternalIDs, []byte("test"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("0"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("10307"))
	h, _ := hex.DecodeString("3e66337d3d1262e13c557fde662a2fbd015410501cc8c3c8095bc182dfcf40f6")
	at.ChainID = primitives.NewHash(h)
	workingTests = append(workingTests, at)

	at = new(test)
	at.ExternalIDs = append(at.ExternalIDs, []byte("test"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("1"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("19484"))
	h, _ = hex.DecodeString("e92ed553b45af51c2a124720e73edb4ab8c0d81176f0c5327c070a1ad823292b")
	at.ChainID = primitives.NewHash(h)
	workingTests = append(workingTests, at)

	at = new(test)
	at.ExternalIDs = append(at.ExternalIDs, []byte("test"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("2"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("3151"))
	h, _ = hex.DecodeString("548a349a671b12777c575c1f47465c29c6bd0822b2bfee8683db36087f353a02")
	at.ChainID = primitives.NewHash(h)
	workingTests = append(workingTests, at)

	at = new(test)
	at.ExternalIDs = append(at.ExternalIDs, []byte("test"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("3"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("2062"))
	h, _ = hex.DecodeString("7b226a7a4821b522f37e5fc3f6abe4f5e86bc1c3f5967e398d3212e6543af8b3")
	at.ChainID = primitives.NewHash(h)
	workingTests = append(workingTests, at)

	at = new(test)
	at.ExternalIDs = append(at.ExternalIDs, []byte("test"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("4"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("9417"))
	h, _ = hex.DecodeString("8d1d56ab04de47e1b1b7f438269f01c951516c289abb7531c5bd427355910b05")
	at.ChainID = primitives.NewHash(h)
	workingTests = append(workingTests, at)

	at = new(test)
	at.ExternalIDs = append(at.ExternalIDs, []byte("test"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("5"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("6697"))
	h, _ = hex.DecodeString("2238ec3a3b1430057c29f35e6464d51b5cf1e72a73fb9e6bd0ecd90f7e65741f")
	at.ChainID = primitives.NewHash(h)
	workingTests = append(workingTests, at)

	state := new(state.State)
	state.FactomNodeName = "me"
	msg := new(RevealEntryMsg)
	msg.CommitChain = new(CommitChainMsg)
	msg.CommitChain.CommitChain = new(entryCreditBlock.CommitChain)
	for i, wt := range workingTests {
		msg.CommitChain.CommitChain.ChainIDHash = primitives.Shad(wt.ChainID.Bytes())
		if !CheckChainID(state, wt.ExternalIDs, msg) {
			t.Error("Failed to match the External IDs to the ChainID as expected, in test", i)
		}
	}

	var failingTests []*test

	// Different number of External IDs
	at = new(test)
	at.ExternalIDs = append(at.ExternalIDs, []byte("test"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("0"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("10307"))
	at.ExternalIDs = append(at.ExternalIDs, []byte(""))
	h, _ = hex.DecodeString("3e66337d3d1262e13c557fde662a2fbd015410501cc8c3c8095bc182dfcf40f6")
	at.ChainID = primitives.NewHash(h)
	failingTests = append(failingTests, at)

	// Diff External ID
	at = new(test)
	at.ExternalIDs = append(at.ExternalIDs, []byte("test"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("2"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("19484"))
	h, _ = hex.DecodeString("e92ed553b45af51c2a124720e73edb4ab8c0d81176f0c5327c070a1ad823292b")
	at.ChainID = primitives.NewHash(h)
	failingTests = append(failingTests, at)

	// Diff Chain ID
	at = new(test)
	at.ExternalIDs = append(at.ExternalIDs, []byte("test"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("2"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("3151"))
	h, _ = hex.DecodeString("548a349a671b12777c575c1f47365c29c6bd0822b2bfee8683db36087f353a02")
	at.ChainID = primitives.NewHash(h)
	failingTests = append(failingTests, at)

	// Diff Chain ID
	at = new(test)
	at.ExternalIDs = append(at.ExternalIDs, []byte("test"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("3"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("2062"))
	h, _ = hex.DecodeString("7b226afa4821b522f37e5fc3f6cbe4f5e86bc1c3f5967e398d3212e6543af8b3")
	at.ChainID = primitives.NewHash(h)
	failingTests = append(failingTests, at)

	// Diff External ID
	at = new(test)
	at.ExternalIDs = append(at.ExternalIDs, []byte("test"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("4"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("94170"))
	h, _ = hex.DecodeString("8d1d56ab04de47e1b1b7f438269f01c951516c289abb7531c5bd427355910b05")
	at.ChainID = primitives.NewHash(h)
	failingTests = append(failingTests, at)

	// Diff External ID
	at = new(test)
	at.ExternalIDs = append(at.ExternalIDs, []byte("testing"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("5"))
	at.ExternalIDs = append(at.ExternalIDs, []byte("6697"))
	h, _ = hex.DecodeString("2238ec3a3b1430057c29f35e6464d51b5cf1e72a73fb9e6bd0ecd90f7e65741f")
	at.ChainID = primitives.NewHash(h)
	failingTests = append(failingTests, at)

	for i, wt := range failingTests {
		msg.CommitChain.CommitChain.ChainIDHash = primitives.Shad(wt.ChainID.Bytes())
		if CheckChainID(state, wt.ExternalIDs, msg) {
			t.Error("Failed to detect a missmatch in the External IDs to the ChainID as expected, in test", i)
		}
	}

}
