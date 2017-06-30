// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker_test

import (
	"testing"

	. "github.com/FactomProject/factomd/blockchainState/blockMaker"
	"github.com/FactomProject/factomd/testHelper"
)

func TestBuildABlock(t *testing.T) {
	bSet := testHelper.CreateTestBlockSet(nil)
	ab := bSet.ABlock
	ab.GetHeader().SetDBHeight(1)

	bm := NewBlockMaker()
	abe := ab.GetABEntries()
	for _, v := range abe {
		err := bm.ProcessABEntry(v)
		if err != nil {
			t.Errorf("%v", err)
		}
	}

	ab2, err := bm.BuildABlockWithHeaderExpansionArea(ab.GetHeader().GetHeaderExpansionArea())
	if err != nil {
		t.Errorf("%v", err)
	}

	if ab.GetHash().IsSameAs(ab2.GetHash()) == false {
		t.Errorf("Wrong block hash - %v vs %v", ab.GetHash(), ab2.GetHash())
		s, _ := ab.JSONString()
		t.Logf("%v", s)
		s, _ = ab2.JSONString()
		t.Logf("%v", s)
	}
}
