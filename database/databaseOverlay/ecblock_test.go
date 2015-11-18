// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	. "github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	"testing"
)

func TestSaveLoadECBlockHead(t *testing.T) {
	b1 := createTestEntryCreditBlock(nil)

	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	err := dbo.SaveECBlockHead(b1)
	if err != nil {
		t.Error(err)
	}

	head, err := dbo.FetchECBlockHead()
	if err != nil {
		t.Error(err)
	}
	if head == nil {
		t.Error("ECBlock head is nil")
	}

	m1, err := b1.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	m2, err := head.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if AreBytesEqual(m1, m2) == false {
		t.Error("Blocks are not equal")
	}

	b2 := createTestEntryCreditBlock(b1)

	err = dbo.SaveECBlockHead(b2)
	if err != nil {
		t.Error(err)
	}

	head, err = dbo.FetchECBlockHead()
	if err != nil {
		t.Error(err)
	}
	if head == nil {
		t.Error("ECBlock head is nil")
	}

	m1, err = b2.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	m2, err = head.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
}

func TestSaveLoadECBlockChain(t *testing.T) {
	blocks := []*ECBlock{}
	max := 10
	var prev *ECBlock = nil
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	for i := 0; i < max; i++ {
		prev = createTestEntryCreditBlock(prev)
		blocks = append(blocks, prev)
		err := dbo.SaveECBlockHead(prev)
		if err != nil {
			t.Error(err)
		}
	}

	current, err := dbo.FetchECBlockHead()
	if err != nil {
		t.Error(err)
	}
	zero := primitives.NewZeroHash()
	fetchedCount := 1
	for {
		keyMR := current.(*ECBlock).Header.PrevHeaderHash
		if keyMR.IsSameAs(zero) {
			break
		}
		t.Logf("KeyMR - %v", keyMR.String())

		current, err = dbo.FetchECBlockByKeyMR(keyMR)
		if err != nil {
			t.Error(err)
		}
		if current == nil {
			t.Fatal("Block not found")
		}
		fetchedCount++
	}
	if fetchedCount != max {
		t.Error("Wrong number of entries fetched - %v vs %v", fetchedCount, max)
	}
}

func createTestEntryCreditBlock(prev *ECBlock) *ECBlock {
	block, err := NextECBlock(prev)
	if err != nil {
		panic(err)
	}
	return block
}
