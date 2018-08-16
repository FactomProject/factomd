// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	"github.com/FactomProject/factomd/testHelper"
)

func TestSaveLoadDBlockHead(t *testing.T) {
	blocks := testHelper.CreateTestBlockSet(nil)
	b1 := blocks.DBlock

	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	err := dbo.SaveDirectoryBlockHead(b1)
	if err != nil {
		t.Error(err)
	}

	head, err := dbo.FetchDirectoryBlockHead()
	if err != nil {
		t.Error(err)
	}
	if head == nil {
		t.Error("DBlock head is nil")
	}

	m1, err := b1.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	m2, err := head.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if primitives.AreBytesEqual(m1, m2) == false {
		t.Error("Blocks are not equal")
	}

	blocks = testHelper.CreateTestBlockSet(blocks)
	b2 := blocks.DBlock

	err = dbo.SaveDirectoryBlockHead(b2)
	if err != nil {
		t.Error(err)
	}

	head, err = dbo.FetchDirectoryBlockHead()
	if err != nil {
		t.Error(err)
	}
	if head == nil {
		t.Error("DBlock head is nil")
	}

	m1, err = b2.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	m2, err = head.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if primitives.AreBytesEqual(m1, m2) == false {
		t.Error("Blocks are not equal")
	}
}

func TestSaveLoadDBlockChain(t *testing.T) {
	blocks := []*DirectoryBlock{}
	max := 10
	var prev *testHelper.BlockSet = nil
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	for i := 0; i < max; i++ {
		prev = testHelper.CreateTestBlockSet(prev)
		blocks = append(blocks, prev.DBlock)
		err := dbo.SaveDirectoryBlockHead(prev.DBlock)
		if err != nil {
			t.Error(err)
		}
	}

	current, err := dbo.FetchDirectoryBlockHead()
	if err != nil {
		t.Error(err)
	}
	zero := primitives.NewZeroHash()
	fetchedCount := 1
	for {
		keyMR := current.GetHeader().GetPrevKeyMR()
		if keyMR.IsSameAs(zero) {
			break
		}
		t.Logf("KeyMR - %v", keyMR.String())
		hash := current.GetHeader().GetPrevFullHash()

		current, err = dbo.FetchDBlockByPrimary(keyMR)
		if err != nil {
			t.Error(err)
		}
		if current == nil {
			t.Fatal("Block not found")
		}
		fetchedCount++

		byHash, err := dbo.FetchDBlockBySecondary(hash)

		same, err := primitives.AreBinaryMarshallablesEqual(current, byHash)
		if err != nil {
			t.Error(err)
		}
		if same == false {
			t.Error("Blocks fetched by keyMR and hash are not identical")
		}
	}
	if fetchedCount != max {
		t.Errorf("Wrong number of entries fetched - %v vs %v", fetchedCount, max)
	}

	all, err := dbo.FetchAllDBlocks()
	if err != nil {
		t.Error(err)
	}
	if len(all) != max {
		t.Errorf("Wrong number of entries fetched - %v vs %v", len(all), max)
	}
	for i := range all {
		same, err := primitives.AreBinaryMarshallablesEqual(blocks[i], all[i])
		if err != nil {
			t.Error(err)
		}
		if same == false {
			t.Error("Blocks fetched by all and original blocks are not identical")
			t.Logf("\n%v\nvs\n%v", blocks[i].String(), all[i].String())
		}
	}
}

func TestLoadUnknownDBlocks(t *testing.T) {
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()
	for i := 0; i < 10; i++ {
		b := testHelper.IntToByteSlice(i)
		hash, err := primitives.NewShaHash(b)
		if err != nil {
			t.Error(err)
		}
		data, err := dbo.FetchDBlockBySecondary(hash)
		if err != nil {
			t.Error(err)
		}
		if data != nil {
			t.Errorf("Fetched entry while we expected nil - %v", data)
		}
		data, err = dbo.FetchDBlockByPrimary(hash)
		if err != nil {
			t.Error(err)
		}
		if data != nil {
			t.Errorf("Fetched entry while we expected nil - %v", data)
		}
		data, err = dbo.FetchDirectoryBlockHead()
		if err != nil {
			t.Error(err)
		}
		if data != nil {
			t.Errorf("Fetched entry while we expected nil - %v", data)
		}
		all, err := dbo.FetchAllDBlocks()
		if err != nil {
			t.Error(err)
		}
		if len(all) != 0 {
			t.Errorf("Fetched entries while we expected nil - %v", all)
		}
	}
}

func TestChainHeightsAndHeads(t *testing.T) {
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	max := 10
	var prev *testHelper.BlockSet = nil
	blocks := []*testHelper.BlockSet{}

	for i := 0; i < max; i++ {
		prev = testHelper.CreateTestBlockSet(prev)
		blocks = append(blocks, prev)
		err := dbo.SaveDirectoryBlockHead(prev.DBlock)
		if err != nil {
			t.Error(err)
		}
		err = dbo.SaveABlock(prev.ABlock)
		if err != nil {
			t.Error(err)
		}
		err = dbo.SaveECBlock(prev.ECBlock, false)
		if err != nil {
			t.Error(err)
		}
		err = dbo.SaveFBlock(prev.FBlock)
		if err != nil {
			t.Error(err)
		}
		err = dbo.SaveEBlock(prev.EBlock, false)
		if err != nil {
			t.Error(err)
		}
		err = dbo.SaveEBlock(prev.AnchorEBlock, false)
		if err != nil {
			t.Error(err)
		}

		//check block heads

		dBlock, err := dbo.FetchDBlockHead()
		if err != nil {
			t.Error(err)
		}
		ok, err := primitives.AreBinaryMarshallablesEqual(prev.DBlock, dBlock)
		if err != nil {
			t.Error(err)
		}
		if ok == false {
			t.Errorf("DBlock heads are not equal")
		}

		aBlock, err := dbo.FetchABlockHead()
		if err != nil {
			t.Error(err)
		}
		ok, err = primitives.AreBinaryMarshallablesEqual(prev.ABlock, aBlock)
		if err != nil {
			t.Error(err)
		}
		if ok == false {
			t.Errorf("ABlock heads are not equal")
		}

		ecBlock, err := dbo.FetchECBlockHead()
		if err != nil {
			t.Error(err)
		}
		ok, err = primitives.AreBinaryMarshallablesEqual(prev.ECBlock, ecBlock)
		if err != nil {
			t.Error(err)
		}
		if ok == false {
			t.Errorf("ECBlock heads are not equal")
		}

		fBlock, err := dbo.FetchFBlockHead()
		if err != nil {
			t.Error(err)
		}
		ok, err = primitives.AreBinaryMarshallablesEqual(prev.FBlock, fBlock)
		if err != nil {
			t.Error(err)
		}
		if ok == false {
			t.Errorf("FBlock heads are not equal")
		}

		eBlock, err := dbo.FetchEBlockHead(prev.EBlock.GetChainID())
		if err != nil {
			t.Error(err)
		}
		ok, err = primitives.AreBinaryMarshallablesEqual(prev.EBlock, eBlock)
		if err != nil {
			t.Error(err)
		}
		if ok == false {
			t.Errorf("EBlock heads are not equal")
		}

		eBlock, err = dbo.FetchEBlockHead(prev.AnchorEBlock.GetChainID())
		if err != nil {
			t.Error(err)
		}
		ok, err = primitives.AreBinaryMarshallablesEqual(prev.AnchorEBlock, eBlock)
		if err != nil {
			t.Error(err)
		}
		if ok == false {
			t.Errorf("Anchor EBlock heads are not equal")
		}
	}

	allEblocks, err := dbo.FetchAllEBlocksByChain(prev.EBlock.GetChainID())
	if err != nil {
		t.Error(err)
	}
	if len(allEblocks) != max {
		t.Errorf("Wrong number of entries fetched - %v vs %v", len(allEblocks), max)
	}
	allAnchorEblocks, err := dbo.FetchAllEBlocksByChain(prev.AnchorEBlock.GetChainID())
	if err != nil {
		t.Error(err)
	}
	if len(allAnchorEblocks) != max {
		t.Errorf("Wrong number of entries fetched - %v vs %v", len(allAnchorEblocks), max)
	}

	//check block heights
	for i := 0; i < max; i++ {
		prev = blocks[i]

		dBlock, err := dbo.FetchDBlockByHeight(prev.DBlock.GetDatabaseHeight())
		if err != nil {
			t.Error(err)
		}
		ok, err := primitives.AreBinaryMarshallablesEqual(prev.DBlock, dBlock)
		if err != nil {
			t.Error(err)
		}
		if ok == false {
			t.Errorf("DBlocks by height are not equal")
		}

		aBlock, err := dbo.FetchABlockByHeight(prev.DBlock.GetDatabaseHeight())
		if err != nil {
			t.Error(err)
		}
		ok, err = primitives.AreBinaryMarshallablesEqual(prev.ABlock, aBlock)
		if err != nil {
			t.Error(err)
		}
		if ok == false {
			t.Errorf("ABlocks by height are not equal")
		}

		ecBlock, err := dbo.FetchECBlockByHeight(prev.DBlock.GetDatabaseHeight())
		if err != nil {
			t.Error(err)
		}
		ok, err = primitives.AreBinaryMarshallablesEqual(prev.ECBlock, ecBlock)
		if err != nil {
			t.Error(err)
		}
		if ok == false {
			t.Errorf("ECBlocks by height are not equal")
		}

		fBlock, err := dbo.FetchFBlockByHeight(prev.DBlock.GetDatabaseHeight())
		if err != nil {
			t.Error(err)
		}
		ok, err = primitives.AreBinaryMarshallablesEqual(prev.FBlock, fBlock)
		if err != nil {
			t.Error(err)
		}
		if ok == false {
			t.Errorf("FBlocks by height are not equal")
		}

		ok, err = primitives.AreBinaryMarshallablesEqual(prev.EBlock, allEblocks[i])
		if err != nil {
			t.Error(err)
		}
		if ok == false {
			t.Error("Blocks fetched by all and original blocks are not identical")
			t.Logf("\n%v\nvs\n%v", allEblocks[i].String(), prev.EBlock.String())
		}

		ok, err = primitives.AreBinaryMarshallablesEqual(prev.AnchorEBlock, allAnchorEblocks[i])
		if err != nil {
			t.Error(err)
		}
		if ok == false {
			t.Error("Blocks fetched by all and original blocks are not identical")
			t.Logf("\n%v\nvs\n%v", allAnchorEblocks[i].String(), prev.AnchorEBlock.String())
		}
	}

	//check fetch all blocks from chain
}
