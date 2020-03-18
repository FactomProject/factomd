// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identityEntries_test

import (
	"encoding/hex"
	"testing"

	. "github.com/FactomProject/factomd/common/identityEntries"
)

// TestNewServerEfficiencyStruct checks a hardcoded external ID can be set into the sever efficiency structure and obtain the correct chain ID
func TestNewServerEfficiencyStruct(t *testing.T) {
	parts := []string{
		"00",
		"53657276657220456666696369656e6379",
		"888888d027c59579fc47a6fc6c4a5c0409c7c39bc38a86cb5fc0069978493762",
		"1358",
		"00000000495EAA80",
		"0125b0e7fd5e68b4dec40ca0cd2db66be84c02fe6404b696c396e3909079820f61",
		"2954c40f889d49a561d0ac419741f7efd11e145a99b67485fb8f7c3e3c42d3c698d50866beffbc09032243ab3d375b4c962745c09d1a184d91e5ba69762b4e09",
	}
	extIDs := [][]byte{}
	for _, v := range parts {
		b, _ := hex.DecodeString(v)
		extIDs = append(extIDs, b)
		//t.Logf("Len %v - %v", i, len(b))
	}
	nses := new(NewServerEfficiencyStruct)
	err := nses.DecodeFromExtIDs(extIDs)
	if err != nil {
		t.Errorf("%v", err)
	}

	h := nses.GetChainID()
	if h.String() != "076651128405474b6b67c4a0c5476a0deb47732b14f699f58e92b5c011ca160e" {
		t.Errorf("Wrong ChainID, expected 076651128405474b6b67c4a0c5476a0deb47732b14f699f58e92b5c011ca160e, got %v", h.String())
	}

	err = nses.VerifySignature(nil)
	if err != nil {
		t.Errorf("%v", err)
	}

	if nses.Efficiency != 4952 {
		t.Errorf("Should be 4952, found %d", nses.Efficiency)
	}
}
