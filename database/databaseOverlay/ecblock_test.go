// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
)

func TestSaveLoadECBlockHead(t *testing.T) {
	b1 := CreateTestEntryCreditBlock(nil)

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

	if primitives.AreBytesEqual(m1, m2) == false {
		t.Error("Blocks are not equal")
	}

	b2 := CreateTestEntryCreditBlock(b1)

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
	blocks := []interfaces.IEntryCreditBlock{}
	max := 10
	var prev interfaces.IEntryCreditBlock = nil
	dbo := NewOverlay(new(mapdb.MapDB))
	defer dbo.Close()

	for i := 0; i < max; i++ {
		prev = CreateTestEntryCreditBlock(prev)
		err := dbo.SaveECBlockHead(prev)
		if err != nil {
			t.Error(err)
		}
		blocks = append(blocks, prev)
	}

	current, err := dbo.FetchECBlockHead()
	if err != nil {
		t.Error(err)
	}
	zero := primitives.NewZeroHash()
	fetchedCount := 1
	for {
		keyMR := current.GetHeader().GetPrevHeaderHash()
		if keyMR.IsSameAs(zero) {
			break
		}
		//t.Logf("KeyMR - %v", keyMR.String())
		hash := current.GetHeader().GetPrevLedgerKeyMR()

		current, err = dbo.FetchECBlockByKeyMR(keyMR)
		if err != nil {
			t.Error(err)
		}
		if current == nil {
			t.Fatal("Block not found")
		}
		fetchedCount++

		byHash, err := dbo.FetchECBlockByHash(hash)

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

	all, err := dbo.FetchAllECBlocks()
	if err != nil {
		t.Error(err)
	}
	if len(all) != max {
		t.Error("Wrong number of entries fetched - %v vs %v", len(all), max)
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
