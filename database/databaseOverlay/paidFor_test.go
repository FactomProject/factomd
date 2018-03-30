// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/testHelper"

	"testing"
)

func TestPaidFor(t *testing.T) {
	blocks := CreateFullTestBlockSet()
	dbo := CreateEmptyTestDatabaseOverlay()

	for i := range blocks {
		err := dbo.SavePaidForMultiFromBlock(blocks[i].ECBlock, false)
		if err != nil {
			t.Error(err)
		}
	}

	for _, block := range blocks {
		ecEntries := block.ECBlock.GetBody().GetEntries()
		for _, entry := range ecEntries {
			if entry.ECID() != constants.ECIDChainCommit && entry.ECID() != constants.ECIDEntryCommit {
				continue
			}
			var entryHash interfaces.IHash

			if entry.ECID() == constants.ECIDChainCommit {
				entryHash = entry.(*entryCreditBlock.CommitChain).EntryHash
			}
			if entry.ECID() == constants.ECIDEntryCommit {
				entryHash = entry.(*entryCreditBlock.CommitEntry).EntryHash
			}

			h, err := dbo.FetchPaidFor(entryHash)
			if err != nil {
				t.Error(err)
			}
			if h == nil {
				t.Error("PaidFor not found")
			}

			if h.IsSameAs(entry.Hash()) == false {
				t.Error("PaidFor answer does not match")
			}
		}

		for _, entry := range block.Entries {
			h, err := dbo.FetchPaidFor(entry.GetHash())
			if err != nil {
				t.Error(err)
			}
			if h == nil {
				t.Error("PaidFor not found for entry")
			}
		}
	}
}
