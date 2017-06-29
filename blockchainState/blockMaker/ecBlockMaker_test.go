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

func TestBuildECBlocks(t *testing.T) {
	bSet := testHelper.CreateTestBlockSet(nil)
	ecb := bSet.ECBlock
	ecb.GetHeader().SetDBHeight(1)

	bm := NewBlockMaker()
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

	ecb2, err := bm.BuildECBlock()
	if err != nil {
		t.Errorf("%v", err)
	}

	if ecb.GetHash().IsSameAs(ecb2.GetHash()) == false {
		t.Errorf("Wrong block hash - %v vs %v", ecb.GetHash(), ecb2.GetHash())
		s, _ := ecb.JSONString()
		t.Logf("%v", s)
		s, _ = ecb2.JSONString()
		t.Logf("%v", s)
	}
}
