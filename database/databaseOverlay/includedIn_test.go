// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	/*
		"github.com/FactomProject/factomd/common/primitives"*/

	. "github.com/FactomProject/factomd/common/entryBlock"
	. "github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
)

func TestIncludedIn(t *testing.T) {
	blocks := []*EBlock{}
	max := 10
	var prev *EBlock = nil
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	for i := 0; i < max; i++ {
		prev, _ = CreateTestEntryBlock(prev)
		blocks = append(blocks, prev)
		err := dbo.SaveEBlockHead(prev, false)
		if err != nil {
			t.Error(err)
		}
	}

	for _, block := range blocks {
		for _, entry := range block.GetEntryHashes() {
			blockHash, err := dbo.LoadIncludedIn(entry)
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
		prev, _ = CreateTestEntryBlockWithContentN(prev, 1)
		blocks = append(blocks, prev)
		err := dbo.SaveEBlockHead(prev, true)
		if err != nil {
			t.Error(err)
		}
	}

	for i, block := range blocks {
		for _, entry := range block.GetEntryHashes() {
			blockHash, err := dbo.LoadIncludedIn(entry)
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
		prev, _ = CreateTestEntryBlockWithContentN(prev, 1)
		blocks = append(blocks, prev)
		err := dbo.SaveEBlockHead(prev, false)
		if err != nil {
			t.Error(err)
		}
	}

	for i, block := range blocks {
		for _, entry := range block.GetEntryHashes() {
			blockHash, err := dbo.LoadIncludedIn(entry)
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
