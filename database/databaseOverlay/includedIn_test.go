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
		err := dbo.SaveEBlockHead(prev)
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
