// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identityEntries_test

import (
	"encoding/hex"
	"testing"

	. "github.com/PaulSnow/factom2d/common/identityEntries"
)

func TestNewCoinbaseCancelStruct(t *testing.T) {
	parts := []string{
		"00",
		"436f696e626173652043616e63656c",
		"888888d027c59579fc47a6fc6c4a5c0409c7c39bc38a86cb5fc0069978493762",
		"00030d40",
		"00000005",
		"0125b0e7fd5e68b4dec40ca0cd2db66be84c02fe6404b696c396e3909079820f61",
		"68c06b195771f801ff216c0ba98de485e54410c0765d662118aac389e319dcfdee12d11915206ab7d35f6f028584406156840fc30219111750bb1b0bc2b06106",
	}
	extIDs := [][]byte{}
	for _, v := range parts {
		b, _ := hex.DecodeString(v)
		extIDs = append(extIDs, b)
		//t.Logf("Len %v - %v", i, len(b))
	}
	nses := new(NewCoinbaseCancelStruct)
	err := nses.DecodeFromExtIDs(extIDs)
	if err != nil {
		t.Errorf("%v", err)
	}

	h := nses.GetChainID()
	if h.String() != "2d07e0f58224e0d9447e413a6cf708fde4c4cc26a7ebcdd72daba0a9e18c0ca3" {
		t.Errorf("Wrong ChainID, expected 2d07e0f58224e0d9447e413a6cf708fde4c4cc26a7ebcdd72daba0a9e18c0ca3, got %v", h.String())
	}

	err = nses.VerifySignature(nil)
	if err != nil {
		t.Errorf("%v", err)
	}

}
