// +build all

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	"github.com/FactomProject/factomd/testHelper"
)

func TestSaveLoadFBlockHead(t *testing.T) {
	b1 := testHelper.CreateTestFactoidBlock(nil)

	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	err := dbo.SaveFactoidBlockHead(b1)
	if err != nil {
		t.Error(err)
	}

	head, err := dbo.FetchFactoidBlockHead()
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

	b2 := testHelper.CreateTestFactoidBlock(b1)

	err = dbo.SaveFactoidBlockHead(b2)
	if err != nil {
		t.Error(err)
	}

	head, err = dbo.FetchFactoidBlockHead()
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

func TestSaveLoadFBlockChain(t *testing.T) {
	blocks := []interfaces.IFBlock{}
	max := 10
	var prev interfaces.IFBlock = nil
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	for i := 0; i < max; i++ {
		prev = testHelper.CreateTestFactoidBlock(prev)
		blocks = append(blocks, prev)
		err := dbo.SaveFactoidBlockHead(prev)
		if err != nil {
			t.Error(err)
		}
	}

	current, err := dbo.FetchFactoidBlockHead()
	if err != nil {
		t.Error(err)
	}
	zero := primitives.NewZeroHash()
	fetchedCount := 1
	for {
		keyMR := current.GetPrevKeyMR()
		if keyMR.IsSameAs(zero) {
			break
		}
		t.Logf("KeyMR - %v", keyMR.String())

		current, err = dbo.FetchFBlockByPrimary(keyMR)
		if err != nil {
			t.Error(err)
		}
		if current == nil {
			t.Fatal("Block not found")
		}
		fetchedCount++
		hash := current.GetHash()

		byHash, err := dbo.FetchFBlockBySecondary(hash)

		same, err := primitives.AreBinaryMarshallablesEqual(current, byHash)
		if err != nil {
			t.Error(err)
		}
		if same == false {
			one, _ := current.JSONString()
			two, _ := byHash.JSONString()
			t.Errorf("Blocks fetched by keyMR and hash are not identical\n%v\nvs\n%v", one, two)
		}
	}
	if fetchedCount != max {
		t.Errorf("Wrong number of entries fetched - %v vs %v", fetchedCount, max)
	}

	all, err := dbo.FetchAllFBlocks()
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

func TestLoadUnknownFBlocks(t *testing.T) {
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()
	for i := 0; i < 10; i++ {
		b := testHelper.IntToByteSlice(i)
		hash, err := primitives.NewShaHash(b)
		if err != nil {
			t.Error(err)
		}
		data, err := dbo.FetchFBlockByPrimary(hash)
		if err != nil {
			t.Error(err)
		}
		if data != nil {
			t.Errorf("Fetched entry while we expected nil - %v", data)
		}
		data, err = dbo.FetchFBlockBySecondary(hash)
		if err != nil {
			t.Error(err)
		}
		if data != nil {
			t.Errorf("Fetched entry while we expected nil - %v", data)
		}
		data, err = dbo.FetchFactoidBlockHead()
		if err != nil {
			t.Error(err)
		}
		if data != nil {
			t.Errorf("Fetched entry while we expected nil - %v", data)
		}
		all, err := dbo.FetchAllFBlocks()
		if err != nil {
			t.Error(err)
		}
		if len(all) != 0 {
			t.Errorf("Fetched entries while we expected nil - %v", all)
		}
	}
}
