// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identityEntries_test

import (
	"encoding/hex"
	"testing"

	. "github.com/FactomProject/factomd/common/identityEntries"
)

// TestServerManagementStructure checks a hardcoded external ID can be set into the server management structure and obtain the correct chain ID
func TestServerManagementStructure(t *testing.T) {
	parts := []string{
		"00",
		"536572766572204D616E6167656D656E74",
		"888888d027c59579fc47a6fc6c4a5c0409c7c39bc38a86cb5fc0069978493762",
		"98765432103e2fbb",
	}

	extIDs := [][]byte{}
	for _, v := range parts {
		b, _ := hex.DecodeString(v)
		extIDs = append(extIDs, b)
		//t.Logf("Len %v - %v", i, len(b))
	}
	sm := new(ServerManagementStructure)
	err := sm.DecodeFromExtIDs(extIDs)
	if err != nil {
		t.Errorf("%v", err)
	}
	h := sm.GetChainID()
	if h.String() != "8888881d59de393d9acc2b89116bc5a2dd0d0377af7a5e04bc7394149a6dbe23" {
		t.Errorf("Wrong ChainID, expected 8888881d59de393d9acc2b89116bc5a2dd0d0377af7a5e04bc7394149a6dbe23, got %v", h.String())
	}
}
