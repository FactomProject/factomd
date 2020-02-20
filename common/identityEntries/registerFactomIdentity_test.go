// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identityEntries_test

import (
	"encoding/hex"
	"testing"

	"math/rand"
	"time"

	"bytes"

	. "github.com/FactomProject/factomd/common/identityEntries"
)

func TestRegisterFactomIdentityStructure(t *testing.T) {
	parts := []string{
		"00",
		"526567697374657220466163746F6D204964656E74697479",
		"888888d027c59579fc47a6fc6c4a5c0409c7c39bc38a86cb5fc0069978493762",
		"0125b0e7fd5e68b4dec40ca0cd2db66be84c02fe6404b696c396e3909079820f61",
		"764974ae61de0d57507b80da61a809382e699cf0e31be44a5d357bd6c93d12fa6746b29c80f7184bd3c715eb910035d4dac2d8ecb1c4b731692e68631c69a503",
	}

	extIDs := [][]byte{}
	for _, v := range parts {
		b, _ := hex.DecodeString(v)
		extIDs = append(extIDs, b)
		//t.Logf("Len %v - %v", i, len(b))
	}
	rfi := new(RegisterFactomIdentityStructure)
	err := rfi.DecodeFromExtIDs(extIDs)
	if err != nil {
		t.Errorf("%v", err)
	}
	h := rfi.GetChainID()
	if h.String() != "9d40f0a4250a19c7d2cafc3bdbec5711edf840e53340748b4f746b8544700293" {
		t.Errorf("Wrong ChainID, expected 9d40f0a4250a19c7d2cafc3bdbec5711edf840e53340748b4f746b8544700293, got %v", h.String())
	}

	err = rfi.VerifySignature(nil)
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestRegisterFactomIdentityStructureMarshal(t *testing.T) {
	for i := 0; i < 100; i++ {
		rand.Seed(time.Now().UnixNano())
		r := RandomRegisterFactomIdentityStructure()
		data, err := r.MarshalBinary()
		if err != nil {
			t.Error(err)
		}

		r2 := new(RegisterFactomIdentityStructure)
		nd, err := r2.UnmarshalBinaryData(data)
		if err != nil {
			t.Error(err)
		}

		if len(nd) != 0 {
			t.Errorf("left over %d bytes", len(nd))
		}

		if !r.IsSameAs(r2) {
			t.Errorf("Not same")
		}

		data2, err := r2.MarshalBinary()
		if err != nil {
			t.Error(err)
		}

		if bytes.Compare(data, data2) != 0 {
			t.Errorf("Bytes different")
		}
	}
}
