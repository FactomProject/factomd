// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockMaker_test

import (
	"testing"

	. "github.com/FactomProject/factomd/blockchainState/blockMaker"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/testHelper"
)

func TestBuildBlockSet(t *testing.T) {
	bSet := testHelper.CreateTestBlockSet(nil)
	fb := bSet.FBlock
	fb.SetDBHeight(1)
	ecb := bSet.ECBlock
	ecb.GetHeader().SetDBHeight(1)
	ab := bSet.ABlock
	ab.GetHeader().SetDBHeight(1)

	bm := NewBlockMaker()
	bm.BState.ExchangeRate = fb.GetExchRate()

	fts := fb.GetTransactions()
	for _, v := range fts {
		err := bm.ProcessFactoidTransaction(v)
		if err != nil {
			t.Errorf("%v", err)
		}
	}

	bm.SetCurrentMinute(0)

	ecEntries := ecb.GetBody().GetEntries()
	for _, v := range ecEntries {
		if v.ECID() == entryCreditBlock.ECIDMinuteNumber {
			bm.SetCurrentMinute(int(v.(*entryCreditBlock.MinuteNumber).Number))
		} else {
			if v.ECID() == entryCreditBlock.ECIDChainCommit {
				bm.BState.ECBalances[v.(*entryCreditBlock.CommitChain).ECPubKey.String()] = 1000
			}
			if v.ECID() == entryCreditBlock.ECIDEntryCommit {
				bm.BState.ECBalances[v.(*entryCreditBlock.CommitEntry).ECPubKey.String()] = 1000
			}
			err := bm.ProcessECEntry(v)
			if err != nil {
				t.Errorf("%v", err)
			}
		}
	}

	bm.SetCurrentMinute(0)

	abe := ab.GetABEntries()
	for _, v := range abe {
		err := bm.ProcessABEntry(v)
		if err != nil {
			t.Errorf("%v", err)
		}
	}
	bm.SetABlockHeaderExpansionArea(ab.GetHeader().GetHeaderExpansionArea())

	bm.SetCurrentMinute(0)

	eb, es := testHelper.CreateTestEntryBlock(nil)
	eb.GetHeader().SetDBHeight(1)
	bm.BState.PushCommit(es[0].GetHash(), es[0].GetHash())
	err := bm.ProcessEBEntry(es[0])
	if err != nil {
		t.Errorf("%v", err)
	}

	bSet2, err := bm.BuildBlockSet()
	if err != nil {
		t.Errorf("%v", err)
	}

	ebs := bSet2.EBlocks
	fb2 := bSet2.FBlock
	ecb2 := bSet2.ECBlock
	ab2 := bSet2.ABlock

	if len(ebs) != 1 {
		t.Errorf("Invalid amount of EBlocks created - %v", len(ebs))
		t.FailNow()
	}
	if ebs[0].GetHash().IsSameAs(eb.GetHash()) == false {
		t.Errorf("Wrong block hash - %v vs %v", ebs[0].GetHash(), eb.GetHash())
		s, _ := ebs[0].JSONString()
		t.Logf("%v", s)
		s, _ = eb.JSONString()
		t.Logf("%v", s)
	}

	if fb.GetHash().IsSameAs(fb2.GetHash()) == false {
		t.Errorf("Wrong block hash - %v vs %v", fb.GetHash(), fb2.GetHash())
		s, _ := fb.JSONString()
		t.Logf("%v", s)
		s, _ = fb2.JSONString()
		t.Logf("%v", s)
	}
	if ecb.GetHash().IsSameAs(ecb2.GetHash()) == false {
		t.Errorf("Wrong block hash - %v vs %v", ecb.GetHash(), ecb2.GetHash())
		s, _ := ecb.JSONString()
		t.Logf("%v", s)
		s, _ = ecb2.JSONString()
		t.Logf("%v", s)
	}
	if ab.GetHash().IsSameAs(ab2.GetHash()) == false {
		t.Errorf("Wrong block hash - %v vs %v", ab.GetHash(), ab2.GetHash())
		s, _ := ab.JSONString()
		t.Logf("%v", s)
		s, _ = ab2.JSONString()
		t.Logf("%v", s)
	}
}
