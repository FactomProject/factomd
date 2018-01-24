// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/state"
)

func TestGenerateGenesisBlocks(t *testing.T) {
	d, a, f, ec := GenerateGenesisBlocks(constants.MAIN_NETWORK_ID, nil)

	if a.DatabasePrimaryIndex().String() != "4fb409d5369fad6aa7768dc620f11cd219f9b885956b631ad050962ca934052e" {
		t.Errorf("Invalid ABlock")
	}
	if ec.DatabasePrimaryIndex().String() != "f87cfc073df0e82cdc2ed0bb992d7ea956fd32b435b099fc35f4b0696948507a" {
		t.Errorf("Invalid ECBlock")
	}
	if f.DatabasePrimaryIndex().String() != "a164ccbb77a21904edc4f2bb753aa60635fb2b60279c06ae01aa211f37541736" {
		t.Errorf("Invalid FBlock")
	}
	if d.DatabasePrimaryIndex().String() != "64d4352b134280305599363ea388c2a9c3c64dc3ee6e0100893262e372bf064b" {
		t.Errorf("Invalid DBlock")
	}
}
