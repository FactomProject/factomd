// +build all 

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	//"github.com/FactomProject/factomd/common/primitives"

	. "github.com/FactomProject/factomd/common/entryBlock"
	. "github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	"github.com/FactomProject/factomd/testHelper"

	"testing"
)

func TestIncludedIn(t *testing.T) {
	blocks := []*EBlock{}
	max := 10
	var prev *EBlock = nil
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	for i := 0; i < max; i++ {
		prev, _ = testHelper.CreateTestEntryBlock(prev)
		blocks = append(blocks, prev)
		err := dbo.SaveEBlockHead(prev, false)
		if err != nil {
			t.Error(err)
		}
	}

	for _, block := range blocks {
		for _, entry := range block.GetEntryHashes() {
			if entry.IsMinuteMarker() {
				continue
			}
			blockHash, err := dbo.FetchIncludedIn(entry)
			if err != nil {
				t.Error(err)
			}
			if blockHash.IsSameAs(block.DatabasePrimaryIndex()) == false {
				t.Error("Wrong IncludedIn result")
			}
		}
	}
}

func TestIncludedInOverwriting(t *testing.T) {
	blocks := []*EBlock{}
	max := 10
	var prev *EBlock = nil
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	for i := 0; i < max; i++ {
		prev, _ = testHelper.CreateTestEntryBlockWithContentN(prev, 1)
		blocks = append(blocks, prev)
		err := dbo.SaveEBlockHead(prev, true)
		if err != nil {
			t.Error(err)
		}
	}

	for i, block := range blocks {
		for _, entry := range block.GetEntryHashes() {
			if entry.IsMinuteMarker() {
				continue
			}
			blockHash, err := dbo.FetchIncludedIn(entry)
			if err != nil {
				t.Error(err)
			}
			if i < 2 {
				if blockHash.IsSameAs(block.DatabasePrimaryIndex()) == false {
					t.Error("Wrong IncludedIn result")
				}
			} else {
				if blockHash.IsSameAs(blocks[1].DatabasePrimaryIndex()) == false {
					t.Error("Wrong IncludedIn result")
				}
			}
		}
	}

	blocks = []*EBlock{}
	prev = nil
	dbo.Close()
	dbo = NewOverlay(new(mapdb.MapDB))

	for i := 0; i < max; i++ {
		prev, _ = testHelper.CreateTestEntryBlockWithContentN(prev, 1)
		blocks = append(blocks, prev)
		err := dbo.SaveEBlockHead(prev, false)
		if err != nil {
			t.Error(err)
		}
	}

	for i, block := range blocks {
		for _, entry := range block.GetEntryHashes() {
			if entry.IsMinuteMarker() {
				continue
			}
			blockHash, err := dbo.FetchIncludedIn(entry)
			if err != nil {
				t.Error(err)
			}
			if i < 1 {
				if blockHash.IsSameAs(block.DatabasePrimaryIndex()) == false {
					t.Error("Wrong IncludedIn result")
				}
			} else {
				if blockHash.IsSameAs(blocks[max-1].DatabasePrimaryIndex()) == false {
					t.Error("Wrong IncludedIn result")
				}
			}
		}
	}
}

func TestIncludedInFromAllBlocks(t *testing.T) {
	dbo := testHelper.CreateAndPopulateTestDatabaseOverlay()

	dBlocks, err := dbo.FetchAllDBlocks()
	if err != nil {
		t.Error(err)
	}

	for _, block := range dBlocks {
		blockHash := block.DatabasePrimaryIndex()
		entries := block.GetEntryHashes()
		for _, entry := range entries {
			in, err := dbo.FetchIncludedIn(entry)
			if err != nil {
				t.Error(err)
			}
			if in.IsSameAs(blockHash) == false {
				t.Errorf("Entry not found in dBlocks - %v vs %v for %v", in.String(), blockHash.String(), entry)
			}
		}
	}

	fBlocks, err := dbo.FetchAllFBlocks()
	if err != nil {
		t.Error(err)
	}

	for _, block := range fBlocks {
		blockHash := block.DatabasePrimaryIndex()
		entries := block.GetEntryHashes()
		for _, entry := range entries {
			in, err := dbo.FetchIncludedIn(entry)
			if err != nil {
				t.Error(err)
			}
			if in.IsSameAs(blockHash) == false {
				t.Errorf("Entry not found in fBlocks - %v vs %v for %v", in.String(), blockHash.String(), entry)
			}
		}
		entries = block.GetEntrySigHashes()
		for _, entry := range entries {
			in, err := dbo.FetchIncludedIn(entry)
			if err != nil {
				t.Error(err)
			}
			if in.IsSameAs(blockHash) == false {
				t.Errorf("Entry not found in fBlocks - %v vs %v for %v", in.String(), blockHash.String(), entry)
			}
		}
	}

	ecBlocks, err := dbo.FetchAllECBlocks()
	if err != nil {
		t.Error(err)
	}

	for _, block := range ecBlocks {
		blockHash := block.DatabasePrimaryIndex()
		entries := block.GetEntryHashes()
		for _, entry := range entries {
			in, err := dbo.FetchIncludedIn(entry)
			if err != nil {
				t.Error(err)
			}
			if in.IsSameAs(blockHash) == false {
				t.Errorf("Entry not found in ecBlocks - %v vs %v for %v", in.String(), blockHash.String(), entry)
			}
		}
		entries = block.GetEntrySigHashes()
		for _, entry := range entries {
			in, err := dbo.FetchIncludedIn(entry)
			if err != nil {
				t.Error(err)
			}
			if in.IsSameAs(blockHash) == false {
				t.Errorf("Entry not found in ecBlocks - %v vs %v for %v", in.String(), blockHash.String(), entry)
			}
		}
	}

	eBlocks, err := dbo.FetchAllEBlocksByChain(testHelper.GetChainID())
	if err != nil {
		t.Error(err)
	}

	for _, block := range eBlocks {
		blockHash := block.DatabasePrimaryIndex()
		entries := block.GetEntryHashes()
		for _, entry := range entries {
			if entry.IsMinuteMarker() {
				continue
			}
			in, err := dbo.FetchIncludedIn(entry)
			if err != nil {
				t.Error(err)
			}
			if in == nil {
				t.Errorf("IncludedIn not found for %v", entry.String())
			}
			if in.IsSameAs(blockHash) == false {
				t.Errorf("Entry not found in eBlocks - %v vs %v for %v", in.String(), blockHash.String(), entry)
			}
		}
	}

	anchorBlocks, err := dbo.FetchAllEBlocksByChain(testHelper.GetAnchorChainID())
	if err != nil {
		t.Error(err)
	}

	for _, block := range anchorBlocks {
		blockHash := block.DatabasePrimaryIndex()
		entries := block.GetEntryHashes()
		for _, entry := range entries {
			if entry.IsMinuteMarker() {
				continue
			}
			in, err := dbo.FetchIncludedIn(entry)
			if err != nil {
				t.Error(err)
			}
			if in == nil {
				t.Errorf("IncludedIn not found for %v", entry.String())
			}
			if in.IsSameAs(blockHash) == false {
				t.Errorf("Entry not found in anchorBlocks - %v vs %v for %v", in.String(), blockHash.String(), entry)
			}
		}
	}
}
