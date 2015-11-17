// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	. "github.com/FactomProject/factomd/common/entryBlock"
	//"github.com/FactomProject/factomd/common/interfaces"
	"encoding/hex"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	"testing"
)

func TestSaveLoadEBlockHead(t *testing.T) {
	b1 := createTestEntryBlock(nil)

	chain, err := primitives.NewShaHash(b1.GetChainID())
	if err != nil {
		t.Error(err)
	}

	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	err = dbo.SaveEBlockHead(b1)
	if err != nil {
		t.Error(err)
	}

	head, err := dbo.FetchEBlockHead(chain)
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

	if AreBytesEqual(m1, m2) == false {
		t.Error("Blocks are not equal")
	}

	b2 := createTestEntryBlock(b1)

	err = dbo.SaveEBlockHead(b2)
	if err != nil {
		t.Error(err)
	}

	head, err = dbo.FetchEBlockHead(chain)
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

func TestSaveLoadEBlockChain(t *testing.T) {
	blocks := []*EBlock{}
	max := 10
	var prev *EBlock = nil
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	for i := 0; i < max; i++ {
		prev = createTestEntryBlock(prev)
		blocks = append(blocks, prev)
		err := dbo.SaveEBlockHead(prev)
		if err != nil {
			t.Error(err)
		}
	}

	chain, err := primitives.NewShaHash(prev.GetChainID())
	if err != nil {
		t.Error(err)
	}

	current, err := dbo.FetchEBlockHead(chain)
	if err != nil {
		t.Error(err)
	}
	zero := primitives.NewZeroHash()
	fetchedCount := 1
	for {
		keyMR := current.(*EBlock).GetHeader().GetPrevKeyMR()
		if keyMR.IsSameAs(zero) {
			break
		}
		t.Logf("KeyMR - %v", keyMR.String())

		current, err = dbo.FetchEBlockByKeyMR(keyMR)
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

func createTestEntryBlock(prev *EBlock) *EBlock {
	e := NewEBlock()
	entryStr := "4bf71c177e71504032ab84023d8afc16e302de970e6be110dac20adbf9a1974625f25d9375533b44505964af993212ef7c13314736b2c76a37c73571d89d8b21c6180f7430677d46d93a3e17b68e6a25dc89ecc092cee1459101578859f7f6969d171a092a1d04f067d55628b461c6a106b76b4bc860445f87b0052cdc5f2bfd000002d800001b080000000272d72e71fdee4984ecb30eedcc89cb171d1f5f02bf9a8f10a8b2cfbaf03efe1c0000000000000000000000000000000000000000000000000000000000000001"
	h, err := hex.DecodeString(entryStr)
	if err != nil {
		panic(err)
	}
	err = e.UnmarshalBinary(h)
	if err != nil {
		panic(err)
	}

	if prev != nil {
		keyMR, err := prev.KeyMR()
		if err != nil {
			panic(err)
		}

		e.Header.PrevKeyMR = keyMR
		e.Header.DBHeight = prev.Header.DBHeight + 1
	} else {
		e.Header.PrevKeyMR = primitives.NewZeroHash()
		e.Header.DBHeight = 0
	}
	return e
}
