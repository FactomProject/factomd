// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker_test

import (
	"testing"

	. "github.com/FactomProject/factomd/blockchainState/blockMaker"
	"github.com/FactomProject/factomd/testHelper"
)

func TestBuildEBlocks(t *testing.T) {
	eb, es := testHelper.CreateTestEntryBlock(nil)
	eb.GetHeader().SetDBHeight(1)
	bm := NewBlockMaker()
	err := bm.ProcessEBEntry(es[0])
	if err != nil {
		t.Errorf("%v", err)
	}
	ebs, err := bm.BuildEBlocks()
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(ebs) != 1 {
		t.Errorf("Invalid amount of EBlocks created - %v", len(ebs))
	}
	if ebs[0].GetHash().IsSameAs(eb.GetHash()) == false {
		t.Errorf("Wrong block hash - %v vs %v", ebs[0].GetHash(), eb.GetHash())
		s, _ := ebs[0].JSONString()
		t.Logf("%v", s)
		s, _ = eb.JSONString()
		t.Logf("%v", s)
	}
}
