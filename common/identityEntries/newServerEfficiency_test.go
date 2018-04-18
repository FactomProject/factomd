// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identityEntries_test

import (
	"encoding/hex"
	"testing"

	"fmt"

	. "github.com/FactomProject/factomd/common/identityEntries"
)

func TestNewServerEfficiencyStruct(t *testing.T) {
	parts := []string{
		"00",
		"53657276657220456666696369656e6379",
		"888888d027c59579fc47a6fc6c4a5c0409c7c39bc38a86cb5fc0069978493762",
		"1358",
		"00000000495EAA80",
		"0125b0e7fd5e68b4dec40ca0cd2db66be84c02fe6404b696c396e3909079820f61",
		"08eed980e5c1c3bfb25b6b64db02caf7421d99349fd0fe03463fe67ea1de7ca4f3fb1782365940d22561c3b1c69cd792ca024865ff25d5279eba1fa0b8856500",
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

	fmt.Printf("%x\n", nses.MarshalForSig())

	h := nses.GetChainID()
	if h.String() != "7b7dc6e511afbca5693bfe7fd2d26b2d8269be3d85f7f84bdc237f4985d6eafa" {
		t.Errorf("Wrong ChainID, expected 7b7dc6e511afbca5693bfe7fd2d26b2d8269be3d85f7f84bdc237f4985d6eafa, got %v", h.String())
	}

	err = nses.VerifySignature(nil)
	if err != nil {
		t.Errorf("%v", err)
	}
}
