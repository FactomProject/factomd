// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker_test

import (
	"testing"

	. "github.com/FactomProject/factomd/blockchainState/blockMaker"
	"github.com/FactomProject/factomd/testHelper"
)

func TestBuildFBlock(t *testing.T) {
	bSet := testHelper.CreateTestBlockSet(nil)
	fb := bSet.FBlock
	fb.SetDBHeight(1)

	bm := NewBlockMaker()
	bm.BState.ExchangeRate = fb.GetExchRate()
	fts := fb.GetTransactions()
	for _, v := range fts {
		err := bm.ProcessFactoidTransaction(v)
		if err != nil {
			t.Errorf("%v", err)
		}
	}

	fb2, err := bm.BuildFBlock()
	if err != nil {
		t.Errorf("%v", err)
	}

	if fb.GetHash().IsSameAs(fb2.GetHash()) == false {
		t.Errorf("Wrong block hash - %v vs %v", fb.GetHash(), fb2.GetHash())
		s, _ := fb.JSONString()
		t.Logf("%v", s)
		s, _ = fb2.JSONString()
		t.Logf("%v", s)
	}
}
