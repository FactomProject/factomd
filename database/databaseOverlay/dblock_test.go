// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	. "github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
)

func TestSaveLoadDBlockHead(t *testing.T) {
	b1 := CreateTestDirectoryBlock(nil)

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

	b2 := CreateTestDirectoryBlock(b1)

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
	var prev *DirectoryBlock = nil
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	for i := 0; i < max; i++ {
		prev = CreateTestDirectoryBlock(prev)
		blocks = append(blocks, prev)
		err := dbo.SaveDirectoryBlockHead(prev)
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
		t.Error("Wrong number of entries fetched - %v vs %v", fetchedCount, max)
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
		b := IntToByteSlice(i)
		hash, err := primitives.NewShaHash(b)
		if err != nil {
			t.Error(err)
		}
		data, err := dbo.FetchDBlockBySecondary(hash)
		if err != nil {
			t.Error(err)
		}
		if data != nil {
			t.Error("Fetched entry while we expected nil - %v", data)
		}
		data, err = dbo.FetchDBlockByPrimary(hash)
		if err != nil {
			t.Error(err)
		}
		if data != nil {
			t.Error("Fetched entry while we expected nil - %v", data)
		}
		data, err = dbo.FetchDirectoryBlockHead()
		if err != nil {
			t.Error(err)
		}
		if data != nil {
			t.Error("Fetched entry while we expected nil - %v", data)
		}
		all, err := dbo.FetchAllDBlocks()
		if err != nil {
			t.Error(err)
		}
		if len(all) != 0 {
			t.Error("Fetched entries while we expected nil - %v", all)
		}
	}
}
