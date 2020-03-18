// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identityEntries_test

import (
	"encoding/hex"
	"testing"

	. "github.com/FactomProject/factomd/common/identityEntries"
)

// TestRegisterServerManagementStructure checks a hardcoded external ID can be set into the regiester server management structure and obtain the correct chain ID
func TestRegisterServerManagementStructure(t *testing.T) {
	parts := []string{
		"00",
		"526567697374657220536572766572204D616E6167656D656E74",
		"8888881d59de393d9acc2b89116bc5a2dd0d0377af7a5e04bc7394149a6dbe23",
		"0125b0e7fd5e68b4dec40ca0cd2db66be84c02fe6404b696c396e3909079820f61",
		"fcb3b9dd3cc9f09b61a07e859d13a569d481508f0d5e672f9412080255ee398428fb2c488e0c3d291218f573612badf84efa63439bbcdd3ca265a31074107e04",
	}

	extIDs := [][]byte{}
	for _, v := range parts {
		b, _ := hex.DecodeString(v)
		extIDs = append(extIDs, b)
		//t.Logf("Len %v - %v", i, len(b))
	}
	rsms := new(RegisterServerManagementStructure)
	err := rsms.DecodeFromExtIDs(extIDs)
	if err != nil {
		t.Errorf("%v", err)
	}
	h := rsms.GetChainID()
	if h.String() != "925dd7acd19cee29379bf5f7ee0b60ceb2a540dca7c71490540a4b9544c7dc6e" {
		t.Errorf("Wrong ChainID, expected 925dd7acd19cee29379bf5f7ee0b60ceb2a540dca7c71490540a4b9544c7dc6e, got %v", h.String())
	}

	err = rsms.VerifySignature(nil)
	if err != nil {
		t.Errorf("%v", err)
	}
}
