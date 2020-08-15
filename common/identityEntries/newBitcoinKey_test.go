// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identityEntries_test

import (
	"encoding/hex"
	"testing"

	. "github.com/PaulSnow/factom2d/common/identityEntries"
)

func TestNewBitcoinKeyStructure(t *testing.T) {
	parts := []string{
		"00",
		"4e657720426974636f696e204b6579",
		"888888d027c59579fc47a6fc6c4a5c0409c7c39bc38a86cb5fc0069978493762",
		"00",
		"00",
		"c5b7fd920dce5f61934e792c7e6fcc829aff533d",
		"00000000495EAA80",
		"0125b0e7fd5e68b4dec40ca0cd2db66be84c02fe6404b696c396e3909079820f61",
		"379d64dd36ba724539ce19adb05b9a6a98cc3e3171785553e2985f5542a3ce3bf470ef78a884eee2ba75c9f2cfa64f21d3ace4dc981daeb3c00352dbb19a1e0c",
	}
	extIDs := [][]byte{}
	for _, v := range parts {
		b, _ := hex.DecodeString(v)
		extIDs = append(extIDs, b)
		//t.Logf("Len %v - %v", i, len(b))
	}
	nbks := new(NewBitcoinKeyStructure)
	err := nbks.DecodeFromExtIDs(extIDs)
	if err != nil {
		t.Errorf("%v", err)
	}
	h := nbks.GetChainID()
	if h.String() != "18665ed9cc13f6b664033aa45179002f391395657ebc6e8a755a0658980580c6" {
		t.Errorf("Wrong ChainID, expected 18665ed9cc13f6b664033aa45179002f391395657ebc6e8a755a0658980580c6, got %v", h.String())
	}

	err = nbks.VerifySignature(nil)
	if err != nil {
		t.Errorf("%v", err)
	}
}
