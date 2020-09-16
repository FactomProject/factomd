// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identityEntries_test

import (
	"encoding/hex"
	"testing"

	. "github.com/FactomProject/factomd/common/identityEntries"
)

// TestIdentityChainStructure creates a new external ID and reads it into a new IdentityChainStructure, and checks its chain ID is correct
func TestIdentityChainStructure(t *testing.T) {
	parts := []string{
		"00",
		"4964656E7469747920436861696E",
		"3f2b77bca02392c95149dc769a78bc758b1037b6a546011b163af0d492b1bcc0",
		"58190cd60b8a3dd32f3e836e8f1f0b13e9ca1afff16416806c798f8d944c2c72",
		"b246833125481636108cedc2961338c1368c41c73e2c6e016e224dfe41f0ac23",
		"12db35739303a13861c14862424e90f116a594eaee25811955423dce33e500b6",
		"0000000000c512c7",
	}
	extIDs := [][]byte{}
	for _, v := range parts {
		b, _ := hex.DecodeString(v)
		extIDs = append(extIDs, b)
		//t.Logf("Len %v - %v", i, len(b))
	}
	ics := new(IdentityChainStructure)
	err := ics.DecodeFromExtIDs(extIDs)
	if err != nil {
		t.Errorf("%v", err)
	}
	h := ics.GetChainID()
	if h.String() != "888888d027c59579fc47a6fc6c4a5c0409c7c39bc38a86cb5fc0069978493762" {
		t.Errorf("Wrong ChainID, expected 888888d027c59579fc47a6fc6c4a5c0409c7c39bc38a86cb5fc0069978493762, got %v", h.String())
	}

	ics2, err := DecodeIdentityChainStructureFromExtIDs(extIDs)
	if err != nil {
		t.Errorf("%v", err)
	}
	h2 := ics2.GetChainID()
	if h2.String() != "888888d027c59579fc47a6fc6c4a5c0409c7c39bc38a86cb5fc0069978493762" {
		t.Errorf("Wrong ChainID #2, expected 888888d027c59579fc47a6fc6c4a5c0409c7c39bc38a86cb5fc0069978493762, got %v", h.String())
	}
}
