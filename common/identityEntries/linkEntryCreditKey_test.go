// +build all

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identityEntries_test

/*
import (
	"encoding/hex"
	"testing"

	. "github.com/FactomProject/factomd/common/identityEntries"
)

func TestLinkEntryCreditKeyStructure(t *testing.T) {
	parts := []string{
		"00",
		"526567697374657220466163746F6D204964656E74697479",
		"888888d00082a172e4f0c8d03a83d327b4197e68bcc36e88eeefb00b6cec7936",
		"0125b0e7fd5e68b4dec40ca0cd2db66be84c02fe6404b696c396e3909079820f61",
		"aab1cbbd72c8b7db32f45cb89e511793f8d47e0551665679a25ef8444248e045f858701351e0cc17aeb74e4f6aa425ee71663d3a4ca6abfe6fac88d66e0c2c01",
	}

	extIDs := [][]byte{}
	for _, v := range parts {
		b, _ := hex.DecodeString(v)
		extIDs = append(extIDs, b)
		//t.Logf("Len %v - %v", i, len(b))
	}
	msc := new(LinkEntryCreditKeyStructure)
	err := msc.DecodeFromExtIDs(extIDs)
	if err != nil {
		t.Errorf("%v", err)
	}
	h := msc.GetChainID()
	if h.String() != "8888881d59de393d9acc2b89116bc5a2dd0d0377af7a5e04bc7394149a6dbe23" {
		t.Errorf("Wrong ChainID, expected 8888881d59de393d9acc2b89116bc5a2dd0d0377af7a5e04bc7394149a6dbe23, got %v", h.String())
	}
}
*/
