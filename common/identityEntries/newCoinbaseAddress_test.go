// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identityEntries_test

import (
	"encoding/hex"
	"testing"

	. "github.com/FactomProject/factomd/common/identityEntries"
)

// TestNewCoinbaseAddressStruct checks if a hardcoded external ID can be set into the coinbase address structure and obtain the correct chain ID
func TestNewCoinbaseAddressStruct(t *testing.T) {
	parts := []string{
		"00",
		"436F696E626173652041646472657373",
		"888888d027c59579fc47a6fc6c4a5c0409c7c39bc38a86cb5fc0069978493762",
		"031cce24bcc43b596af105167de2c03603c20ada3314a7cfb47befcad4883e6f",
		"00000000495EAA80",
		"0125b0e7fd5e68b4dec40ca0cd2db66be84c02fe6404b696c396e3909079820f61",
		"e08f8c763b1512d05bb6a6cf503e884a24ea6b7af0d30df1dff30444a9b9ba2db20d40555afddfcd5e03f737afaa7be78b6129787d9a561417531d263eaabb04",
	}
	extIDs := [][]byte{}
	for _, v := range parts {
		b, _ := hex.DecodeString(v)
		extIDs = append(extIDs, b)
		//t.Logf("Len %v - %v", i, len(b))
	}
	nbsk := new(NewCoinbaseAddressStruct)
	err := nbsk.DecodeFromExtIDs(extIDs)
	if err != nil {
		t.Errorf("%v", err)
	}
	h := nbsk.GetChainID()
	if h.String() != "d99b5c1ce77b1960c1071cc57dc81a13ee66b25f9e415546db4af520a7ff9c48" {
		t.Errorf("Wrong ChainID, expected d99b5c1ce77b1960c1071cc57dc81a13ee66b25f9e415546db4af520a7ff9c48, got %v", h.String())
	}

	err = nbsk.VerifySignature(nil)
	if err != nil {
		t.Errorf("%v", err)
	}
}
