// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"
	//"time"

	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/testHelper"
	//. "github.com/FactomProject/factomd/state"
)

func TestIsStateFullySynced(t *testing.T) {
	s1_good := CreateAndPopulateTestState()
	t.Log("IsStateFullySynced():", s1_good.IsStateFullySynced())
	if !s1_good.IsStateFullySynced() {
		t.Error("test state is show to be not fully synced")
	}

	//// we can't test the negative here because when we set the bad DBHeight the
	//// state.ValidatorLoop() will panic before we call IsStateFullySynced()
	//	s2_bad := CreateAndPopulateTestState()
	//// the next line will cause the ValidatorLoop to panic
	//	s2_bad.ProcessLists.DBHeightBase = s2_bad.ProcessLists.LastList().DBHeight+10
	//	t.Log("IsStateFullySynced:", s2_bad.IsStateFullySynced())

}

func TestFetchECTransactionByHash(t *testing.T) {
	s1 := CreateAndPopulateTestState()
	blocks := CreateFullTestBlockSet()

	for _, block := range blocks {
		for _, tx := range block.ECBlock.GetEntries() {
			if tx.ECID() != entryCreditBlock.ECIDChainCommit &&
				tx.ECID() != entryCreditBlock.ECIDEntryCommit ||
				tx.ECID() == entryCreditBlock.ECIDBalanceIncrease {
				continue
			}

			dtx, err := s1.FetchECTransactionByHash(tx.Hash())
			if err != nil {
				t.Error("Could not fetch transaction:", err)
			}
			if dtx == nil {
				t.Error("transaction not found in database")
				continue
			}

			p1, err := tx.MarshalBinary()
			if err != nil {
				t.Error(err)
			}
			p2, err := tx.MarshalBinary()
			if err != nil {
				t.Error(err)
			}
			if !primitives.AreBytesEqual(p1, p2) {
				t.Error("database transaction does not match transaction")
			}
		}
	}
}
