// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	. "github.com/FactomProject/factomd/common/factoid/block"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	"testing"
)

func TestSaveLoadFBlockHead(t *testing.T) {
	b1 := createTestFactoidBlock(nil)

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

	b2 := createTestFactoidBlock(b1)

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
}

func TestSaveLoadFBlockChain(t *testing.T) {
	blocks := []interfaces.IFBlock{}
	max := 10
	var prev interfaces.IFBlock = nil
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	for i := 0; i < max; i++ {
		prev = createTestFactoidBlock(prev)
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

		current, err = dbo.FetchFBlockByKeyMR(keyMR)
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

func createTestFactoidBlock(prev interfaces.IFBlock) interfaces.IFBlock {
	return NewFBlockFromPreviousBlock(1, prev)
}
