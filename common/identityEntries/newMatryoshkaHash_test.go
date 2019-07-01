// +build all 

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identityEntries_test

import (
	"encoding/hex"
	"testing"

	. "github.com/FactomProject/factomd/common/identityEntries"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestNewMatryoshkaHashStructure(t *testing.T) {
	parts := []string{
		"00",
		"4e6577204d617472796f73686b612048617368",
		"888888d027c59579fc47a6fc6c4a5c0409c7c39bc38a86cb5fc0069978493762",
		"bf1e78e5755851242a2ebf703e8bf6aca1af9dbae09ebc495cd2da220e5d370f",
		"00000000495EAA80",
		"0125b0e7fd5e68b4dec40ca0cd2db66be84c02fe6404b696c396e3909079820f61",
		"b1bc034cf75d4ebf7c4025a6b6b15c8f11a4384dcb043160711f19da9f4efb1315d84811b2247bb703732c2116b464781daf5efe75efd4adc641fee220ec660c",
	}
	extIDs := [][]byte{}
	for _, v := range parts {
		b, _ := hex.DecodeString(v)
		extIDs = append(extIDs, b)
		//t.Logf("Len %v - %v", i, len(b))
	}
	nmh := new(NewMatryoshkaHashStructure)
	err := nmh.DecodeFromExtIDs(extIDs)
	if err != nil {
		t.Errorf("%v", err)
	}
	h := nmh.GetChainID()
	if h.String() != "631d82b86861ad552b1bb3e8311a9f04960e5d966c2830f0ada4caace517a914" {
		t.Errorf("Wrong ChainID, expected 631d82b86861ad552b1bb3e8311a9f04960e5d966c2830f0ada4caace517a914, got %v", h.String())
	}

	h, err = primitives.NewShaHashFromStr("3f2b77bca02392c95149dc769a78bc758b1037b6a546011b163af0d492b1bcc0")
	if err != nil {
		t.Errorf("%v", err)
	}
	err = nmh.VerifySignature(h)
	if err != nil {
		t.Errorf("%v", err)
	}
}
