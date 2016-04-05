// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/testHelper"
	"github.com/FactomProject/factomd/util"
	"testing"
)

var _ = log.Print
var _ = util.ReadConfig

func TestInit(t *testing.T) {
	testHelper.CreateEmptyTestState()
}

func TestDirBlockHead(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	height := state.GetHighestRecordedBlock()
	if height != 10 {
		t.Errorf("Invalid DBLock Height - got %v, expected 10", height)
	}
	d := state.GetDirectoryBlockByHeight(height)
	if d.GetKeyMR().String() != "93b9d8bc11869819aed5e11ff15c865435a58d7b57c9f27fe4638dfc23f13b34" {
		t.Errorf("Invalid DBLock KeyMR - got %v, expected 93b9d8bc11869819aed5e11ff15c865435a58d7b57c9f27fe4638dfc23f13b34", d.GetKeyMR().String())
	}
}
