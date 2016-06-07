// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/testHelper"
	"github.com/FactomProject/factomd/util"
)

var _ = log.Print
var _ = util.ReadConfig

func TestInit(t *testing.T) {
	testHelper.CreateEmptyTestState()
}

func TestDirBlockHead(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	height := state.GetHighestRecordedBlock()
	if height != 9 {
		t.Errorf("Invalid DBLock Height - got %v, expected 10", height+1)
	}
	d := state.GetDirectoryBlockByHeight(height)
	if d.GetKeyMR().String() != "d405071f6e382adeca5954be80fd012758bd0a298a7f4730e87c37efc11094c6" {
		t.Errorf("Invalid DBLock KeyMR - got %v, expected d405071f6e382adeca5954be80fd012758bd0a298a7f4730e87c37efc11094c6", d.GetKeyMR().String())
	}
}

func TestGetDirectoryBlockByHeight(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	blocks := testHelper.CreateFullTestBlockSet()
	for i, block := range blocks {
		dBlock := state.GetDirectoryBlockByHeight(uint32(i))
		if dBlock.GetKeyMR().IsSameAs(block.DBlock.GetKeyMR()) == false {
			t.Errorf("DBlocks are not the same at height %v", i+1)
			continue
		}
		if dBlock.GetFullHash().IsSameAs(block.DBlock.GetFullHash()) == false {
			t.Errorf("DBlocks are not the same at height %v", i+1)
			continue
		}
	}
}
