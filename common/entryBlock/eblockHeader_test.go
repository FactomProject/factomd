// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryBlock_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	. "github.com/PaulSnow/factom2d/common/entryBlock"
)

// TestUnmarshalNilEBlockHeader checks that unmarshalling the nil or empty interface throws the appropriate errors
func TestUnmarshalNilEBlockHeader(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(EBlockHeader)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

// TestEBlockHeaderMisc checks that a specific entry block can be marshalled and unmarshalled correctly, and that the unmarshalled data
// produces the same returns from miscellaineous functions in the entry block header (ie, checks integrity of the data)
func TestEBlockHeaderMisc(t *testing.T) {
	h := newEntryBlockHeader()

	if h.GetChainID().String() != "4bf71c177e71504032ab84023d8afc16e302de970e6be110dac20adbf9a19746" {
		t.Errorf("Invalid GetChainID - %v", h.GetChainID())
	}
	if h.GetBodyMR().String() != "25f25d9375533b44505964af993212ef7c13314736b2c76a37c73571d89d8b21" {
		t.Errorf("Invalid GetBodyMR - %v", h.GetBodyMR())
	}
	if h.GetPrevKeyMR().String() != "c6180f7430677d46d93a3e17b68e6a25dc89ecc092cee1459101578859f7f696" {
		t.Errorf("Invalid GetPrevKeyMR - %v", h.GetPrevKeyMR())
	}
	if h.GetPrevFullHash().String() != "9d171a092a1d04f067d55628b461c6a106b76b4bc860445f87b0052cdc5f2bfd" {
		t.Errorf("Invalid GetPrevFullHash - %v", h.GetPrevFullHash())
	}
	if h.GetEBSequence() != 728 {
		t.Errorf("Invalid GetEBSequence - %v", h.GetEBSequence())
	}
	if h.GetDBHeight() != 6920 {
		t.Errorf("Invalid GetDBHeight - %v", h.GetDBHeight())
	}
	if h.GetEntryCount() != 2 {
		t.Errorf("Invalid GetEntryCount - %v", h.GetEntryCount())
	}

	data, err := h.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	h2 := new(EBlockHeader)
	nd, err := h2.UnmarshalBinaryData(data)
	if err != nil {
		t.Error(err)
	}
	if len(nd) > 0 {
		t.Errorf("Should be no bytes left, found %d", len(nd))
	}

	if h.String() != h2.String() {
		t.Error("Strings should be the same")
	}

	j1, err := h.JSONString()
	if err != nil {
		t.Error(err)
	}
	j2, err := h2.JSONString()
	if err != nil {
		t.Error(err)
	}

	if j1 != j2 {
		t.Error("JsonStrings should be the same")
	}

	jb1, err := h.JSONByte()
	if err != nil {
		t.Error(err)
	}
	jb2, err := h2.JSONByte()
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(jb1, jb2) != 0 {
		t.Error("JsonBytes should be the same")
	}
}

// newEntryBlockHeader creates a new entry block header for testing
func newEntryBlockHeader() *EBlockHeader {
	e := NewEBlock()
	entryStr := "4bf71c177e71504032ab84023d8afc16e302de970e6be110dac20adbf9a1974625f25d9375533b44505964af993212ef7c13314736b2c76a37c73571d89d8b21c6180f7430677d46d93a3e17b68e6a25dc89ecc092cee1459101578859f7f6969d171a092a1d04f067d55628b461c6a106b76b4bc860445f87b0052cdc5f2bfd000002d800001b080000000272d72e71fdee4984ecb30eedcc89cb171d1f5f02bf9a8f10a8b2cfbaf03efe1c0000000000000000000000000000000000000000000000000000000000000001"
	h, err := hex.DecodeString(entryStr)
	if err != nil {
		panic(err)
	}
	err = e.UnmarshalBinary(h)
	if err != nil {
		panic(err)
	}
	return e.Header.(*EBlockHeader)
}
